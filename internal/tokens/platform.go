package tokens

// Auth: Magic Link only (DB + HTTP). Roles USER/ADMIN via sqlc enums.
const AuthMethodMagicLink = "MAGIC_LINK"

// AuthMethodsAllowlist is the only permitted auth method (contract-enforced).
var AuthMethodsAllowlist = []string{AuthMethodMagicLink}

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

// ProductServiceSpec is the catalog SoT mirrored by product_services seed + contracts.
type ProductServiceSpec struct {
	Code    string
	Title   string
	Shipped bool
}

// ProductServiceCatalog is the full catalog (shipped + reserved).
var ProductServiceCatalog = []ProductServiceSpec{
	{Code: ServiceCodeListingSearch, Title: ServiceTitleListingSearch, Shipped: true},
	{Code: ServiceCodeListingWarmup, Title: ServiceTitleListingWarmup, Shipped: false},
}

// ShippedServiceCodes derived from ProductServiceCatalog.
func ShippedServiceCodes() []string {
	out := make([]string, 0, len(ProductServiceCatalog))
	for _, s := range ProductServiceCatalog {
		if s.Shipped {
			out = append(out, s.Code)
		}
	}
	return out
}

// ReservedServiceCodes derived from ProductServiceCatalog.
func ReservedServiceCodes() []string {
	out := make([]string, 0, len(ProductServiceCatalog))
	for _, s := range ProductServiceCatalog {
		if !s.Shipped {
			out = append(out, s.Code)
		}
	}
	return out
}

// Plan task limits — must match plan_limits seed (FREE/PRO/ULTRA).
const (
	PlanMaxTasksFREE  int32 = 1
	PlanMaxTasksPRO   int32 = 20
	PlanMaxTasksULTRA int32 = 100
)

// SchemaPlatformTables: DB tables that form the platform surface (HTTP may lag).
var SchemaPlatformTables = []string{
	"product_services",
	"avito_accounts",
	"proxies",
	"users",
	"user_service_entitlements",
	"magic_link_challenges",
	"sessions",
}

// TableUsersPasswordColumn must never exist (Magic Link only).
const TableUsersPasswordColumn = "password"

// ForbiddenPackagePathFragments: Day-0 negative surface (must not appear yet).
var ForbiddenPackagePathFragments = []string{
	"telegram",
	"tgbot",
	"aiogram",
	"browser_session",
	"feed_matcher",
}

// ForbiddenArtifactPaths are stale generated/hand paths that must not exist on disk.
var ForbiddenArtifactPaths = []string{
	"web/admin/face.html",
}

// ForbiddenGoModPathFragments: go.mod must not pull these until contracted.
var ForbiddenGoModPathFragments = []string{
	"github.com/go-telegram",
	"gopkg.in/telebot",
	"github.com/redis/go-redis",
	"github.com/go-redis/redis",
}
