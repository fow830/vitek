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
// Do not include "esim" — it false-positives on real "iPhone … SIM + eSIM" titles.
var ListingFilterAppleDenySubstrings = []string{
	"vivo",
	"samsung",
	"xiaomi",
	"чехол",
	"чехлы",
	"case",
	"airtag",
	"airpods",
	"whoop",
	"ремешок",
	"браслет",
	"кольцо",
	"oura",
	"fitbit",
	"ваучер",
	"рукав",
}

// ListingFilterApplePhoneRequireSubstrings: for apple + mobilnye_telefony, title must match one.
var ListingFilterApplePhoneRequireSubstrings = []string{
	"iphone",
}

// ListingFilterPhoneCategoryPath marks phone SERP paths that need the phone title require-list.
const ListingFilterPhoneCategoryPath = "mobilnye_telefony"

var SchemaWatchMetaStatuses = []string{
	WatchMetaStatusPending, WatchMetaStatusReady, WatchMetaStatusFailed,
}
