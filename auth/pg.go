package auth

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type PGAuthPlugin struct {
	db        *sql.DB
	query     string
	whitelist []string
}

func NewPGAuthPlugin(pgURI, query string, whitelist []string) *PGAuthPlugin {
	// connStr := "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	db, err := sql.Open("postgres", pgURI)
	if err != nil {
		panic(err)
	}

	return &PGAuthPlugin{
		db:        db,
		query:     query,
		whitelist: whitelist,
	}
}

func (ap *PGAuthPlugin) Auth(apiToken string) bool {
	if ap.whitelist != nil {
		for _, token := range ap.whitelist {
			if token == apiToken {
				return true
			}
		}
	}

	log.Println(ap.query)
	rows, err := ap.db.Query(ap.query, apiToken)
	if err != nil {
		panic(err)
	}

	if rows.Next() {
		return true
	}

	return false
}
