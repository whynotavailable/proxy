package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

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

type tokenHandler struct {
	sync.RWMutex
	token          string
	tokenRetriever func() string
}

func (t *tokenHandler) SetToken(token string) {
	t.Lock()
	t.token = token
	t.Unlock()
}

func (t *tokenHandler) GetToken() string {
	var token string
	t.RLock()
	token = t.token
	t.RUnlock()

	if t.token == "" { // condition to get token
		tgetter := func() { // go off and get token
			token := t.tokenRetriever()
			t.SetToken(token)
		}

		go tgetter()
	}

	token = t.token
	return token
}

func tokenGetter() string {
	return "my token"
}

func forwarderWrap() func(*http.Request) proxy.ForwarderOptions {
	t := tokenHandler{
		tokenRetriever: tokenGetter,
	}

	t.SetToken(t.tokenRetriever()) // set initial value

	return func(req *http.Request) proxy.ForwarderOptions {
		return proxy.ForwarderOptions{
			URL: "http://localhost:5000" + strings.Replace(req.URL.RequestURI(), "proxy", "api", 1),
			ExtraHeaders: map[string]string{
				"token": t.GetToken(),
			},
		}
	}
}

func userGetter(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(userIDMessage{
		"john",
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func testerRouter(r *http.Request) (string, error) {
	return "http://localhost:5000" + strings.Replace(r.URL.RequestURI(), "tester", "api", 1), nil
}

func main() {
	mainProxy := proxy.Proxy{
		Path:   "/tester/",
		Router: testerRouter,
	}

	apiProxy := proxy.Proxy{
		Path:       "/api/",
		TargetHost: "http://localhost:5000",
	}

	mainProxy.PreRequest = func(inbound, outbound *http.Request) (error, int) {
		outbound.Header.Set("x-auth", "my value")
		return nil, 0
	}

	mainProxy.Register()
	apiProxy.Register()

	http.HandleFunc("/proxy/", proxy.Forwarder(forwarderWrap(), nil))

	http.HandleFunc("/userid", middleware(userGetter))

	http.HandleFunc("/", proxy.StaticFileHoster("./static", "/example.txt"))

	log.Println("Starting Server")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
