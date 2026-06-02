package category_cmd

import (
	"path/filepath"
	"testing"

	"github.com/macar-x/cashlenx-server/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestCategoryPersistentPreRunUsesAuthenticatedUser(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	original := requireCategoryUser
	requireCategoryUser = func(target *string) error {
		*target = authenticatedUserID
		return nil
	}
	t.Cleanup(func() {
		requireCategoryUser = original
		userId = ""
	})

	if err := CategoryCmd.PersistentPreRunE(CategoryCmd, nil); err != nil {
		t.Fatalf("PersistentPreRunE returned error: %v", err)
	}
	if userId != authenticatedUserID {
		t.Fatalf("userId = %q, want %q", userId, authenticatedUserID)
	}
}

func TestCategoryCreateListUpdateDeleteAndTreePassStateToServices(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	categoryID := primitive.NewObjectID()
	parentID := primitive.NewObjectID().Hex()
	type mutationArgs struct {
		id       string
		name     string
		typ      string
		remark   string
		parentID string
		userID   string
		force    bool
	}
	var createArgs, updateArgs, deleteArgs mutationArgs
	var listArgs struct {
		userID string
		typ    string
		limit  int
		offset int
	}
	var treeArgs struct {
		userID string
		typ    string
	}

	originalCreate := createCategoryForUser
	createCategoryForUser = func(name, typ, remark, serviceParentID, serviceUserID string) (model.CategoryEntity, error) {
		createArgs = mutationArgs{name: name, typ: typ, remark: remark, parentID: serviceParentID, userID: serviceUserID}
		return model.CategoryEntity{Id: categoryID, Name: name, Type: typ, Remark: remark}, nil
	}
	originalList := queryCategoriesForUser
	queryCategoriesForUser = func(serviceUserID, typ string, limit, offset int) ([]model.CategoryEntity, int64, error) {
		listArgs.userID = serviceUserID
		listArgs.typ = typ
		listArgs.limit = limit
		listArgs.offset = offset
		return []model.CategoryEntity{{Id: categoryID, Name: "Food", Type: typ}}, 1, nil
	}
	originalUpdate := updateCategoryForUser
	updateCategoryForUser = func(id, name, typ, remark, serviceParentID, serviceUserID string) (model.CategoryEntity, error) {
		updateArgs = mutationArgs{id: id, name: name, typ: typ, remark: remark, parentID: serviceParentID, userID: serviceUserID}
		return model.CategoryEntity{Id: categoryID, Name: name, Type: typ, Remark: remark}, nil
	}
	originalDelete := deleteCategoryForUser
	deleteCategoryForUser = func(id, serviceUserID string, serviceForce bool) (model.CategoryEntity, error) {
		deleteArgs = mutationArgs{id: id, userID: serviceUserID, force: serviceForce}
		return model.CategoryEntity{Id: categoryID, Name: "Food"}, nil
	}
	originalTree := queryCategoryTreeForUser
	queryCategoryTreeForUser = func(serviceUserID, typ string) ([]model.CategoryTree, error) {
		treeArgs.userID = serviceUserID
		treeArgs.typ = typ
		return []model.CategoryTree{{Id: categoryID.Hex(), Name: "Food", Type: typ}}, nil
	}
	t.Cleanup(func() {
		createCategoryForUser = originalCreate
		queryCategoriesForUser = originalList
		updateCategoryForUser = originalUpdate
		deleteCategoryForUser = originalDelete
		queryCategoryTreeForUser = originalTree
		resetCategoryCommandState()
	})

	userId = authenticatedUserID
	categoryName, catType, categoryRemark, parentPlainId = "Food", "expense", "Daily meals", parentID
	if err := createCmd.RunE(createCmd, nil); err != nil {
		t.Fatalf("create RunE returned error: %v", err)
	}
	if createArgs.name != "Food" || createArgs.typ != "expense" || createArgs.remark != "Daily meals" || createArgs.parentID != parentID || createArgs.userID != authenticatedUserID {
		t.Fatalf("create args = %+v", createArgs)
	}

	listCategoryType, categoryLimit, categoryOffset = "expense", 7, 14
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatalf("list RunE returned error: %v", err)
	}
	if listArgs.userID != authenticatedUserID || listArgs.typ != "expense" || listArgs.limit != 7 || listArgs.offset != 14 {
		t.Fatalf("list args = %+v", listArgs)
	}

	plainId, categoryName, catType, categoryRemark, parentPlainId = categoryID.Hex(), "Meals", "expense", "Updated", parentID
	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatalf("update RunE returned error: %v", err)
	}
	if updateArgs.id != categoryID.Hex() || updateArgs.name != "Meals" || updateArgs.typ != "expense" || updateArgs.remark != "Updated" || updateArgs.parentID != parentID || updateArgs.userID != authenticatedUserID {
		t.Fatalf("update args = %+v", updateArgs)
	}

	force = true
	if err := deleteCmd.RunE(deleteCmd, nil); err != nil {
		t.Fatalf("delete RunE returned error: %v", err)
	}
	if deleteArgs.id != categoryID.Hex() || deleteArgs.userID != authenticatedUserID || !deleteArgs.force {
		t.Fatalf("delete args = %+v", deleteArgs)
	}

	treeCategoryType = "expense"
	if err := treeCmd.RunE(treeCmd, nil); err != nil {
		t.Fatalf("tree RunE returned error: %v", err)
	}
	if treeArgs.userID != authenticatedUserID || treeArgs.typ != "expense" {
		t.Fatalf("tree args = %+v", treeArgs)
	}
}

