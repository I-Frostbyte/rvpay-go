package payments

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestNoDirectPawaPayCall verifies that the clients/payments package never
// directly imports or calls any pawaPay package. The pawaPay boundary is
// enforced by the Transactions service; Clients must delegate all payment
// execution and reconciliation to Transactions via gRPC.
func TestNoDirectPawaPayCall(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse clients/payments directory: %v", err)
	}

	pkg, ok := pkgs["payments"]
	if !ok {
		t.Fatal("payments package not found")
	}

	for _, f := range pkg.Files {
		for _, imp := range f.Imports {
			path := imp.Path.Value
			// Reject any import path containing "pawapay" or "pawa".
			if strings.Contains(path, "pawapay") || strings.Contains(path, "/pawa") {
				t.Errorf("file %s imports %s which violates the pawaPay boundary: Clients must not directly call pawaPay", fset.Position(f.Pos()).Filename, path)
			}
		}
	}
}
