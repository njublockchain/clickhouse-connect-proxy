# Set User Quota

Please remember to replace `[xxx]` to yours

```sql
CREATE USER IF NOT EXISTS
    [USERNAME]
IDENTIFIED BY [PASSWORD]
DEFAULT DATABASE NONE;

GRANT 
    SELECT ON ethereum.*, 
    SELECT ON tron.*,
    SELECT ON arbitrumNova.*,
    SELECT ON arbitrumOne.* -- list the readable databases
TO [USERNAME];

CREATE QUOTA IF NOT EXISTS
    [QUOTANAME]
KEYED BY 
    client_key
FOR INTERVAL 1 minute
MAX 
    queries=3,
    result_bytes=1_000_000,
    execution_time=10000 -- modify the settings as your need
TO [QUOTANAME]
```

