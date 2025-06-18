SELECT
  datname
FROM
  pg_database
WHERE
  datname NOT IN ('hcatalog', 'template0', 'template1')
ORDER BY
  datname ASC
