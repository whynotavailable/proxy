package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/whynotavailable/proxy"
)

const port uint = 8080

// Super simple example of a forwarder proxy with the basicist of settings

func main() {
	apiProxy := proxy.Proxy{
		Path:       "/api/",
		TargetHost: "http://localhost:5000",
	}
	apiProxy.Register()

	log.Println(fmt.Sprintf("Starting example2 on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
