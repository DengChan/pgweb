package meta

type TableBasicInfo struct {
	Name      string `json:"name"`
	Schema    string `json:"schema"`
	Comment   string `json:"comment"`
	Table     string `json:"table"`
	TableType string `json:"table_type"`

	PartitionKeys  string `json:"partition_keys"`
	DistributeKeys string `json:"distribute_keys"`

	TableSpace string `json:"table_space"`

	Owner string `json:"owner"`
}
