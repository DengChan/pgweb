package api

import (
	"github.com/gin-gonic/gin"

	"github.com/sosedoff/pgweb/pkg/command"
	"github.com/sosedoff/pgweb/pkg/metrics"
)

func SetupMiddlewares(group *gin.RouterGroup) {
	if command.Opts.Cors {
		group.Use(corsMiddleware())
	}

	group.Use(dbCheckMiddleware())
}

func SetupRoutes(router *gin.Engine) {
	root := router.Group(command.Opts.Prefix)

	root.GET("/", gin.WrapH(GetHome(command.Opts.Prefix)))
	root.GET("/static/*path", gin.WrapH(GetAssets(command.Opts.Prefix)))
	root.GET("/connect/:resource", ConnectWithBackend)

	api := root.Group("/api")

	// CORS middleware for all API routes
	if command.Opts.Cors {
		api.Use(corsMiddleware())
	}

	// Routes that don't need database connection
	api.POST("/connect", Connect)
	api.GET("/bookmarks", GetBookmarks)
	api.GET("/last_connection", GetLastConnection)
	api.POST("/last_connection", SaveLastConnection)
	api.GET("/ping", Ping)
	api.GET("/history", GetHistory)
	api.GET("/info", GetInfo)

	if command.Opts.Sessions {
		api.GET("/sessions", GetSessions)
	}

	// Database-related routes with dbCheckMiddleware
	dbApi := api.Group("/db")
	dbApi.Use(dbCheckMiddleware())

	dbApi.POST("/disconnect", Disconnect)
	dbApi.POST("/switchdb", SwitchDb)
	dbApi.GET("/databases", GetDatabases)
	dbApi.GET("/connection", GetConnectionInfo)
	dbApi.GET("/server_settings", GetServerSettings)
	dbApi.GET("/activity", GetActivity)
	dbApi.GET("/schemas", GetSchemas)
	dbApi.GET("/objects", GetObjects)
	dbApi.GET("/tables/:table", GetTable)
	dbApi.GET("/tables/:table/rows", GetTableRows)
	dbApi.GET("/tables/:table/info", GetTableInfo)
	dbApi.GET("/tables/:table/indexes", GetTableIndexes)
	dbApi.GET("/tables/:table/constraints", GetTableConstraints)
	dbApi.GET("/tables_stats", GetTablesStats)
	dbApi.GET("/functions/:id", GetFunction)
	dbApi.GET("/query", RunQuery)
	dbApi.POST("/query", RunQuery)
	dbApi.GET("/explain", ExplainQuery)
	dbApi.POST("/explain", ExplainQuery)
	dbApi.GET("/analyze", AnalyzeQuery)
	dbApi.POST("/analyze", AnalyzeQuery)
	dbApi.GET("/export", DataExport)
	dbApi.GET("/local_queries", requireLocalQueries(), GetLocalQueries)
	dbApi.GET("/local_queries/:id", requireLocalQueries(), RunLocalQuery)
	dbApi.POST("/local_queries/:id", requireLocalQueries(), RunLocalQuery)
}

func SetupMetrics(engine *gin.Engine) {
	if command.Opts.MetricsEnabled && command.Opts.MetricsAddr == "" {
		// NOTE: We're not supporting the MetricsPath CLI option here to avoid the route conflicts.
		engine.GET("/metrics", gin.WrapH(metrics.NewHandler()))
	}
}
