package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"vitek/internal/tokens"
)

// DefaultHTTPProxyProbe GETs targetURL through proxyEndpoint (or direct when endpoint is direct).
func DefaultHTTPProxyProbe(ctx context.Context, proxyEndpoint, targetURL string) error {
	proxyEndpoint = strings.TrimSpace(proxyEndpoint)
	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		return fmt.Errorf("%s", tokens.ErrMsgProxyProbeFailed)
	}
	client, err := HTTPClientViaProxy(proxyEndpoint, tokens.ProxyHealthTimeout)
	if err != nil {
		return fmt.Errorf("%s: %w", tokens.ErrMsgProxyProbeFailed, err)
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
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, int64(tokens.AvitoHTTPProbeDiscardBytes)))
	if !tokens.HTTPStatusOK(res.StatusCode) {
		return fmt.Errorf("%s: status %d", tokens.ErrMsgProxyProbeFailed, res.StatusCode)
	}
	return nil
}
