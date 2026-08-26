package productlanguage

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Remaining Zitadel tokens are forbidden in executable command strings.
// The only exception is a third-party IdP type, which is not in cmd/.
var allowedTechnicalLiterals = map[string]map[string]struct{}{}

func TestCommandLanguageUsesNomenProductName(t *testing.T) {
	t.Parallel()

	root := "../.."
	fset := token.NewFileSet()
	err := fs.WalkDir(osDirFS(root), "cmd", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, root+"/"+path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if _, isImport := node.(*ast.ImportSpec); isImport {
				return false
			}
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil || !strings.Contains(strings.ToLower(value), "zitadel") {
				return true
			}
			if _, allowed := allowedTechnicalLiterals[path][value]; allowed {
				return true
			}
			t.Errorf("%s contains upstream product language in an executable string", fset.Position(literal.Pos()))
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// osDirFS is a seam so the scan stays rooted and behaves identically on the
// Linux and Windows repository runners.
func osDirFS(root string) fs.FS {
	return os.DirFS(root)
}
