SELECT STRING_AGG(distribution_key, ',' ORDER BY order_num) AS concatenated_names
FROM 
(
SELECT 
    a.attname AS distribution_key,
    distkey_with_ordinality.ordinality as order_num
FROM 
    gp_distribution_policy p
JOIN 
    pg_class c ON p.localoid = c.oid
JOIN 
    pg_namespace n ON c.relnamespace = n.oid
JOIN 
    pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(p.distkey)

JOIN 
    UNNEST(p.distkey) WITH ORDINALITY AS distkey_with_ordinality(distkey, ordinality)
    ON 
    a.attnum = distkey_with_ordinality.distkey
WHERE 
   c.oid=$1
) aa