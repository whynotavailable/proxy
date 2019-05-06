package proxy

import (
	"fmt"
	"net/http"
	"strings"
)

/*
Overlord is a proxy management system that takes over the http pipeline to make
advanded proxies easier. Not ready to be used.
*/
type Overlord struct {
	Proxies                []Proxy
	GobalRequestValidators [](func(*http.Request) (error, int))
	proxyCache             map[string](func(w http.ResponseWriter, r *http.Request))
}

// ServeHTTP actual do the thing
func (o *Overlord) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for i := range o.GobalRequestValidators {
		err, code := o.GobalRequestValidators[i](r)
		if err != nil {
			http.Error(w, err.Error(), code)
			return
		}
	}

	currentRoute := ""

	for key := range o.proxyCache {
		if strings.Index(r.URL.RequestURI(), key) == 0 {
			if len(key) > len(currentRoute) {
				currentRoute = key
			}
		}
	}

	if currentRoute != "" {
		o.proxyCache[currentRoute](w, r)
	}

	http.Error(w, "Not Found", 404)
}

func (o *Overlord) buildProxyCache() {
	for i := range o.Proxies {
		proxy := o.Proxies[i]
		pFunc, err := proxy.BuildProxy()

		if err == nil {
			o.proxyCache[proxy.Path] = pFunc
		}
	}
}

// Takeover launches the proxy manager
func (o *Overlord) Takeover(port uint16) {
	o.buildProxyCache()
	http.ListenAndServe(fmt.Sprintf(":%d", port), o)
}
