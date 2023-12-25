# clickhouse-connect-proxy

This is a simple reverse proxy for Clickhouse database, developed by and deployed on the Kylink Platform.

Status: active updating

Features:
- basic reverse proxying Clickhouse HTTP(S) Service 
- external user authorization (by MongoDB)
- external user quota limit (by native)

## Usage

0. Install the latest golang
1. Create users as [the guidance](./USER_QUOTA.md)
2. `git clone https://github.com/njublockchain/clickhouse-connect-proxy && cd clickhouse-connect-proxy`  
3. `mv .env.example .env`
4. Edit `.env` with your configuation
5. `go run ./cmd/proxy`

## TODO

- optimize for higher performance
- support multi-level external users
- security check
- better Docker support

