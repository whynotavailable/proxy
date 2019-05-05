package main

import (
	"net/http"
	"sync"
	"time"
)

// OAuth struct for dealing with client creds
type OAuth struct {
	mutex        sync.RWMutex
	client       http.Client
	clientID     string
	clientSecret string
	token        string
	tokenExpire  int64
}

// GetToken get the token
func (o *OAuth) GetToken() string {
	token := ""
	var expiration int64

	o.mutex.RLock()
	token = o.token
	expiration = o.tokenExpire
	o.mutex.RUnlock()

	currentTime := time.Now().Unix()

	if currentTime > expiration {
		if expiration == 0 {
			// never gotten before, so run synchronously
			o.tokenGetter()
		} else {
			// we already have one, go get it in the background
			go o.tokenGetter()
		}
	}

	return token
}

// the actual method for getting the token
func (o *OAuth) tokenGetter() {
	o.setToken("my token")
}

func (o *OAuth) setToken(token string) {
	o.mutex.Lock()
	o.token = token
	o.tokenExpire = time.Now().Unix() + (10 * 60)
	o.mutex.Unlock()
}
