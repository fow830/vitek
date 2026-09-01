package tokens

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

// MagicLinkConsumeURL is the mailer POST target on the app domain.
func MagicLinkConsumeURL() string {
	return HTTPSAppBase() + PathV1AuthMagicLinkConsume
}

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
)
