package tokens

// Auth: both end-users and admins sign in via Magic Link only (no passwords).
// Schema skeleton only until HTTP/UI is built — see magic_link_challenges.
const AuthMethodMagicLink = "MAGIC_LINK"

// Product services: Vitek is a multi-service platform. Listing search is the
// first shipped service — not the hard-coded identity of the product.
// Roles (USER/ADMIN) live only in Postgres enum + sqlc (repository.UserRole*).
const (
	ServiceCodeListingSearch = "listing_search"
	ServiceCodeListingWarmup = "listing_warmup" // reserved; not shipped yet
)

const (
	ServiceTitleListingSearch = "Similar listings search"
	ServiceTitleListingWarmup = "Listing warmup (clicks)"
)

// ShippedServiceCodes must match product_services rows with shipped = true.
var ShippedServiceCodes = []string{
	ServiceCodeListingSearch,
}

// ReservedServiceCodes may exist in catalog with shipped = false; must never ship yet.
var ReservedServiceCodes = []string{
	ServiceCodeListingWarmup,
}

// AdminManagedTables: schema-ready resources for the future admin web UI
// (Magic Link auth, Avito accounts, proxies, entitlements). No HTTP admin yet.
var AdminManagedTables = []string{
	"product_services",
	"avito_accounts",
	"proxies",
	"users",
	"user_service_entitlements",
	"magic_link_challenges",
}

// TableUsersPasswordColumn must never exist (Magic Link only).
const TableUsersPasswordColumn = "password"
