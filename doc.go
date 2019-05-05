/*
Package proxy is a simple proxy/gateway system for building http proxies.

The main purpose of this is to quickly build simple to advanced proxies with little to no effort.
It's used in my own projects as the root of an api gateway, centralizing access to a series of microservices
by route.

The main method of using this class is the Proxy object. Building up a Proxy object and calling the Register
method will build out the handler and register it with the default http.

There is also a static file hoster, for really specific situations.
*/
package proxy
