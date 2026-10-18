package architecture

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

type goPackage struct {
	ImportPath string
	Imports    []string
}

func TestDomainPackagesDoNotImportInfrastructure(t *testing.T) {
	cmd := exec.Command("go", "list", "-json", "../../internal/...")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(out)))
	for decoder.More() {
		var pkg goPackage
		if err := decoder.Decode(&pkg); err != nil {
			t.Fatalf("decode go list output: %v", err)
		}
		if !strings.Contains(pkg.ImportPath, "/domain") {
			continue
		}
		for _, imported := range pkg.Imports {
			if forbiddenDomainImport(imported) {
				t.Errorf("domain package %s imports infrastructure dependency %s", pkg.ImportPath, imported)
			}
		}
	}
}

func forbiddenDomainImport(path string) bool {
	for _, forbidden := range []string{
		"database/sql",
		"github.com/gin-gonic/gin",
		"github.com/golang-jwt/jwt",
		"github.com/tencentyun/cos-go-sdk-v5",
	} {
		if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
			return true
		}
	}
	return false
}
