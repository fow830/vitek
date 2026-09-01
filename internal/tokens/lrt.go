package tokens

// LRT documentation policy — paths relative to module root, POSIX slashes.
const (
	DocREADME = "README.md"
	DocDEPLOY = "DEPLOY.md"
	DocLRT    = ".cursor/rules/lrt.mdc"
)

// AllowedProseDocs: only these prose files may exist (.md / .txt / .rst scan).
// DocLRT is .mdc and is not matched by the prose extension scan; listed for humans/AI.
var AllowedProseDocs = map[string]struct{}{
	DocREADME: {},
	DocDEPLOY: {},
}

// ProseExtensions scanned by LRT docs policy.
var ProseExtensions = map[string]struct{}{
	".md":  {},
	".txt": {},
	".rst": {},
}

// IgnoredScanDirs skipped while walking the repo for LRT policy.
var IgnoredScanDirs = map[string]struct{}{
	"vendor":       {},
	".git":         {},
	"node_modules": {},
	"bin":          {},
	"dist":         {},
}
