package proxy

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/njublockchain/clickhouse-connect-proxy/auth"
)

func copyHeader(dst, src http.Header) {
	log.Printf("Copying header: %v", src)
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

type ProxyMiddleware struct {
	clickhouseURI *url.URL

	adminKey   string
	authPlugin auth.AuthPlugin
}

func NewProxyMiddleware(clickhouseURI, adminKey string, authPlugin auth.AuthPlugin) *ProxyMiddleware {
	u, err := url.Parse(clickhouseURI)
	if err != nil {
		log.Fatal()
	}

	log.Printf("Proxying to %s", u.Host)

	if len(adminKey) < 8 {
		log.Fatalf("Admin key is too short")
	}

	return &ProxyMiddleware{
		clickhouseURI: u,
		adminKey:      adminKey,
		authPlugin:    authPlugin,
	}
}

// proxy the http request to the real host
func (pm *ProxyMiddleware) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Try to parse POST form values (when applicable) without consuming the body we need to forward.
	var bodyBackup []byte
	parsedForm := false
	if r.Method == http.MethodPost {
		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data") {
			var err error
			bodyBackup, err = io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				http.Error(w, "Bad request.", http.StatusBadRequest)
				return
			}
			// restore body for parsing
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBackup))
			if strings.HasPrefix(ct, "multipart/form-data") {
				if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
					log.Printf("Error parsing multipart form: %v", err)
				} else {
					parsedForm = true
				}
			} else {
				if err := r.ParseForm(); err != nil {
					log.Printf("Error parsing form: %v", err)
				} else {
					parsedForm = true
				}
			}
			// restore body again for forwarding later
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBackup))
		}
	}

	// get the api token from basic auth
	user, _, ok := r.BasicAuth()
	var apiToken string
	if !ok {
		// Try header first
		apiToken = r.Header.Get("X-Clickhouse-User")
		// Fallback to POST form body fields when present
		if apiToken == "" && parsedForm {
			// Standard ClickHouse field name
			apiToken = r.PostFormValue("user")
			if apiToken == "" {
				// Non-standard but sometimes used
				apiToken = r.PostFormValue("X-Clickhouse-User")
			}
		}
		if apiToken == "" {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	} else {
		apiToken = user
	}

	// check admin key
	if apiToken != pm.adminKey {
		// check auth
		if pm.authPlugin != nil {
			if !pm.authPlugin.Auth(apiToken) {
				log.Printf("Unauthorized apiToken: %s", apiToken)
				http.Error(w, "Unauthorized.", http.StatusUnauthorized)
				return
			}
		}
	}

	// override the url query: set ClickHouse auth using backend defaults, but allow API token as quota key
	urlQuery := r.URL.Query()
	// Default user/password from backend DSN
	userFromDSN := pm.clickhouseURI.User.Username()
	passFromDSN, hasPass := pm.clickhouseURI.User.Password()

	// Allow user/password coming from POST form body to be used instead of DSN if provided
	if parsedForm {
		if v := r.PostFormValue("user"); v != "" {
			userFromDSN = v
		}
		if v := r.PostFormValue("password"); v != "" {
			passFromDSN = v
			hasPass = true
		}
	}

	urlQuery.Set("user", userFromDSN)
	if hasPass {
		urlQuery.Set("password", passFromDSN)
	}
	urlQuery.Set("quota_key", apiToken)
	r.URL.RawQuery = urlQuery.Encode()

	// clear basic auth
	r.Header.Del("Authorization")
	// clear clickhouse user/key header
	r.Header.Del("X-Clickhouse-User")
	r.Header.Del("X-Clickhouse-Key")

	// connect to the remote server
	remote, err := net.Dial("tcp", pm.clickhouseURI.Host)
	if err != nil {
		log.Printf("Error dialing remote: %v", err)
		http.Error(w, "Error connecting to remote server.", http.StatusInternalServerError)
		return
	}
	defer remote.Close()

	// set the request host to the real host
	r.Host = pm.clickhouseURI.Host
	// write the request to the remote
	r.Write(remote)

	// read the response from the remote
	resp, err := http.ReadResponse(bufio.NewReader(remote), r)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, "Error reading response.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// copy the response to the client
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// copy the response body to the client
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response to client: %v", err)
		http.Error(w, "Error copying response to client.", http.StatusInternalServerError)
		return
	}
}
