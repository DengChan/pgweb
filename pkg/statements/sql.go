package statements

// 数据源类型常量
const (
	DataSourcePostgreSQL = "postgres"
	DataSourcePanweiAP   = "panweidb-ap"
	DataSourcePanweiTP   = "panweidb-tp"
)

// Activity queries for specific PG versions (保持现有逻辑)
var Activity = map[string]string{
	"default": "SELECT * FROM pg_stat_activity WHERE datname = current_database()",
	"9.1":     "SELECT datname, current_query, waiting, query_start, procpid as pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
	"9.2":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
	"9.3":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
	"9.4":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
	"9.5":     "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
	"9.6":     "SELECT datname, query, state, wait_event, wait_event_type, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()",
}

// 全局数据源SQL映射
var dataSourceSQLMap = map[string]map[string]map[string]string{
	DataSourcePostgreSQL: postgresqlSQLMap,
	// 预留给panweidb-ap和panweidb-tp的映射，暂时为空
	DataSourcePanweiAP: panweidbAPSQLMap,
	DataSourcePanweiTP: {},
}

// SQLProvider 为指定的数据源类型和版本提供SQL语句
type SQLProvider struct {
	DataSourceType string
	Version        string
}

// NewSQLProvider 创建一个新的SQL提供器
func NewSQLProvider(dataSourceType, version string) *SQLProvider {
	return &SQLProvider{
		DataSourceType: dataSourceType,
		Version:        version,
	}
}

// getSQL 根据查询类型获取对应的SQL语句
func (p *SQLProvider) getSQL(queryType string) string {
	// 获取数据源的SQL映射
	dataSourceMap, exists := dataSourceSQLMap[p.DataSourceType]
	if !exists {
		// 如果数据源不存在，回退到PostgreSQL
		dataSourceMap = dataSourceSQLMap[DataSourcePostgreSQL]
	}

	// 获取版本特定的SQL映射
	versionMap, exists := dataSourceMap[p.Version]
	if !exists {
		// 如果版本不存在，回退到默认版本
		versionMap = dataSourceMap["default"]
	}

	// 获取该版本查询类型对应的SQL 不存在则回退到默认版本
	sql, exists := versionMap[queryType]
	if !exists {
		versionMap = dataSourceMap["default"]
		sql = versionMap[queryType]
	}

	return sql
}

// Databases 获取数据库列表的SQL
func (p *SQLProvider) Databases() string {
	return p.getSQL("databases")
}

// Schemas 获取模式列表的SQL
func (p *SQLProvider) Schemas() string {
	return p.getSQL("schemas")
}

// Info 获取服务器信息的SQL
func (p *SQLProvider) Info() string {
	return p.getSQL("info")
}

// InfoSimple 获取简单服务器信息的SQL
func (p *SQLProvider) InfoSimple() string {
	return p.getSQL("info_simple")
}

// EstimatedTableRowCount 获取表行数估计的SQL
func (p *SQLProvider) EstimatedTableRowCount() string {
	return p.getSQL("estimated_row_count")
}

// TableIndexes 获取表索引信息的SQL
func (p *SQLProvider) TableIndexes() string {
	return p.getSQL("table_indexes")
}

// TableConstraints 获取表约束信息的SQL
func (p *SQLProvider) TableConstraints() string {
	return p.getSQL("table_constraints")
}

// TableStatInfo 获取表信息的SQL
func (p *SQLProvider) TableStatInfo() string {
	return p.getSQL("table_stat_info")
}

// TableInfoCockroach 获取CockroachDB表信息的SQL
func (p *SQLProvider) TableInfoCockroach() string {
	return p.getSQL("table_info_cockroach")
}

// TableSchema 获取表结构信息的SQL
func (p *SQLProvider) TableSchema() string {
	return p.getSQL("table_schema")
}

// MaterializedView 获取物化视图信息的SQL
func (p *SQLProvider) MaterializedView() string {
	return p.getSQL("materialized_view")
}

// Objects 获取数据库对象列表的SQL
func (p *SQLProvider) Objects() string {
	return p.getSQL("objects")
}

// TablesStats 获取表统计信息的SQL
func (p *SQLProvider) TablesStats() string {
	return p.getSQL("tables_stats")
}

// TablesPartitionKeys 获取表的分区键
func (p *SQLProvider) TablePartitionKeys() string {
	return p.getSQL("table_partition_keys")
}

// TableBasicInfo 获取表基本信息的SQL
func (p *SQLProvider) TableBasicInfo() string {
	return p.getSQL("table_basic_info")
}

// Function 获取函数信息的SQL
func (p *SQLProvider) Function() string {
	return p.getSQL("function")
}

// Settings 获取设置信息的SQL
func (p *SQLProvider) Settings() string {
	return p.getSQL("settings")
}

// Settings 获取Activity查询SQL（保持向后兼容）
func (p *SQLProvider) Activity() string {
	return p.getSQL("activity")
}

func (p *SQLProvider) SqlMap() map[string]string {
	// 获取数据源的SQL映射
	dataSourceMap, exists := dataSourceSQLMap[p.DataSourceType]
	if !exists {
		// 如果数据源不存在，回退到PostgreSQL
		dataSourceMap = dataSourceSQLMap[DataSourcePostgreSQL]
	}

	// 获取版本特定的SQL映射
	versionMap, exists := dataSourceMap[p.Version]
	if !exists {
		// 如果版本不存在，回退到默认版本
		versionMap = dataSourceMap["default"]
	}
	return versionMap
}

func (p *SQLProvider) DefaultVersionSqlMap() map[string]string {
	// 获取数据源的SQL映射
	dataSourceMap, exists := dataSourceSQLMap[p.DataSourceType]
	if !exists {
		// 如果数据源不存在，回退到PostgreSQL
		dataSourceMap = dataSourceSQLMap[DataSourcePostgreSQL]
	}
	return dataSourceMap["default"]
}

// 向后兼容的全局变量和函数，保持现有代码不受影响
var (
	// 保持向后兼容的全局变量
	Databases              = postgresqlDatabasesDefault
	Schemas                = postgresqlSchemasDefault
	Info                   = postgresqlInfoDefault
	InfoSimple             = postgresqlInfoSimpleDefault
	EstimatedTableRowCount = postgresqlEstimatedTableRowCountDefault
	TableIndexes           = postgresqlTableIndexesDefault
	TableConstraints       = postgresqlTableConstraintsDefault
	TableInfo              = postgresqlTableInfoDefault
	TableInfoCockroach     = postgresqlTableInfoCockroachDefault
	TableSchema            = postgresqlTableSchemaDefault
	MaterializedView       = postgresqlMaterializedViewDefault
	Objects                = postgresqlObjectsDefault
	TablesStats            = postgresqlTablesStatsDefault
	Function               = postgresqlFunctionDefault
	Settings               = postgresqlSettingsDefault
)
