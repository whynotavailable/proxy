package main

import (
	"encoding/json"
	"github.com/whynotavailable/proxy/proxy"
	"log"
	"net"
	"net/http"
	"strings"
)

func middleware(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}
}

type userIDMessage struct {
	UserID string `json:"userId"`
}

func main() {
	// cidr testing
	target := net.ParseIP("128.0.1.24")
	_, network, err := net.ParseCIDR("128.0.0.34/24")
	if err == nil {
		log.Println(network.Contains(target))
		log.Println(network.IP.String())
	} else {
		log.Println(err.Error())
	}

	// actual main
	http.HandleFunc("/proxy/", proxy.Forwarder(func(req *http.Request) proxy.ForwarderOptions {
		return proxy.ForwarderOptions{
			URL: strings.Replace(req.URL.RequestURI(), "proxy", "api", 1),
		}
	}))

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
