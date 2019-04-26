package proxy

import (
	"fmt"
	"io"
	"net/http"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	client := http.Client{}

	req, _ := http.NewRequest("GET", "http://example.com", r.Body)

	resp, err := client.Do(req)

	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	CopyHeaders(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

// ForwarderOptions asd
type ForwarderOptions struct {
	URL                string
	WhitelistedHeaders []string
	ExtraHeaders       []string
}

// Forwarder method to forward API requests to target location.
func Forwarder(urlGetter func(*http.Request) ForwarderOptions) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := urlGetter(r)
		fmt.Fprintf(w, "Hi there, I love %s!", opts.URL)
	}
}
