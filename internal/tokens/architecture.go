package tokens

// Internal layer paths (repo-relative, POSIX).
const (
	PathTransport = "internal/transport"
	PathDomain    = "internal/domain"
	PathService   = "internal/service"
)

// ModuleImport builds a fully-qualified Go import path from a repo-relative layer path.
func ModuleImport(relPath string) string {
	return ModulePath + "/" + relPath
}

// ArchitectureDomainForbiddenImports are layers domain must never depend on.
var ArchitectureDomainForbiddenImports = []string{
	PathTransport,
	PathRepository,
	PathService,
}
