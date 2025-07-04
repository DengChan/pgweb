package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuvistavie/securerandom"

	"github.com/sosedoff/pgweb/pkg/bookmarks"
	"github.com/sosedoff/pgweb/pkg/client"
	"github.com/sosedoff/pgweb/pkg/command"
	"github.com/sosedoff/pgweb/pkg/connection"
	"github.com/sosedoff/pgweb/pkg/logger"
	"github.com/sosedoff/pgweb/pkg/metrics"
	"github.com/sosedoff/pgweb/pkg/queries"
	"github.com/sosedoff/pgweb/pkg/shared"
	"github.com/sosedoff/pgweb/static"
)

var (
	// DbClient represents the active database connection in a single-session mode
	DbClient *client.Client

	// DbSessions represents the mapping for client connections
	DbSessions *SessionManager

	// QueryStore reads the SQL queries stores in the home directory
	QueryStore *queries.Store
)

// DB returns a database connection from the client context
func DB(c *gin.Context) *client.Client {
	if command.Opts.Sessions {
		return DbSessions.Get(getSessionId(c.Request))
	}
	return DbClient
}

// setClient sets the database client connection for the sessions
func setClient(c *gin.Context, newClient *client.Client) error {
	currentClient := DB(c)
	if currentClient != nil {
		currentClient.Close()
	}

	if !command.Opts.Sessions {
		DbClient = newClient
		return nil
	}

	sid := getSessionId(c.Request)
	if sid == "" {
		return errSessionRequired
	}

	DbSessions.Add(sid, newClient)
	logger.Debugf("session %s set new client, clent source type: %s", sid, newClient.GetDataSourceType())
	return nil
}

// GetHome renders the home page
func GetHome(prefix string) http.Handler {
	if prefix != "" {
		prefix = "/" + prefix
	}
	return http.StripPrefix(prefix, static.GetHandler())
}

func GetAssets(prefix string) http.Handler {
	if prefix != "" {
		prefix = "/" + prefix + "static/"
	} else {
		prefix = "/static/"
	}
	return http.StripPrefix(prefix, static.GetHandler())
}

// GetSessions renders the number of active sessions
func GetSessions(c *gin.Context) {
	// In debug mode endpoint will return a lot of sensitive information
	// like full database connection string and all query history.
	if command.Opts.Debug {
		successResponse(c, DbSessions.Sessions())
		return
	}
	successResponse(c, gin.H{"sessions": DbSessions.Len()})
}

