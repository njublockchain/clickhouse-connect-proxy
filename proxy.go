package proxy

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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

	authPlugin *AuthPlugin
}

func NewProxyMiddleware(clickhouseURI string, authPlugin *AuthPlugin) *ProxyMiddleware {
	u, err := url.Parse(clickhouseURI)
	if err != nil {
		log.Fatal()
	}

	log.Printf("Proxying to %s", u.Host)

	return &ProxyMiddleware{
		clickhouseURI: u,
		authPlugin:    authPlugin,
	}
}

// proxy the http request to the real host
func (pm *ProxyMiddleware) ProxyRequest(w http.ResponseWriter, r *http.Request) {
	// get the api token from basic auth
	apiToken, _, _ := r.BasicAuth()

	// check auth
	if pm.authPlugin != nil {
		if !pm.authPlugin.Auth(apiToken) {
			log.Printf("Unauthorized apiToken: %s", apiToken)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	}

	// override the url query
	urlQuery := r.URL.Query()
	urlQuery.Set("user", pm.clickhouseURI.User.Username())
	password, isSet := pm.clickhouseURI.User.Password()
	if isSet {
		urlQuery.Set("password", password)
	}
	urlQuery.Set("quota_key", apiToken)
	r.URL.RawQuery = urlQuery.Encode()

	// clear basic auth
	r.Header.Del("Authorization")

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
