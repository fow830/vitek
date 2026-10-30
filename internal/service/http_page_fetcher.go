package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vitek/internal/tokens"
)

// HTTPPageFetcher loads Avito HTML via SOCKS/HTTP proxy (no Chrome).
// Factory rod http-mode uses AvitoClient+Chrome instead; this type remains for
// cookie/storage_state helpers and direct HTTP SERP tests.
type HTTPPageFetcher struct {
	httpBase string
	timeout  time.Duration
}

func NewHTTPPageFetcher(httpBase string) *HTTPPageFetcher {
	return &HTTPPageFetcher{
		httpBase: strings.TrimSpace(httpBase),
		timeout:  tokens.AvitoHTTPClientTimeout,
	}
}

func (f *HTTPPageFetcher) FetchHTML(ctx context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error) {
	pageURL = tokens.ListingSearchRewriteHTTPBase(pageURL, f.httpBase)
	client, err := HTTPClientViaProxy(proxyEndpoint, f.timeout)
	if err != nil {
		return "", newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPUserAgent)
	req.Header.Set(tokens.HeaderAccept, tokens.MIMETextHTML)
	req.Header.Set(tokens.HeaderAcceptLanguage, tokens.AvitoHTTPAcceptLanguage)
	if cookie := CookieHeaderFromStorageState(userDataDir, pageURL); cookie != "" {
		req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPBrowserUserAgent)
		req.Header.Set(tokens.HeaderCookie, cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(tokens.AvitoHTTPMaxBodyBytes)))
	if err != nil {
		return "", err
	}
	html := string(body)
	state := tokens.ClassifyAvitoHTTPResponse(res.StatusCode, html)
	if tokens.AvitoFetchStateIsContent(state) {
		if strings.TrimSpace(html) == "" {
			return "", newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, res.StatusCode, false)
		}
		return html, nil
	}
	return "", errorFromAvitoFetchState(state, res.StatusCode)
}

type storageStateFile struct {
	Cookies []storageStateCookie `json:"cookies"`
}

type storageStateCookie struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Domain  string  `json:"domain"`
	Path    string  `json:"path"`
	Expires float64 `json:"expires"`
	Secure  bool    `json:"secure"`
}

// CookieHeaderFromStorageState builds a Cookie header from a Playwright storage_state.json
// stored in the binding profile directory. Missing file yields an empty header.
func CookieHeaderFromStorageState(userDataDir, pageURL string) string {
	dir := strings.TrimSpace(userDataDir)
	if dir == "" {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(dir, tokens.AvitoStorageStateFile))
	if err != nil || len(raw) == 0 {
		return ""
	}
	var state storageStateFile
	if err := json.Unmarshal(raw, &state); err != nil {
		return ""
	}
	u, err := url.Parse(pageURL)
	if err != nil || u.Hostname() == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	path := u.Path
	if path == "" {
		path = "/"
	}
	https := strings.EqualFold(u.Scheme, "https")
	now := float64(time.Now().Unix())
	allow := tokens.AvitoCookieAllowlistSet()
	parts := make([]string, 0, len(tokens.AvitoCookieAllowlist))
	seen := map[string]struct{}{}
	for _, c := range state.Cookies {
		name := strings.TrimSpace(c.Name)
		if name == "" || strings.ContainsAny(name, "\r\n;") || strings.ContainsAny(c.Value, "\r\n;") {
			continue
		}
		if _, ok := allow[name]; !ok {
			continue
		}
		if c.Expires > 0 && c.Expires < now {
			continue
		}
		if c.Secure && !https {
			continue
		}
		if !cookieDomainMatches(host, c.Domain) || !cookiePathMatches(path, c.Path) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		parts = append(parts, name+"="+c.Value)
		joined := strings.Join(parts, "; ")
		if len(joined) > tokens.AvitoCookieHeaderMaxBytes {
			parts = parts[:len(parts)-1]
			break
		}
	}
	return strings.Join(parts, "; ")
}

func cookieDomainMatches(host, domain string) bool {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return false
	}
	if strings.HasPrefix(domain, ".") {
		root := strings.TrimPrefix(domain, ".")
		return host == root || strings.HasSuffix(host, "."+root)
	}
	return host == domain
}

func cookiePathMatches(reqPath, cookiePath string) bool {
	if cookiePath == "" || cookiePath == "/" {
		return true
	}
	if reqPath == cookiePath {
		return true
	}
	if strings.HasPrefix(reqPath, cookiePath) {
		if strings.HasSuffix(cookiePath, "/") {
			return true
		}
		return len(reqPath) > len(cookiePath) && reqPath[len(cookiePath)] == '/'
	}
	return false
}
