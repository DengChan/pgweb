SELECT
  column_name,
  data_type,
  is_nullable,
  CASE 
    WHEN data_type = 'numeric' THEN numeric_precision::text || ',' || numeric_scale::text
    WHEN data_type IN ('timestamp', 'timestamp with time zone', 'time', 'timestamp without time zone', 'interval') THEN datetime_precision::text
    ELSE character_maximum_length::text
  END AS character_maximum_length,
  character_set_catalog,
  column_default,
  pg_catalog.col_description(('"' || $1::text || '"."' || $2::text || '"')::regclass::oid, ordinal_position) as comment
FROM
  information_schema.columns
WHERE
  table_schema = $1
  AND table_name = $2
