package statements

import (
	_ "embed"
)

var (
	// PostgreSQL默认SQL文件嵌入
	//go:embed sql/databases.sql
	postgresqlDatabasesDefault string

	//go:embed sql/schemas.sql
	postgresqlSchemasDefault string

	//go:embed sql/info.sql
	postgresqlInfoDefault string

	//go:embed sql/info_simple.sql
	postgresqlInfoSimpleDefault string

	//go:embed sql/estimated_row_count.sql
	postgresqlEstimatedTableRowCountDefault string

	//go:embed sql/table_indexes.sql
	postgresqlTableIndexesDefault string

	//go:embed sql/table_constraints.sql
	postgresqlTableConstraintsDefault string

	//go:embed sql/table_stat_info.sql
	postgresqlTableInfoDefault string

	//go:embed sql/table_info_cockroach.sql
	postgresqlTableInfoCockroachDefault string

	//go:embed sql/table_schema.sql
	postgresqlTableSchemaDefault string

	//go:embed sql/table_partition_keys.sql
	postgresqlTablePartitionKeysDefault string

	//go:embed sql/table_basic_info.sql
	postgresqlTableBasicInfoDefault string

	//go:embed sql/materialized_view.sql
	postgresqlMaterializedViewDefault string

	//go:embed sql/objects.sql
	postgresqlObjectsDefault string

	//go:embed sql/tables_stats.sql
	postgresqlTablesStatsDefault string

	//go:embed sql/function.sql
	postgresqlFunctionDefault string

	//go:embed sql/settings.sql
	postgresqlSettingsDefault string

	postgresqlActivityDefault = "SELECT * FROM pg_stat_activity WHERE datname = current_database()"

	postgresqlActivity91 = "SELECT datname, current_query, waiting, query_start, procpid as pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()"

	postgresqlActivity92_95 = "SELECT datname, query, state, waiting, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()"

	postgresqlActivity96 = "SELECT datname, query, state, wait_event, wait_event_type, query_start, state_change, pid, datid, application_name, client_addr FROM pg_stat_activity WHERE datname = current_database()"

	// PostgreSQL版本化SQL映射
	postgresqlSQLMap = map[string]map[string]string{
		"default": {
			"databases":            postgresqlDatabasesDefault,
			"schemas":              postgresqlSchemasDefault,
			"info":                 postgresqlInfoDefault,
			"info_simple":          postgresqlInfoSimpleDefault,
			"estimated_row_count":  postgresqlEstimatedTableRowCountDefault,
			"table_indexes":        postgresqlTableIndexesDefault,
			"table_constraints":    postgresqlTableConstraintsDefault,
			"table_stat_info":      postgresqlTableInfoDefault,
			"table_info_cockroach": postgresqlTableInfoCockroachDefault,
			"table_schema":         postgresqlTableSchemaDefault,
			"materialized_view":    postgresqlMaterializedViewDefault,
			"objects":              postgresqlObjectsDefault,
			"tables_stats":         postgresqlTablesStatsDefault,
			"function":             postgresqlFunctionDefault,
			"settings":             postgresqlSettingsDefault,
			"activity":             postgresqlActivityDefault,
			"table_partition_keys": postgresqlTablePartitionKeysDefault,
			"table_basic_info":     postgresqlTableBasicInfoDefault,
		},
		"9.1": {
			"activity": postgresqlActivity91,
		},
		"9.2": {
			"activity": postgresqlActivity92_95,
		},
		"9.3": {
			"activity": postgresqlActivity92_95,
		},
		"9.4": {
			"activity": postgresqlActivity92_95,
		},
		"9.5": {
			"activity": postgresqlActivity92_95,
		},
		"9.6": {
			"activity": postgresqlActivity96,
		},
	}
)
