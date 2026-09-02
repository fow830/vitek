package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"vitek/internal/tokens"
)

// DefaultHTTPProxyProbe GETs targetURL through proxyEndpoint.
func DefaultHTTPProxyProbe(ctx context.Context, proxyEndpoint, targetURL string) error {
	proxyEndpoint = strings.TrimSpace(proxyEndpoint)
	targetURL = strings.TrimSpace(targetURL)
	if proxyEndpoint == "" || targetURL == "" {
		return fmt.Errorf("%s", tokens.ErrMsgProxyProbeFailed)
	}
	proxyURL, err := url.Parse(proxyEndpoint)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: tokens.ProxyHealthTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPUserAgent)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", tokens.ErrMsgProxyProbeFailed, err)
	}
	defer res.Body.Close()
	if res.StatusCode > tokens.ProxyProbeMaxStatusCode {
		return fmt.Errorf("%s: status %d", tokens.ErrMsgProxyProbeFailed, res.StatusCode)
	}
	return nil
}
