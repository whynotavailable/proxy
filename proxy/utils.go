package proxy

import "net/http"

func isHeaderBad(header string) bool {
	badHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, val := range badHeaders {
		if val == header {
			return true
		}
	}
	return false
}

// CopyHeaders copy headers from one to another
func CopyHeaders(src, dest http.Header) {
	for key, parts := range src {
		if !isHeaderBad(key) {
			for _, val := range parts {
				dest.Add(key, val)
			}
		}
	}
}

func copyWhitelist(src, dest http.Header) {

}
