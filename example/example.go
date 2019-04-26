package main

import (
	"encoding/json"
	"github.com/whynotavailable/proxy"
	"log"
	"net"
	"net/http"
	"strings"
)

var targetRange *net.IPNet

func getClientIP(r *http.Request) net.IP {
	return net.ParseIP(r.Header.Get("TrueClient-Ip"))
}

func middleware(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !targetRange.Contains(ip) {
			http.Error(w, "Maintenence", http.StatusServiceUnavailable)
			return
		}

		handler(w, r)
	}
}

type userIDMessage struct {
	UserID string `json:"userId"`
}

func mainForwarder(req *http.Request) proxy.ForwarderOptions {
	return proxy.ForwarderOptions{
		URL: "http://example.com" + strings.Replace(req.URL.RequestURI(), "proxy", "api", 1),
	}
}

func main() {
	// cidr testing
	_, network, err := net.ParseCIDR("128.0.0.34/24")
	if err == nil {
		targetRange = network
	}

	// actual main
	http.HandleFunc("/proxy/", middleware(proxy.Forwarder(mainForwarder)))

	http.HandleFunc("/userid", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(userIDMessage{
			"john",
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})

	http.HandleFunc("/", middleware(proxy.StaticFileHoster("./static", "/example.txt")))
	log.Println("Starting Server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
