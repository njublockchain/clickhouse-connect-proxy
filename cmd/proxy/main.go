package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	proxy "github.com/njublockchain/clickhouse-connect-proxy"
	"github.com/njublockchain/clickhouse-connect-proxy/auth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// require CLICKHOUSE_URI
	if os.Getenv("CLICKHOUSE_URI") == "" {
		log.Fatal("Missing CLICKHOUSE_URI")
	}

	var whitelist []string
	if os.Getenv("MONGO_WHITELIST") != "" {
		whitelist = strings.Split(os.Getenv("MONGO_WHITELIST"), ",")
	}

	var authPlugin auth.AuthPlugin

	switch os.Getenv("ENABLE_AUTH") {
	case "mongo":
		authPlugin = auth.NewMongoAuthPlugin(
			os.Getenv("MONGO_URI"),
			os.Getenv("MONGO_DB"),
			os.Getenv("MONGO_COLL"),
			os.Getenv("MONGO_APITOKEN_KEY"),
			whitelist,
		)
		log.Printf("Mongo Auth enabled")
	case "pg":
		authPlugin = auth.NewPGAuthPlugin(
			os.Getenv("PG_URI"),
			os.Getenv("PG_QUERY"),
			whitelist,
		)
		log.Printf("Postgres Auth enabled")
	}
	adminKey := os.Getenv("ADMIN_KEY")
	chURI := os.Getenv("CLICKHOUSE_URI")
	middleware := proxy.NewProxyMiddleware(chURI, adminKey, authPlugin)

	// create a http/https server to proxy the request
	http.HandleFunc("/", middleware.ProxyRequest)

	listen := os.Getenv("LISTEN")
	log.Printf("Listening on %s", listen)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
