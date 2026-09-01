package tokens

// LRT documentation policy — paths relative to module root, POSIX slashes.

// AllowedProseDocs: only these prose files may exist (.md / .txt / .rst scan).
var AllowedProseDocs = map[string]struct{}{
	PathREADME: {},
	PathDEPLOY: {},
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
