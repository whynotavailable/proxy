package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Proxy more structured proxy class
type Proxy struct {
	Path string
	// These MUST be in canonical format
	HeaderWhitelist []string
	// just before the request is made
	PreRequest func(inbound, outbound *http.Request) (error, int)
	// To validate the incoming request, before any outbound request building happens
	ValidateRequest func(r *http.Request) (error, int)
	// This returns the actual URI that the proxy will call
	Router func(*http.Request) (string, error)
	// If this is set, just use the provided host
	TargetHost string
	// Will find/replace what's in the Path (only when using TargetHost)
	ReplacementPath string
	// Use the old X-Forwarded-For header instead of Forwarded
	UseXForwardedFor bool
}

// BuildProxy build the function for the proxy
func (p *Proxy) BuildProxy() (func(w http.ResponseWriter, r *http.Request), error) {
	err := p.validateProxy()

	if err != nil {
		// Consumer configured proxy wrong, die to prevent unexpected errors.
		return nil, err
	}

	// load test this on the stack/heap
	client := &http.Client{}

	// pulled these from https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
	hopByHopHeaders := map[string]bool{
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

	if p.TargetHost != "" {
		p.Router = func(r *http.Request) (string, error) {
			uri := r.URL.RequestURI()
			if p.ReplacementPath != "" {
				uri = strings.Replace(uri, p.Path, p.ReplacementPath, 1)
			}

			return p.TargetHost + uri, nil
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if p.ValidateRequest != nil {
			err, code := p.ValidateRequest(r)

			if err != nil {
				http.Error(w, err.Error(), code)
			}
		}

		uri, err := p.Router(r)

		if err != nil {
			// if error on router, it's a bad request
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req, _ := http.NewRequest(r.Method, uri, r.Body)

		if p.HeaderWhitelist != nil {
			// use the whitelist
			for key, parts := range r.Header {
				if whitelist[key] && !hopByHopHeaders[key] {
					for _, val := range parts {
						req.Header.Add(key, val)
					}
				}
			}
		} else {
			// just use the blacklist
			for key, parts := range r.Header {
				if !hopByHopHeaders[key] {
					for _, val := range parts {
						req.Header.Add(key, val)
					}
				}
			}
		}

		clientIP := r.RemoteAddr

		if strings.ContainsRune(clientIP, ':') {
			clientIP = "\"" + clientIP + "\""
		}

		if p.UseXForwardedFor {
			req.Header.Add("X-Forwarded-For", clientIP)
		} else {
			req.Header.Add("Forwarded", fmt.Sprintf("For=%s", clientIP))
		}

		if p.PreRequest != nil {
			err, status := p.PreRequest(r, req)
			if err != nil {
				// error in PreRequest means we don't move forward, persist the status from consumer
				http.Error(w, err.Error(), status)
				return
			}
		}

		resp, err := client.Do(req)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		for key, parts := range resp.Header {
			if !hopByHopHeaders[key] {
				for _, val := range parts {
					w.Header().Add(key, val)
				}
			}
		}

		w.WriteHeader(resp.StatusCode)

		io.Copy(w, resp.Body)
	}, nil
}

// Register builds and registers proxy
func (p *Proxy) Register() {
	pr, err := p.BuildProxy()

	if err != nil {
		log.Println("validation failed")
		return
	}

	http.HandleFunc(p.Path, pr)
}

// Apply settings from one proxy to another
func (p *Proxy) Apply(from Proxy) {
	if from.HeaderWhitelist != nil {
		p.HeaderWhitelist = from.HeaderWhitelist
	}

	if from.PreRequest != nil {
		p.PreRequest = from.PreRequest
	}

	if from.ValidateRequest != nil {
		p.ValidateRequest = from.ValidateRequest
	}

	p.UseXForwardedFor = from.UseXForwardedFor
}

func (p *Proxy) validateProxy() error {
	if p.Router == nil && p.TargetHost == "" {
		return errors.New("Router or TargetHost not set")
	}

	if p.Path == "" {
		return errors.New("Path not set")
	}

	return nil
}
