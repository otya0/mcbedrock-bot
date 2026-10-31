// Checks the direction of the imports inside internal: record imports nothing
// of this module, act and arrive import only record and serverlink, and act
// never imports act.
package mcbedrock

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const internalPath = "github.com/otya0/mcbedrock-bot/internal/"

func TestImportsGoOneWay(t *testing.T) {
	allowed := map[string][]string{
		"record": {},
		"act":    {"record/", "serverlink"},
		"arrive": {"record/", "serverlink"},
	}
	for folder, allowedPrefixes := range allowed {
		err := filepath.WalkDir(filepath.Join("internal", folder), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") {
				return err
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range file.Imports {
				imported := strings.Trim(spec.Path.Value, `"`)
				if !strings.HasPrefix(imported, internalPath) {
					continue
				}
				imported = strings.TrimPrefix(imported, internalPath)
				if !hasAnyPrefix(imported, allowedPrefixes) {
					t.Errorf("%s imports internal/%s, but %s may import only %v", path, imported, folder, allowedPrefixes)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func hasAnyPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