// ConnectWithBackend creates a new connection based on backend resource
func ConnectWithBackend(c *gin.Context) {
	// Setup a new backend client
	backend := Backend{
		Endpoint:    command.Opts.ConnectBackend,
		Token:       command.Opts.ConnectToken,
		PassHeaders: strings.Split(command.Opts.ConnectHeaders, ","),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Fetch connection credentials
	cred, err := backend.FetchCredential(ctx, c.Param("resource"), c)
	if err != nil {
		badRequest(c, err)
		return
	}

	// Make the new session
	sid, err := securerandom.Uuid()
	if err != nil {
		badRequest(c, err)
		return
	}
	c.Request.Header.Add("x-session-id", sid)

	// Connect to the database
	cl, err := client.NewFromUrl(cred.DatabaseURL, nil)
	if err != nil {
		badRequest(c, err)
		return
	}
	cl.External = true

	// Finalize session seetup
	_, err = cl.Info()
	if err == nil {
		err = setClient(c, cl)
	}
	if err != nil {
		cl.Close()
		badRequest(c, err)
		return
	}

	redirectURI := fmt.Sprintf("/%s?session=%s", command.Opts.Prefix, sid)
	c.Redirect(302, redirectURI)
}

// Connect creates a new client connection
func Connect(c *gin.Context) {
	if command.Opts.LockSession {
		badRequest(c, errSessionLocked)
		return
	}

	var (
		cl         *client.Client
		err        error
		bookmarkID string
	)

	bookmarkID = c.Request.FormValue("bookmark_id")
	if bookmarkID != "" {
		cl, err = ConnectWithBookmark(bookmarkID)
	} else if command.Opts.BookmarksOnly {
		err = errNotPermitted
	} else {
		cl, err = ConnectWithURL(c)
	}
	if err != nil {
		badRequest(c, err)
		return
	}

	err = cl.Test()
	if err != nil {
		badRequest(c, err)
		return
	}

	info, err := cl.Info()
	if err == nil {
		err = setClient(c, cl)
	}
	if err != nil {
		cl.Close()
		badRequest(c, err)
		return
	}

	// Save last connection configuration (only for direct connections, not bookmarks)
	if bookmarkID == "" {
		saveLastConnection(c, cl)
	}

	successResponse(c, info.Format()[0])
}

func ConnectWithURL(c *gin.Context) (*client.Client, error) {
	url := c.Request.FormValue("url")
	if url == "" {
		return nil, errURLRequired
	}

	url, err := connection.FormatURL(command.Options{
		URL:      url,
		Passfile: command.Opts.Passfile,
	})
	if err != nil {
		return nil, err
	}

	var sshInfo *shared.SSHInfo
	if c.Request.FormValue("ssh") != "" {
		sshInfo = parseSshInfo(c)
	}

	cl, err := client.NewFromUrl(url, sshInfo)
	if err != nil {
		return nil, err
	}

	// 设置数据源类型
	dataSourceType := c.Request.FormValue("data_source_type")
	if dataSourceType == "" {
		dataSourceType = "postgres" // 默认为PostgreSQL
	}
	cl.SetDataSourceType(dataSourceType)
	logger.Debug("ConnectWithURL dataSourceType", dataSourceType)

	return cl, nil
}

func ConnectWithBookmark(id string) (*client.Client, error) {
	manager := bookmarks.NewManager(command.Opts.BookmarksDir)

	bookmark, err := manager.Get(id)
	if err != nil {
		return nil, err
	}

	return client.NewFromBookmark(bookmark)
}

// SwitchDb perform database switch for the client connection
func SwitchDb(c *gin.Context) {
	if command.Opts.LockSession {
		badRequest(c, errSessionLocked)
		return
	}

	name := c.Request.URL.Query().Get("db")
	if name == "" {
		name = c.Request.FormValue("db")
	}
	if name == "" {
		badRequest(c, errDatabaseNameRequired)
		return
	}

	conn := DB(c)
	if conn == nil {
		badRequest(c, errNotConnected)
		return
	}
	dataSourceType := conn.GetDataSourceType()

	// Do not allow switching databases for connections from third-party backends
	if conn.External {
		badRequest(c, errSessionLocked)
		return
	}

	currentURL, err := neturl.Parse(conn.ConnectionString)
	if err != nil {
		badRequest(c, errInvalidConnString)
		return
	}
	currentURL.Path = name

	cl, err := client.NewFromUrl(currentURL.String(), nil)
	if err != nil {
		badRequest(c, err)
		return
	}
	cl.SetDataSourceType(dataSourceType)

	err = cl.Test()
	if err != nil {
		badRequest(c, err)
		return
	}

	info, err := cl.Info()
	if err == nil {
		err = setClient(c, cl)
	}
	if err != nil {
		logger.Error("SwitchDb error: ", err.Error())
		cl.Close()
		badRequest(c, err)
		return
	}

	conn.Close()

	successResponse(c, info.Format()[0])
}

// Disconnect closes the current database connection
func Disconnect(c *gin.Context) {
	if command.Opts.LockSession {
		badRequest(c, errSessionLocked)
		return
	}

	if command.Opts.Sessions {
		result := DbSessions.Remove(getSessionId(c.Request))
		successResponse(c, gin.H{"success": result})
		return
	}

	conn := DB(c)
	if conn == nil {
		badRequest(c, errNotConnected)
		return
	}

	err := conn.Close()
	if err != nil {
		badRequest(c, err)
		return
	}

	DbClient = nil
	successResponse(c, gin.H{"success": true})
}

// RunQuery executes the query
func RunQuery(c *gin.Context) {
	query := cleanQuery(c.Request.FormValue("query"))

	if query == "" {
		badRequest(c, errQueryRequired)
		return
	}

	HandleQuery(query, c)
}

// ExplainQuery renders query explain plan
func ExplainQuery(c *gin.Context) {
	query := cleanQuery(c.Request.FormValue("query"))

	if query == "" {
		badRequest(c, errQueryRequired)
		return
	}

	HandleQuery(fmt.Sprintf("EXPLAIN %s", query), c)
}

// AnalyzeQuery renders query explain plan and analyze profile
func AnalyzeQuery(c *gin.Context) {
	query := cleanQuery(c.Request.FormValue("query"))

	if query == "" {
		badRequest(c, errQueryRequired)
		return
	}

	HandleQuery(fmt.Sprintf("EXPLAIN ANALYZE %s", query), c)
}

// GetTableRows renders table rows
func GetTableRows(c *gin.Context) {
	offset, err := parseIntFormValue(c, "offset", 0)
	if err != nil {
		badRequest(c, err)
		return
	}

	limit, err := parseIntFormValue(c, "limit", 100)
	if err != nil {
		badRequest(c, err)
		return
	}

	opts := client.RowsOptions{
		Limit:      limit,
		Offset:     offset,
		SortColumn: c.Request.FormValue("sort_column"),
		SortOrder:  c.Request.FormValue("sort_order"),
		Where:      c.Request.FormValue("where"),
	}

	res, err := DB(c).TableRows(c.Params.ByName("table"), opts)
	if err != nil {
		badRequest(c, err)
		return
	}

	countRes, err := DB(c).TableRowsCount(c.Params.ByName("table"), opts)
	if err != nil {
		badRequest(c, err)
		return
	}

	numFetch := int64(opts.Limit)
	numOffset := int64(opts.Offset)
	numRows := countRes.Rows[0][0].(int64)
	numPages := numRows / numFetch

	if numPages*numFetch < numRows {
		numPages++
	}

	res.Pagination = &client.Pagination{
		Rows:    numRows,
		Page:    (numOffset / numFetch) + 1,
		Pages:   numPages,
		PerPage: numFetch,
	}

	serveResult(c, res, err)
}

// GetHistory renders a list of recent queries
func GetHistory(c *gin.Context) {
	successResponse(c, DB(c).History)
}

// GetConnectionInfo renders information about current connection
func GetConnectionInfo(c *gin.Context) {
	conn := DB(c)

	if err := conn.TestWithTimeout(5 * time.Second); err != nil {
		badRequest(c, err)
		return
	}

	res, err := conn.Info()
	if err != nil {
		badRequest(c, err)
		return
	}

	info := res.Format()[0]
	info["session_lock"] = command.Opts.LockSession

	successResponse(c, info)
}

// GetServerSettings renders a list of all server settings
func GetServerSettings(c *gin.Context) {
	res, err := DB(c).ServerSettings()
	serveResult(c, res, err)
}

// GetActivity renders a list of running queries
func GetActivity(c *gin.Context) {
	res, err := DB(c).Activity()
	serveResult(c, res, err)
}

// HandleQuery runs the database query
func HandleQuery(query string, c *gin.Context) {
	metrics.IncrementQueriesCount()

	rawQuery, err := base64.StdEncoding.DecodeString(desanitize64(query))
	if err == nil {
		query = string(rawQuery)
	}

	result, err := DB(c).Query(query)
	if err != nil {
		badRequest(c, err)
		return
	}

	format := getQueryParam(c, "format")
	filename := getQueryParam(c, "filename")

	if filename == "" {
		filename = fmt.Sprintf("pgweb-%v.%v", time.Now().Unix(), format)
	}

	if format != "" {
		c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	}

	switch format {
	case "csv":
		c.Data(200, "text/csv", result.CSV())
	case "json":
		c.Data(200, "application/json", result.JSON())
	case "xml":
		c.XML(200, result)
	default:
		c.JSON(200, result)
	}
}

// GetBookmarks renders the list of available bookmarks
func GetBookmarks(c *gin.Context) {
	manager := bookmarks.NewManager(command.Opts.BookmarksDir)
	ids, err := manager.ListIDs()
	serveResult(c, ids, err)
}

// GetInfo renders the pgweb system information
func GetInfo(c *gin.Context) {
	successResponse(c, gin.H{
		"app": command.Info,
		"features": gin.H{
			"session_lock":   command.Opts.LockSession,
			"query_timeout":  command.Opts.QueryTimeout,
			"local_queries":  QueryStore != nil,
			"bookmarks_only": command.Opts.BookmarksOnly,
		},
	})
}

// DataExport performs database table export
func DataExport(c *gin.Context) {
	db := DB(c)

	info, err := db.Info()
	if err != nil {
		badRequest(c, err)
		return
	}

	dump := client.Dump{
		Table: strings.TrimSpace(c.Request.FormValue("table")),
	}

	// Perform validation of pg_dump command availability and compatibility.
	// Must be done before the actual command is executed to display errors.
	if err := dump.Validate(db.ServerVersion()); err != nil {
		badRequest(c, err)
		return
	}

	formattedInfo := info.Format()[0]
	filename := formattedInfo["current_database"].(string)
	if dump.Table != "" {
		filename = filename + "_" + dump.Table
	}

	filename = sanitizeFilename(filename)
	filename = fmt.Sprintf("%s_%s", filename, time.Now().Format("20060102_150405"))

	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s.sql.gz"`, filename),
	)

	err = dump.Export(c.Request.Context(), db.ConnectionString, c.Writer)
	if err != nil {
		logger.WithError(err).Error("pg_dump command failed")
		badRequest(c, err)
	}
}

