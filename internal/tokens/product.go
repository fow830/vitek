package tokens

import "net/url"

// Product identity — single source for display and binaries.
const (
	ProductName      = "Vitek"
	ProductNameLocal = "Витёк"
	ModulePath       = "vitek"
	ProductDomainLanding = "vitek.tech"
	ProductDomainWWW     = "www.vitek.tech"
	ProductDomainApp     = "app.vitek.tech"
)

// ProductDomain is the marketing/landing apex (alias of ProductDomainLanding).
const ProductDomain = ProductDomainLanding

const HTTPSScheme = "https://"

// HTTPSAppBase returns the platform origin (mailer + redirects).
func HTTPSAppBase() string {
	return HTTPSScheme + ProductDomainApp
}

// MagicLinkConsumeURL is the JSON POST consume target on the app domain.
func MagicLinkConsumeURL() string {
	return HTTPSAppBase() + PathV1AuthMagicLinkConsume
}

// MagicLinkOpenURL returns the clickable magic link on the app domain (mailer SoT).
func MagicLinkOpenURL(rawToken string) string {
	v := url.Values{}
	v.Set(QueryParamToken, rawToken)
	return HTTPSAppBase() + PathV1AuthMagicLinkOpen + "?" + v.Encode()
}

// ProductDomainUnknownHost is used in contract fuzz (must never match real routes).
const ProductDomainUnknownHost = "evil.example"

// ProductEmail builds an email on ProductDomain (contracts / fixtures).
func ProductEmail(local string) string {
	return local + "@" + ProductDomain
}

// Production deploy names (vdserv / shared docker network).
const (
	ComposeNetworkProd           = "vitek_net"
	ComposeContainerPostgresProd = "vitek-postgres"
)

// Binary names (compose / CI image tags).
const (
	BinaryAPI    = "api"
	BinaryWorker = "worker"
)

// Worker log formats.
const (
	LogWorkerProxiesErr = "proxies: %v"
	LogWorkerTickActive = "tick: active_proxies=%d"
	LogWorkerListingSearchErr = "listing_search: %v"
)
