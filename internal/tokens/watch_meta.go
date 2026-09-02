package tokens

// Watch meta + SERP depth / fidelity tokens (wave C).
const (
	WatchMetaStatusPending = "PENDING"
	WatchMetaStatusReady   = "READY"
	WatchMetaStatusFailed  = "FAILED"

	ListingSearchSERPMaxItems  = 50
	ListingSearchSERPMaxPages  = 5
	ListingSearchSERPPageParam = "p"

	JSONFieldMetaStatus = "meta_status"
	JSONFieldMeta       = "meta"

	AppClassMetaPending = "meta-pending"
	AppClassMetaFail    = "meta-fail"
	AppClassMetaReady   = "meta-ready"

	FixtureNonAppleTitle = "vivo V30, 8/128 ГБ, 2 SIM"

	ListingFilterAppleLabel = "apple"
)

// ListingFilterAppleDenySubstrings titles matching these are dropped for apple-labelled filters.
var ListingFilterAppleDenySubstrings = []string{
	"vivo",
	"samsung",
	"xiaomi",
}

var SchemaWatchMetaStatuses = []string{
	WatchMetaStatusPending, WatchMetaStatusReady, WatchMetaStatusFailed,
}
