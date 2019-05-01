package proxy

import (
	"io"
	"net/http"
)

// ForwarderOptions asd
type ForwarderOptions struct {
	URL          string
	ExtraHeaders map[string]string
}

// Forwarder method to forward API requests to target location.
func Forwarder(urlGetter func(*http.Request) ForwarderOptions, whitelist []string) func(http.ResponseWriter, *http.Request) {
	wl := make(map[string]bool)

	for _, val := range whitelist {
		wl[val] = true
	}

	badHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"Te":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		opts := urlGetter(r)

		client := http.Client{}

		req, _ := http.NewRequest(r.Method, opts.URL, r.Body)

		for key := range r.Header {
			if wl[key] {
				req.Header.Set(key, r.Header.Get(key))
			}
		}

		for key, val := range opts.ExtraHeaders {
			req.Header.Set(key, val)
		}

		resp, err := client.Do(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for key := range resp.Header {
			if !badHeaders[key] {
				w.Header().Set(key, resp.Header.Get(key))
			}
		}

		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	}
}
