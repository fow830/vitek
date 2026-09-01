package tokens

// Product identity — single source for display and binaries.
const (
	ProductName      = "Vitek"
	ProductNameLocal = "Витёк"
	ModulePath       = "vitek"
	ProductDomain    = "vitek.tech"
	ProductDomainWWW = "www.vitek.tech"
)

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
