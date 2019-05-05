package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/whynotavailable/proxy"
)

const port uint = 8080

func main() {
	f, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Fatal(err.Error())
	}

	var proxies []proxy.Proxy

	baseProxy := &proxy.Proxy{
		PreRequest: func(inbound, outbound *http.Request) (error, int) {
			client := inbound.Header.Get("client")
			if client == "" {
				return errors.New("not available"), 503
			}
			return nil, 0
		},
	}

	json.Unmarshal(f, &proxies)

	log.Println(proxies)

	for i := range proxies {
		proxy := proxies[i]
		proxy.Apply(baseProxy)
		log.Println("Registering proxy at " + proxy.Path)
		proxy.Register()
	}

	log.Println(fmt.Sprintf("Starting gateway on port %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
