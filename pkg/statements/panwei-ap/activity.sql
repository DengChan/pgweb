SELECT datname, query, state, wait_event, wait_event_type, query_start, state_change, pid, datid, application_name, client_addr 
FROM pg_stat_activity 
WHERE datname = current_database()