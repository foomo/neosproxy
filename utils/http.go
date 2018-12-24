package utils

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func GetDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func GetDefaultTransportFor(name string) (t *http.Transport, err error) {
	proxyFunc, errProxyFunc := ProxyFuncFromEnvironmentFor(name)
	if errProxyFunc != nil {
		err = errProxyFunc
		return
	}
	t = GetDefaultTransport()
	if proxyFunc != nil {
		t.Proxy = proxyFunc
	}
	return
}

func ProxyFuncFromEnvironmentFor(name string) (func(req *http.Request) (*url.URL, error), error) {
	u := os.Getenv("HTTP_PROXY_FOR_" + name)
	if u == "" {
		return nil, nil
	}
	proxyURL, errProxyURL := url.Parse(u)
	if errProxyURL != nil {
		return nil, errProxyURL
	}
	if proxyURL.Scheme != "socks5" {
		return nil, errors.New("the proxy scheme must be socks5")
	}

	return func(req *http.Request) (u *url.URL, e error) {
		return proxyURL, nil
	}, nil
}
