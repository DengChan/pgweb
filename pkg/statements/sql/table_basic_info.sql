SELECT 
  c.relname AS name,
  COALESCE(pg_catalog.obj_description(c.oid), '') AS comment,
  c.relname AS table_name,
  CASE c.relkind
    WHEN 'r' THEN 'table'
    WHEN 'v' THEN 'view' 
    WHEN 'm' THEN 'materialized_view'
    WHEN 'S' THEN 'sequence'
    WHEN 'f' THEN 'foreign_table'
    ELSE 'unknown'
  END AS table_type,
  COALESCE(ts.spcname, 'pg_default') AS table_space,
  pg_catalog.pg_get_userbyid(c.relowner) AS owner,
  n.nspname AS schema_name,
  pg_am.amname
FROM 
  pg_catalog.pg_class c
LEFT JOIN 
  pg_catalog.pg_tablespace ts ON ts.oid = c.reltablespace
LEFT JOIN 
  pg_catalog.pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN 
  pg_am on pg_am.oid = c.relam
WHERE 
  c.oid = $1::oid; 