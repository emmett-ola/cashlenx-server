package category_cmd

import (
	"fmt"

	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/service/category_service"
	"github.com/spf13/cobra"
)

var (
	treeDeep         int
	treeCategoryType string
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "list categories in tree structure",
	Long:  `List all categories in a hierarchical tree structure with optional depth control.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get category tree from service
		categoryTree, err := category_service.GetCategoryTreeByUser(userId, treeCategoryType)
		if err != nil {
			return err
		}

		// Print the tree structure
		for _, root := range categoryTree {
			printCategoryTreeNode(root, "", true, 1)
		}

		return nil
	},
}

// printCategoryTreeNode recursively prints the category tree with indentation
func printCategoryTreeNode(node model.CategoryTree, indent string, isLast bool, depth int) {
	// Print current node
	fmt.Printf("%s%s%s\n", indent, getBranchSymbol(isLast), node.Name)

	if treeDeep > 0 && depth >= treeDeep {
		return
	}

	// Print children
	childIndent := indent
	if isLast {
		childIndent += "    "
	} else {
		childIndent += "│   "
	}

	for i, child := range node.Children {
		printCategoryTreeNode(child, childIndent, i == len(node.Children)-1, depth+1)
	}
}

// getBranchSymbol returns the appropriate branch symbol based on whether the node is last
func getBranchSymbol(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

func init() {
	treeCmd.Flags().IntVarP(
		&treeDeep, "deep", "d", 0, "maximum depth of the tree (0 means unlimited)")
	treeCmd.Flags().StringVarP(
		&treeCategoryType, "type", "t", "", "filter by category type (income/expense)")
	CategoryCmd.AddCommand(treeCmd)
}