func TestCategoryQueryPassesCriteriaToServices(t *testing.T) {
	const authenticatedUserID = "507f1f77bcf86cd799439011"
	categoryID := primitive.NewObjectID()
	parentID := primitive.NewObjectID().Hex()
	var byIDArgs, byNameArgs struct {
		value  string
		userID string
	}
	var childArgs struct {
		parentID string
		userID   string
		typ      string
	}

	originalByID := queryCategoryByIDForUser
	queryCategoryByIDForUser = func(id, serviceUserID string) (model.CategoryEntity, error) {
		byIDArgs.value = id
		byIDArgs.userID = serviceUserID
		return model.CategoryEntity{Id: categoryID, Name: "Food"}, nil
	}
	originalByName := queryCategoryByNameForUser
	queryCategoryByNameForUser = func(name, serviceUserID string) (model.CategoryEntity, error) {
		byNameArgs.value = name
		byNameArgs.userID = serviceUserID
		return model.CategoryEntity{Id: categoryID, Name: name}, nil
	}
	originalChildren := queryCategoryChildrenForUser
	queryCategoryChildrenForUser = func(serviceParentID, serviceUserID, typ string) ([]model.CategoryEntity, error) {
		childArgs.parentID = serviceParentID
		childArgs.userID = serviceUserID
		childArgs.typ = typ
		return []model.CategoryEntity{{Id: categoryID, Name: "Child", Type: typ}}, nil
	}
	t.Cleanup(func() {
		queryCategoryByIDForUser = originalByID
		queryCategoryByNameForUser = originalByName
		queryCategoryChildrenForUser = originalChildren
		resetCategoryCommandState()
	})

	userId = authenticatedUserID
	plainId = categoryID.Hex()
	if err := queryCmd.RunE(queryCmd, nil); err != nil {
		t.Fatalf("query by id RunE returned error: %v", err)
	}
	if byIDArgs.value != categoryID.Hex() || byIDArgs.userID != authenticatedUserID {
		t.Fatalf("by id args = %+v", byIDArgs)
	}

	plainId, categoryName = "", "Food"
	if err := queryCmd.RunE(queryCmd, nil); err != nil {
		t.Fatalf("query by name RunE returned error: %v", err)
	}
	if byNameArgs.value != "Food" || byNameArgs.userID != authenticatedUserID {
		t.Fatalf("by name args = %+v", byNameArgs)
	}

	categoryName, parentPlainId, catType = "", parentID, "income"
	if err := queryCmd.RunE(queryCmd, nil); err != nil {
		t.Fatalf("query children RunE returned error: %v", err)
	}
	if childArgs.parentID != parentID || childArgs.userID != authenticatedUserID || childArgs.typ != "income" {
		t.Fatalf("child args = %+v", childArgs)
	}
}

func resetCategoryCommandState() {
	plainId = ""
	parentPlainId = ""
	categoryName = ""
	catType = ""
	categoryRemark = ""
	userId = ""
	force = false
	categoryLimit = 50
	categoryOffset = 0
	listCategoryType = ""
	treeDeep = 0
	treeCategoryType = ""
}
