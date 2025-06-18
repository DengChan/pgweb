package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sosedoff/pgweb/pkg/client"
	"github.com/sosedoff/pgweb/pkg/command"
)

// GetDatabases renders a list of all databases on the server
func GetDatabases(c *gin.Context) {
	if command.Opts.LockSession {
		serveResult(c, []string{}, nil)
		return
	}
	conn := DB(c)
	if conn.External {
		errorResponse(c, 403, errNotPermitted)
		return
	}

	names, err := DB(c).Databases()
	serveResult(c, names, err)
}

// GetObjects renders a list of database objects
func GetObjects(c *gin.Context) {
	result, err := DB(c).Objects()
	if err != nil {
		badRequest(c, err)
		return
	}
	successResponse(c, client.ObjectsFromResult(result))
}

// GetSchemas renders list of available schemas
func GetSchemas(c *gin.Context) {
	res, err := DB(c).Schemas()
	serveResult(c, res, err)
}

// GetTable renders table information
func GetTable(c *gin.Context) {
	var (
		res *client.Result
		err error
	)

	db := DB(c)
	tableName := c.Params.ByName("table")

	switch c.Request.FormValue("type") {
	case client.ObjTypeMaterializedView:
		res, err = db.MaterializedView(tableName)
	case client.ObjTypeFunction:
		res, err = db.Function(tableName)
	default:
		res, err = db.Table(tableName)
	}

	serveResult(c, res, err)
}

// GetTableInfo renders a selected table information
func GetTableInfo(c *gin.Context) {
	res, err := DB(c).TableStatInfo(c.Params.ByName("table"))
	if err == nil {
		successResponse(c, res.Format()[0])
	} else {
		badRequest(c, err)
	}
}

// GetTableBasicInfo renders basic table information by OID
func GetTableBasicInfo(c *gin.Context) {
	oidStr := c.Params.ByName("oid")
	if oidStr == "" {
		badRequest(c, "oid parameter is required")
		return
	}

	oid, err := strconv.Atoi(oidStr)
	if err != nil {
		badRequest(c, "oid must be a valid integer")
		return
	}

	dbClient := DB(c)
	if dbClient == nil {
		badRequest(c, errNotConnected)
		return
	}

	// 调用封装的TableBasicInfo方法
	basicInfo, err := dbClient.TableBasicInfo(oid)
	if err != nil {
		badRequest(c, err)
		return
	}

	successResponse(c, basicInfo)
}

// GetTableIndexes renders a list of database table indexes
func GetTableIndexes(c *gin.Context) {
	res, err := DB(c).TableIndexes(c.Params.ByName("table"))
	serveResult(c, res, err)
}

// GetTableConstraints renders a list of database constraints
func GetTableConstraints(c *gin.Context) {
	res, err := DB(c).TableConstraints(c.Params.ByName("table"))
	serveResult(c, res, err)
}

// GetTablesStats renders data sizes and estimated rows for all tables in the database
func GetTablesStats(c *gin.Context) {
	db := DB(c)

	connCtx, err := db.GetConnContext()
	if err != nil {
		badRequest(c, err)
		return
	}

	res, err := db.TablesStats()
	if err != nil {
		badRequest(c, err)
		return
	}

	format := getQueryParam(c, "format")
	if format == "" {
		format = "json"
	}

	// Save as attachment if exporting parameter is set
	if getQueryParam(c, "export") == "true" {
		ts := time.Now().Format(time.DateOnly)

		filename := fmt.Sprintf("pgweb-dbstats-%s-%s.%s", connCtx.Database, ts, format)
		c.Writer.Header().Set("Content-disposition", "attachment;filename="+filename)
	}

	switch format {
	case "json":
		c.JSON(http.StatusOK, res)
	case "csv":
		c.Data(http.StatusOK, "text/csv", res.CSV())
	case "xml":
		c.XML(200, res)
	default:
		badRequest(c, "invalid format")
	}
}
