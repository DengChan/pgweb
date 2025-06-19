package client

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sosedoff/pgweb/pkg/logger"
	"github.com/sosedoff/pgweb/pkg/statements"
	"github.com/sosedoff/pgweb/pkg/structs/meta"
)

type DBMetaQueryer interface {
	TableBasicInfo(db *sqlx.DB, oid int) (meta.TableBasicInfo, error)
}

func NewDBMetaQueryer(dataSourceType, version string) DBMetaQueryer {
	switch dataSourceType {
	case statements.DataSourcePostgreSQL:
		return NewPGDefaultMetaQueryer(version)
	case statements.DataSourcePanweiAP:
		return NewPanweiAPMetaQueryer(version)
	default:
		return NewPGDefaultMetaQueryer("default")
	}
}

type PGDefaultMetaQueryer struct {
	sqlProvider *statements.SQLProvider
}

func NewPGDefaultMetaQueryer(version string) *PGDefaultMetaQueryer {
	return &PGDefaultMetaQueryer{
		sqlProvider: statements.NewSQLProvider(statements.DataSourcePostgreSQL, version),
	}
}

func (q *PGDefaultMetaQueryer) TableBasicInfo(db *sqlx.DB, oid int) (meta.TableBasicInfo, error) {
	var basicInfo meta.TableBasicInfo

	// 获取表基本信息（包括schema）
	basicInfoSQL := q.sqlProvider.TableBasicInfo()
	row := db.QueryRow(basicInfoSQL, oid)
	var t, am sql.NullString
	err := row.Scan(
		&basicInfo.Name,
		&basicInfo.Comment,
		&t,
		&basicInfo.TableType,
		&basicInfo.TableSpace,
		&basicInfo.Owner,
		&basicInfo.Schema,
		&am,
	)
	if err != nil {
		return basicInfo, fmt.Errorf("failed to get table basic info: %w", err)
	}

	// 获取分区键信息，使用查询到的schema
	partitionKeysSQL := q.sqlProvider.TablePartitionKeys()
	var partitionKeys sql.NullString
	row = db.QueryRow(partitionKeysSQL, oid)
	err = row.Scan(&partitionKeys)
	if err != nil {
		logger.Error("pg query table partition keys error: ", err.Error())
		// 分区键可能为空，这是正常的
		basicInfo.PartitionKeys = ""
	} else {
		basicInfo.PartitionKeys = strings.TrimSpace(partitionKeys.String)
	}

	// DistributeKeys 留空，因为用户指定不需要获取
	basicInfo.DistributeKeys = ""

	return basicInfo, nil
}

type PanweiAPMetaQueryer struct {
	sqlProvider *statements.SQLProvider
}

func NewPanweiAPMetaQueryer(version string) *PanweiAPMetaQueryer {
	return &PanweiAPMetaQueryer{
		sqlProvider: statements.NewSQLProvider(statements.DataSourcePanweiAP, version),
	}
}

func (q *PanweiAPMetaQueryer) TableBasicInfo(db *sqlx.DB, oid int) (meta.TableBasicInfo, error) {
	var basicInfo meta.TableBasicInfo

	// 获取表基本信息（只获取基础信息，分区键和分布键留空）
	basicInfoSQL := q.sqlProvider.TableBasicInfo()
	row := db.QueryRow(basicInfoSQL, oid)
	var t, am, fmttype sql.NullString
	err := row.Scan(
		&basicInfo.Name,
		&basicInfo.Comment,
		&basicInfo.Table,
		&t,
		&basicInfo.TableSpace,
		&basicInfo.Owner,
		&basicInfo.Schema, // 接收schema信息但不使用
		&am,
		&fmttype,
	)
	if err != nil {
		return basicInfo, fmt.Errorf("failed to get table basic info: %w", err)
	}

	// 表类型
	// 判断 是 horc还是 hudi orc 还是 hudi parquet
	if am.String == "hudi" {
		basicInfo.TableType = fmt.Sprintf("%s %s", am.String, fmttype.String)
	} else {
		basicInfo.TableType = am.String
	}

	// 分区键和分布键
	sqlMap := q.sqlProvider.SqlMap()
	distributeKeySql, exist := sqlMap["table_distribute_keys"]
	if !exist || distributeKeySql == "" {
		// 如果当前版本没有table_distribute_keys，则使用默认版本
		distributeKeySql = q.sqlProvider.DefaultVersionSqlMap()["table_distribute_keys"]
	}

	if distributeKeySql == "" {
		return basicInfo, fmt.Errorf("failed to get table distribute keys sql")
	}

	row = db.QueryRow(distributeKeySql, oid)
	err = row.Scan(&basicInfo.DistributeKeys)
	if err != nil {
		return basicInfo, fmt.Errorf("failed to get table distribute keys: %w", err)
	}

	selectSql := "select reloptions from pg_class where oid= $1"
	row = db.QueryRow(selectSql, oid)
	var relOptions string
	err = row.Scan(&relOptions)
	if err != nil {
		return basicInfo, fmt.Errorf("failed to get table reloptions: %w", err)
	}

	// 解析分区列
	relOptions = strings.TrimLeft(relOptions, "{")
	relOptions = strings.TrimRight(relOptions, "}")
	if len(strings.TrimSpace(relOptions)) == 0 { // 没有分区信息时返回空
		return basicInfo, nil
	}
	// 将字符串转换为Reader
	reader := strings.NewReader(relOptions)
	// 创建一个csv.Reader
	csvReader := csv.NewReader(reader)
	// 读取CSV数据，返回[]string
	records, err := csvReader.Read()
	if err != nil {
		return basicInfo, fmt.Errorf("failed to parse reloptions: %w", err)
	}
	keys := ""
	// 遍历记录
	for _, record := range records {
		if strings.Contains(record, "partitionpath_field") {
			if fields := strings.Split(record, "="); len(fields) == 2 {
				keys = fields[1]
			}
			break
		}
	}

	basicInfo.PartitionKeys = keys
	return basicInfo, nil
}
