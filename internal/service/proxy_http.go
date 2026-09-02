package service

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"vitek/internal/tokens"
)

// HTTPClientViaProxy builds an HTTP client that reaches the internet via proxyEndpoint.
// SOCKS5 uses remote DNS (curl socks5h behavior) so Avito hostnames resolve on the proxy side.
func HTTPClientViaProxy(proxyEndpoint string, timeout time.Duration) (*http.Client, error) {
	proxyEndpoint = strings.TrimSpace(proxyEndpoint)
	if proxyEndpoint == "" {
		return &http.Client{Timeout: timeout}, nil
	}
	u, err := url.Parse(proxyEndpoint)
	if err != nil {
		return nil, err
	}
	transport, err := httpTransportForProxy(u)
	if err != nil {
		return nil, err
	}
	return &http.Client{Timeout: timeout, Transport: transport}, nil
}

func httpTransportForProxy(u *url.URL) (*http.Transport, error) {
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "socks5", "socks5h", "socks":
		var auth *proxy.Auth
		if u.User != nil {
			pass, _ := u.User.Password()
			auth = &proxy.Auth{User: u.User.Username(), Password: pass}
		}
		base := &net.Dialer{Timeout: tokens.ProxyHealthTimeout}
		d, err := proxy.SOCKS5("tcp", u.Host, auth, base)
		if err != nil {
			return nil, err
		}
		cd, ok := d.(proxy.ContextDialer)
		if !ok {
			return nil, fmt.Errorf("%s: socks dialer without context", tokens.ErrMsgProxyProbeFailed)
		}
		return &http.Transport{DialContext: cd.DialContext}, nil
	default:
		return &http.Transport{Proxy: http.ProxyURL(u)}, nil
	}
}
