package architecture

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

const internalPrefix = "github.com/nanfang/just-for-you/server/internal/"

var legacyRoots = []string{
	"application/companionship", "application/moment", "application/together",
	"domain/companionship", "domain/moment", "domain/together",
	"infrastructure/ai", "infrastructure/notification",
	"infrastructure/persistence/mysql/companionship", "infrastructure/persistence/mysql/moment",
	"infrastructure/persistence/mysql/together",
	"interfaces/http/companionship", "interfaces/http/moment", "interfaces/http/shared", "interfaces/http/together",
}

type goPackage struct {
	ImportPath string
	Imports    []string
}

func TestInternalPackagesUseGlobalLayers(t *testing.T) {
	for _, pkg := range listInternalPackages(t) {
		relative := strings.TrimPrefix(pkg.ImportPath, internalPrefix)
		top := strings.Split(relative, "/")[0]
		if !isLegacyGlobalLayer(top) {
			continue
		}
		if !allowedTopLevel(top) {
			t.Errorf("package %s uses forbidden domain-first top-level directory %q", pkg.ImportPath, top)
		}
	}
}

func TestGlobalLayerDependencies(t *testing.T) {
	for _, pkg := range listInternalPackages(t) {
		from := layerOf(pkg.ImportPath)
		if !isLegacyGlobalLayer(from) {
			continue
		}
		for _, imported := range pkg.Imports {
			if !strings.HasPrefix(imported, internalPrefix) {
				if from == "domain" && !isStandardLibrary(imported) {
					t.Errorf("domain package %s imports non-standard dependency %s", pkg.ImportPath, imported)
				}
				continue
			}
			to := layerOf(imported)
			if !dependencyAllowed(from, to) {
				t.Errorf("forbidden dependency: %s (%s) -> %s (%s)", pkg.ImportPath, from, imported, to)
			}
		}
	}
}

func TestVerticalSliceDependencies(t *testing.T) {
	for _, pkg := range listInternalPackages(t) {
		relative := strings.TrimPrefix(pkg.ImportPath, internalPrefix)
		parts := strings.Split(relative, "/")
		if len(parts) == 0 {
			continue
		}
		if isLegacyGlobalLayer(parts[0]) {
			if !legacyRootAllowed(relative) {
				t.Errorf("legacy global layer is not allowlisted: %s", pkg.ImportPath)
			}
			continue
		}
		if parts[0] == "architecture" || parts[0] == "bootstrap" || parts[0] == "platform" {
			continue
		}
		if len(parts) < 2 || !isSliceLayer(parts[1]) {
			t.Errorf("package %s is not inside a domain slice layer", pkg.ImportPath)
			continue
		}
		fromContext, fromLayer := parts[0], parts[1]
		for _, imported := range pkg.Imports {
			if !strings.HasPrefix(imported, internalPrefix) {
				if fromLayer == "domain" && !isStandardLibrary(imported) {
					t.Errorf("domain package %s imports non-standard dependency %s", pkg.ImportPath, imported)
				}
				continue
			}
			importedParts := strings.Split(strings.TrimPrefix(imported, internalPrefix), "/")
			if len(importedParts) < 2 {
				continue
			}
			toContext, toLayer := importedParts[0], importedParts[1]
			if fromLayer == "domain" {
				t.Errorf("domain package %s imports internal package %s", pkg.ImportPath, imported)
			}
			if fromLayer == "application" && (toLayer == "infrastructure" || toLayer == "interfaces" || toContext == "bootstrap") {
				t.Errorf("application package %s imports outward package %s", pkg.ImportPath, imported)
			}
			if fromContext != toContext && (toLayer == "infrastructure" || toLayer == "interfaces") {
				t.Errorf("cross-context concrete dependency: %s -> %s", pkg.ImportPath, imported)
			}
		}
	}
}

func legacyRootAllowed(relative string) bool {
	for _, root := range legacyRoots {
		if relative == root || strings.HasPrefix(relative, root+"/") {
			return true
		}
	}
	return false
}

func isLegacyGlobalLayer(name string) bool {
	switch name {
	case "application", "domain", "infrastructure", "interfaces":
		return true
	default:
		return false
	}
}

func isSliceLayer(name string) bool {
	switch name {
	case "domain", "application", "infrastructure", "interfaces":
		return true
	default:
		return false
	}
}

func listInternalPackages(t *testing.T) []goPackage {
	t.Helper()
	cmd := exec.Command("go", "list", "-json", "../../internal/...")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	var packages []goPackage
	for decoder.More() {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
		packages = append(packages, pkg)
	}
	return packages
}

func allowedTopLevel(name string) bool {
	switch name {
	case "application", "architecture", "bootstrap", "domain", "infrastructure", "interfaces":
		return true
	default:
		return false
	}
}

func layerOf(importPath string) string {
	if !strings.HasPrefix(importPath, internalPrefix) {
		return "external"
	}
	return strings.Split(strings.TrimPrefix(importPath, internalPrefix), "/")[0]
}

func dependencyAllowed(from, to string) bool {
	if from == to || to == "architecture" {
		return true
	}
	switch from {
	case "domain":
		return false
	case "application":
		return to == "domain" || to == "platform" || to == "account" || to == "couple"
	case "interfaces":
		return to == "application" || to == "domain" || to == "platform" || to == "account" || to == "couple"
	case "infrastructure":
		return to == "application" || to == "domain" || to == "platform" || to == "account" || to == "couple"
	case "bootstrap", "architecture":
		return true
	default:
		return false
	}
}

func isStandardLibrary(importPath string) bool {
	return !strings.Contains(strings.Split(importPath, "/")[0], ".")
}
