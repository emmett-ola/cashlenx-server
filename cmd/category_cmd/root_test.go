package category_cmd

import (
	"path/filepath"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
)

func TestCategoryRootRequiresLoggedInUserBeforeSubcommandValidation(t *testing.T) {
	t.Setenv("CASHLENX_CLI_SESSION_FILE", filepath.Join(t.TempDir(), "session.json"))
	userId = ""

	if err := CategoryCmd.PersistentPreRunE(CategoryCmd, nil); err == nil {
		t.Fatal("expected missing CLI session error")
	}
}

func TestCategoryRootRequiresSubcommand(t *testing.T) {
	if err := CategoryCmd.RunE(CategoryCmd, nil); err == nil {
		t.Fatal("expected subcommand error")
	}
}

func TestCategoryTreePrintHelpers(t *testing.T) {
	if got := getBranchSymbol(true); got != "└── " {
		t.Fatalf("last branch symbol = %q", got)
	}
	if got := getBranchSymbol(false); got != "├── " {
		t.Fatalf("branch symbol = %q", got)
	}

	originalDepth := treeDeep
	defer func() { treeDeep = originalDepth }()

	treeDeep = 0
	printCategoryTreeNode(model.CategoryTree{
		Name: "Root",
		Children: []model.CategoryTree{
			{Name: "Child A"},
			{Name: "Child B"},
		},
	}, "", true, 1)

	treeDeep = 1
	printCategoryTreeNode(model.CategoryTree{
		Name:     "Root",
		Children: []model.CategoryTree{{Name: "Hidden Child"}},
	}, "", true, 1)
}
