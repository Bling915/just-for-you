package architecture

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

const internalPrefix = "github.com/nanfang/just-for-you/server/internal/"

type goPackage struct {
	ImportPath string
	Imports    []string
}

func TestInternalPackagesUseGlobalLayers(t *testing.T) {
	for _, pkg := range listInternalPackages(t) {
		relative := strings.TrimPrefix(pkg.ImportPath, internalPrefix)
		top := strings.Split(relative, "/")[0]
		if !allowedTopLevel(top) {
			t.Errorf("package %s uses forbidden domain-first top-level directory %q", pkg.ImportPath, top)
		}
	}
}

func TestGlobalLayerDependencies(t *testing.T) {
	for _, pkg := range listInternalPackages(t) {
		from := layerOf(pkg.ImportPath)
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
		return to == "domain"
	case "interfaces":
		return to == "application" || to == "domain"
	case "infrastructure":
		return to == "application" || to == "domain"
	case "bootstrap", "architecture":
		return true
	default:
		return false
	}
}

func isStandardLibrary(importPath string) bool {
	return !strings.Contains(strings.Split(importPath, "/")[0], ".")
}
