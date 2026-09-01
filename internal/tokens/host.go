package tokens

import "strings"

// NormalizeHost strips port and lowercases Host header values.
func NormalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host
}

// IsLandingHost is true for apex and www marketing hosts.
func IsLandingHost(host string) bool {
	host = NormalizeHost(host)
	return host == ProductDomainLanding || host == ProductDomainWWW
}

// IsAppHost is true for the platform host.
func IsAppHost(host string) bool {
	return NormalizeHost(host) == ProductDomainApp
}

// HTTPLandingHosts are marketing Host header values (contracts / probes).
var HTTPLandingHosts = []string{
	ProductDomainLanding,
	ProductDomainWWW,
}