// GetFunction renders function information
func GetFunction(c *gin.Context) {
	res, err := DB(c).Function(c.Param("id"))
	serveResult(c, res, err)
}

func GetLocalQueries(c *gin.Context) {
	connCtx, err := DB(c).GetConnContext()
	if err != nil {
		badRequest(c, err)
		return
	}

	storeQueries, err := QueryStore.ReadAll()
	if err != nil {
		badRequest(c, err)
		return
	}

	queries := []localQuery{}
	for _, q := range storeQueries {
		if !q.IsPermitted(connCtx.Host, connCtx.User, connCtx.Database, connCtx.Mode) {
			continue
		}

		queries = append(queries, localQuery{
			ID:          q.ID,
			Title:       q.Meta.Title,
			Description: q.Meta.Description,
			Query:       cleanQuery(q.Data),
		})
	}

	successResponse(c, queries)
}

func RunLocalQuery(c *gin.Context) {
	query, err := QueryStore.Read(c.Param("id"))
	if err != nil {
		if err == queries.ErrQueryFileNotExist {
			query = nil
		} else {
			badRequest(c, err)
			return
		}
	}
	if query == nil {
		errorResponse(c, 404, "query not found")
		return
	}

	connCtx, err := DB(c).GetConnContext()
	if err != nil {
		badRequest(c, err)
		return
	}

	if !query.IsPermitted(connCtx.Host, connCtx.User, connCtx.Database, connCtx.Mode) {
		errorResponse(c, 404, "query not found")
		return
	}

	if c.Request.Method == http.MethodGet {
		successResponse(c, localQuery{
			ID:          query.ID,
			Title:       query.Meta.Title,
			Description: query.Meta.Description,
			Query:       query.Data,
		})
		return
	}

	statement := cleanQuery(query.Data)
	if statement == "" {
		badRequest(c, errQueryRequired)
		return
	}

	HandleQuery(statement, c)
}

