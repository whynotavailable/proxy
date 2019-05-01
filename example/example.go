package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/whynotavailable/proxy"
)

func getClientIP(r *http.Request) net.IP {
	return net.ParseIP(r.Header.Get("TrueClient-Ip"))
}

func middleware(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	_, target, err := net.ParseCIDR("128.0.0.1/24")

	if err != nil {
		log.Fatalln(err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !target.Contains(ip) {
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
		URL: "http://localhost:5000" + strings.Replace(req.URL.RequestURI(), "proxy", "api", 1),
	}
}

func userGetter(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(userIDMessage{
		"john",
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	// actual main
	whiteList := []string{
		"Content-Type",
	}

	http.HandleFunc("/proxy/", proxy.Forwarder(mainForwarder, whiteList))

	http.HandleFunc("/userid", middleware(userGetter))

	http.HandleFunc("/", proxy.StaticFileHoster("./static", "/example.txt"))

	log.Println("Starting Server")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
