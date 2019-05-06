package proxy

func ExampleOverlord_simple() {
	overlord := Overlord{}

	overlord.Proxies = append(overlord.Proxies, Proxy{
		Path:       "/api/",
		TargetHost: "http://localhost:123",
	})

	overlord.Takeover(8080)
}
