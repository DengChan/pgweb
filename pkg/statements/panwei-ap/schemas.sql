SELECT
  schema_name
FROM
  information_schema.schemata
WHERE
  schema_name NOT IN ('information_schema', 'pg_catalog','pg_aoseg','pg_bitmapindex')
  AND schema_name !~ '^pg_(toast|temp)'
ORDER BY
  schema_name ASC