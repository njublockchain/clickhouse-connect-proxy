package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	proxy "github.com/njublockchain/clickhouse-connect-proxy"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// require CLICKHOUSE_URI
	if os.Getenv("CLICKHOUSE_URI") == "" {
		log.Fatal("Missing CLICKHOUSE_URI")
	}

	enableAuth := len(os.Getenv("ENABLE_AUTH")) > 0 && os.Getenv("ENABLE_AUTH") != "false"
	if enableAuth && os.Getenv("MONGO_URI") == "" {
		log.Fatal("Missing MONGO_URI")
	}

	var whitelist []string
	if os.Getenv("MONGO_WHITELIST") != "" {
		whitelist = strings.Split(os.Getenv("MONGO_WHITELIST"), ",")
	}

	var authPlugin *proxy.AuthPlugin
	if enableAuth {
		authPlugin = proxy.NewAuthPlugin(
			os.Getenv("MONGO_URI"),
			os.Getenv("MONGO_DB"),
			os.Getenv("MONGO_COLL"),
			os.Getenv("MONGO_APITOKEN_KEY"),
			whitelist,
		)
		log.Printf("Auth enabled")
	}

	middleware := proxy.NewProxyMiddleware(os.Getenv("CLICKHOUSE_URI"), authPlugin)

	// create a http/https server to proxy the request
	http.HandleFunc("/", middleware.ProxyRequest)

	listen := os.Getenv("LISTEN")
	log.Printf("Listening on %s", listen)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
