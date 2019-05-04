package proxy

import (
	"errors"
	"io"
	"log"
	"net/http"
)

// Proxy more structured proxy class
type Proxy struct {
	Path            string
	HeaderWhitelist []string
	PreRequest      func(inbound, outbound *http.Request) (error, int)
	// This returns the actual URI that the proxy will call
	Router func(*http.Request) (string, error)
}

// Register the proxy with the http pipeline
func (p *Proxy) Register() {
	err := p.validateProxy()

	if err != nil {
		log.Fatal(err)
	}

	// load test this on the stack/heap
	client := http.Client{}

	// pulled these from https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
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

	whitelist := make(map[string]bool)

	for _, header := range p.HeaderWhitelist {
		whitelist[header] = true
	}

	proxy := func(w http.ResponseWriter, r *http.Request) {

		uri, err := p.Router(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req, _ := http.NewRequest(r.Method, uri, r.Body)

		if p.HeaderWhitelist != nil {
			// use the whitelist
			for key := range r.Header {
				if whitelist[key] && !badHeaders[key] {
					req.Header.Set(key, r.Header.Get(key))
				}
			}
		} else {
			// just use the blacklist
			for key := range r.Header {
				if !badHeaders[key] {
					req.Header.Set(key, r.Header.Get(key))
				}
			}
		}

		if p.PreRequest != nil {
			err, status := p.PreRequest(r, req)
			if err != nil {
				http.Error(w, err.Error(), status)
			}
		}

		resp, err := client.Do(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		for key := range resp.Header {
			if !badHeaders[key] {
				w.Header().Set(key, resp.Header.Get(key))
			}
		}

		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	}

	http.HandleFunc(p.Path, proxy)
}

func (p *Proxy) validateProxy() error {
	if p.Router == nil {
		return errors.New("Router not set")
	}

	if p.Path == "" {
		return errors.New("Path not set")
	}

	return nil
}