func Ping(c *gin.Context) {
	successResponse(c, "pong")
}

// saveLastConnection saves the last used connection configuration
func saveLastConnection(c *gin.Context, cl *client.Client) {
	// Don't save if bookmarks directory is not configured
	if command.Opts.BookmarksDir == "" {
		return
	}

	connCtx, err := cl.GetConnContext()
	if err != nil {
		// Log error but don't fail the connection
		fmt.Fprintf(os.Stderr, "[WARN] Failed to get connection context for saving last connection: %s\n", err)
		return
	}

	manager := bookmarks.NewLastConnectionManager(command.Opts.BookmarksDir)

	// Parse connection URL to get port
	connURL, err := neturl.Parse(cl.ConnectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Failed to parse connection string: %s\n", err)
		return
	}

	port := 5432 // Default PostgreSQL port
	if connURL.Port() != "" {
		if parsed, err := strconv.Atoi(connURL.Port()); err == nil {
			port = parsed
		}
	}

	// Get SSL mode from connection URL
	sslMode := "disable"
	if query := connURL.Query(); query.Get("sslmode") != "" {
		sslMode = query.Get("sslmode")
	}

	// 获取数据源类型
	dataSourceType := cl.GetDataSourceType()
	if dataSourceType == "" {
		dataSourceType = "postgres"
	}

	lastConn := &bookmarks.LastConnection{
		Host:           connCtx.Host,
		Port:           port,
		User:           connCtx.User,
		Database:       connCtx.Database,
		SSLMode:        sslMode,
		DataSourceType: dataSourceType,
	}

	// If this connection used SSH, save SSH info
	if c.Request.FormValue("ssh") != "" {
		lastConn.SSH = parseSshInfo(c)
	}

	err = manager.Save(lastConn)
	if err != nil {
		// Log error but don't fail the connection
		fmt.Fprintf(os.Stderr, "[WARN] Failed to save last connection: %s\n", err)
	}
}

// GetLastConnection retrieves the last used connection configuration
func GetLastConnection(c *gin.Context) {
	manager := bookmarks.NewLastConnectionManager(command.Opts.BookmarksDir)

	lastConn, err := manager.Load()
	if err != nil {
		badRequest(c, err)
		return
	}

	if lastConn == nil {
		successResponse(c, gin.H{"last_connection": nil})
		return
	}

	successResponse(c, gin.H{"last_connection": lastConn})
}

// SaveLastConnection manually saves a connection configuration (for manual bookmarking)
func SaveLastConnection(c *gin.Context) {
	manager := bookmarks.NewLastConnectionManager(command.Opts.BookmarksDir)

	var lastConn bookmarks.LastConnection
	if err := c.ShouldBindJSON(&lastConn); err != nil {
		badRequest(c, err)
		return
	}

	err := manager.Save(&lastConn)
	if err != nil {
		badRequest(c, err)
		return
	}

	successResponse(c, gin.H{"success": true})
}
