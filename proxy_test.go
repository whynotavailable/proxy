package proxy

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxy_Minimal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	simpleProxy := Proxy{
		Path:       "/api/",
		TargetHost: server.URL,
	}

	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/", nil)

	pFunc(rr, req)

	if rr.Body.String() != `OK` {
		t.Error("No good")
	}
}

func TestProxy_Whitelist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("client") != "my client" {
			t.Error("client header not passed")
		}

		if req.Header.Get("upgrade") != "" {
			t.Error("upgrade not trimmed")
		}

		if req.Header.Get("too") != "" {
			t.Error("too not trimmed")
		}

		rw.Write([]byte(`OK`))
	}))
	defer server.Close()

	simpleProxy := Proxy{
		Path:       "/api/",
		TargetHost: server.URL,
	}

	baseProxy := Proxy{
		HeaderWhitelist: []string{
			"Client",
		},
	}

	simpleProxy.Apply(baseProxy)

	t.Log(simpleProxy.HeaderWhitelist)

	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/", nil)
	req.Header.Set("client", "my client")
	req.Header.Set("upgrade", "1")
	req.Header.Set("too", "Three")

	pFunc(rr, req)
}

func TestProxy_Validation(t *testing.T) {
	simpleProxy := Proxy{}
	_, err := simpleProxy.BuildProxy(nil)

	if err == nil {
		t.Error("validation failure not caught")
	}

	simpleProxy.TargetHost = "whatever"

	_, err = simpleProxy.BuildProxy(nil)

	if err == nil {
		t.Error("validation failure not caught")
	}

	simpleProxy.Path = "a path"

	_, err = simpleProxy.BuildProxy(nil)

	if err != nil {
		t.Error("validation failed when should pass")
	}
}

func TestProxy_ExtraSettings(t *testing.T) {
	simpleProxy := Proxy{
		Path:            "/api/payments/",
		TargetHost:      "notreal",
		ReplacementPath: "/api/",
	}

	baseProxy := Proxy{
		UseXForwardedFor: true,
		PreRequest: func(inbound, outbound *http.Request) (error, int) {
			if inbound.Header.Get("test") != "value" {
				t.Error("Header didn't pass through")
			}
			return errors.New("testing"), 503
		},
	}

	simpleProxy.Apply(baseProxy)

	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/payments/", nil)
	req.Header.Set("test", "value")

	pFunc(rr, req)

	if rr.Code != 503 {
		t.Error("Didn't return error")
	}
}

func TestProxy_ClientError(t *testing.T) {
	simpleProxy := Proxy{
		Path:       "/api/",
		TargetHost: "hate://go", // fake to trigger client error
	}
	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/", nil)

	pFunc(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Error("No good")
	}
}

func TestProxy_RouterError(t *testing.T) {
	simpleProxy := Proxy{
		Path: "/api/",
		Router: func(r *http.Request) (string, error) {
			return "", errors.New("Bad route")
		},
	}
	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/", nil)

	pFunc(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("Didn't get bad request")
	}
}

func TestProxy_Registration(t *testing.T) {
	simpleProxy := Proxy{
		Path:       "/api/",
		TargetHost: "asdf",
	}
	simpleProxy.Register()
}

func TestProxy_FailedRegistration(t *testing.T) {
	simpleProxy := Proxy{}
	simpleProxy.Register()
}

func TestProxy_ValidateRequest(t *testing.T) {
	simpleProxy := Proxy{
		Path:       "/api/",
		TargetHost: "doesn't matter",
	}

	proxyBase := Proxy{
		ValidateRequest: func(r *http.Request) (error, int) {
			return errors.New("Server Updating"), 503
		},
	}

	simpleProxy.Apply(proxyBase)

	pFunc, _ := simpleProxy.BuildProxy(nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/", nil)

	pFunc(rr, req)

	if rr.Code != 503 {
		t.Error("Not return error")
	}
}

func ExampleProxy_simple() {
	apiProxy := Proxy{
		Path:       "/api/",
		TargetHost: "http://localhost:5000",
	}
	apiProxy.Register()

	http.ListenAndServe(":8080", nil)
}

func ExampleProxy_router() {
	const writeHost = "https://read.example.zone"
	const readHost = "https://write.example.zone"

	methodMap := map[string]string{
		"GET":    readHost,
		"PUT":    writeHost,
		"POST":   writeHost,
		"DELETE": writeHost,
		"PATCH":  writeHost,
	}

	router := func(r *http.Request) (string, error) {
		if val, ok := methodMap[r.Method]; ok {
			return val + r.URL.RequestURI(), nil
		}
		return "", errors.New("Bad method")
	}

	apiProxy := Proxy{
		Path:   "/api/",
		Router: router,
	}

	apiProxy.Register()
	http.ListenAndServe(":8080", nil)
}

// Use a JSON file to control a number of proxies
func ExampleProxy_gateway() {
	f, err := ioutil.ReadFile("config.json")

	if err != nil {
		log.Fatal(err.Error())
	}

	var proxies []Proxy

	baseProxy := Proxy{
		PreRequest: func(inbound, outbound *http.Request) (error, int) {
			client := inbound.Header.Get("client")
			if client == "" {
				return errors.New("not available"), 503
			}
			return nil, 0
		},
	}

	json.Unmarshal(f, &proxies)

	for i := range proxies {
		proxy := proxies[i]
		proxy.Apply(baseProxy)
		log.Println("Registering proxy at " + proxy.Path)
		proxy.Register()
	}

	http.ListenAndServe(":8080", nil)
}
