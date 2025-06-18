package statements

import (
	_ "embed"
)

var (
	// PostgreSQL默认SQL文件嵌入
	//go:embed panwei-ap/databases.sql
	panweidbAPDatabasesDefault string

	//go:embed panwei-ap/schemas.sql
	panweidbAPSchemasDefault string

	//go:embed panwei-ap/info.sql
	panweidbAPInfoDefault string

	//go:embed panwei-ap/info_simple.sql
	panweidbAPInfoSimpleDefault string

	//go:embed panwei-ap/estimated_row_count.sql
	panweidbAPEstimatedTableRowCountDefault string

	//go:embed panwei-ap/table_indexes.sql
	panweidbAPTableIndexesDefault string

	//go:embed panwei-ap/table_constraints.sql
	panweidbAPTableConstraintsDefault string

	//go:embed panwei-ap/table_info.sql
	panweidbAPTableInfoDefault string

	//go:embed panwei-ap/table_info_cockroach.sql
	panweidbAPTableInfoCockroachDefault string

	//go:embed panwei-ap/table_schema.sql
	panweidbAPTableSchemaDefault string

	//go:embed panwei-ap/materialized_view.sql
	panweidbAPMaterializedViewDefault string

	//go:embed panwei-ap/objects.sql
	panweidbAPObjectsDefault string

	//go:embed panwei-ap/tables_stats.sql
	panweidbAPTablesStatsDefault string

	//go:embed panwei-ap/function.sql
	panweidbAPFunctionDefault string

	//go:embed panwei-ap/settings.sql
	panweidbAPSettingsDefault string

	//go:embed panwei-ap/activity.sql
	panweidbAPActivityDefault string

	// PostgreSQL版本化SQL映射
	panweidbAPSQLMap = map[string]map[string]string{
		"default": {
			"databases":            panweidbAPDatabasesDefault,
			"schemas":              panweidbAPSchemasDefault,
			"info":                 panweidbAPInfoDefault,
			"info_simple":          panweidbAPInfoSimpleDefault,
			"estimated_row_count":  panweidbAPEstimatedTableRowCountDefault,
			"table_indexes":        panweidbAPTableIndexesDefault,
			"table_constraints":    panweidbAPTableConstraintsDefault,
			"table_info":           panweidbAPTableInfoDefault,
			"table_info_cockroach": panweidbAPTableInfoCockroachDefault,
			"table_schema":         panweidbAPTableSchemaDefault,
			"materialized_view":    panweidbAPMaterializedViewDefault,
			"objects":              panweidbAPObjectsDefault,
			"tables_stats":         panweidbAPTablesStatsDefault,
			"function":             panweidbAPFunctionDefault,
			"settings":             panweidbAPSettingsDefault,
			"activity":             panweidbAPActivityDefault,
		},
	}
)
