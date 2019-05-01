# Proxy

This is a proxy base. What that means, is it's a small library to simplify making proxies.
An example is when building a microservice backend. The proxy can direct connections
to the correct places, adding in any authentication required.

There are a lot of reasons to use a system like this, and probably most of them
can be covered with Nginx. The purpose of this is to handle those situations
where that's not possible.
