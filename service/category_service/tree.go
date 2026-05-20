package category_service

import (
	"fmt"
	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetCategoryTreeByUser builds category tree for a specific user
func GetCategoryTreeByUser(userId, categoryType string) ([]model.CategoryTree, error) {
	return defaultCategoryService().GetCategoryTreeByUser(userId, categoryType)
}

func (s *CategoryService) GetCategoryTreeByUser(userId, categoryType string) ([]model.CategoryTree, error) {
	// Validate user ID
	userObjectId := util.Convert2ObjectId(userId)
	if userObjectId == primitive.NilObjectID {
		return nil, errors.NewInvalidInputError("invalid user ID")
	}

	// Validate category type
	if categoryType != "income" && categoryType != "expense" && categoryType != "" {
		return nil, errors.NewInvalidInputError("category type must be 'income', 'expense', or empty")
	}

	// Get root categories with user ID and type filter
	var rootCategories []model.CategoryEntity
	var err error

	if categoryType == "" {
		rootCategories, err = s.categoryMapper.GetRootCategoriesByUser(userObjectId)
	} else {
		rootCategories, err = s.categoryMapper.GetRootCategoriesByUserAndType(userObjectId, categoryType)
	}

	if err != nil {
		return nil, err
	}

	// Build category tree recursively
	var categoryTreeList []model.CategoryTree
	for _, root := range rootCategories {
		categoryTree := s.buildUserCategoryTree(root, userObjectId, categoryType)
		categoryTreeList = append(categoryTreeList, categoryTree)
	}

	return categoryTreeList, nil
}

func buildUserCategoryTree(parent model.CategoryEntity, userId primitive.ObjectID, categoryType string) model.CategoryTree {
	return defaultCategoryService().buildUserCategoryTree(parent, userId, categoryType)
}

func (s *CategoryService) buildUserCategoryTree(parent model.CategoryEntity, userId primitive.ObjectID, categoryType string) model.CategoryTree {
	// Convert entity to tree node
	categoryTree := model.CategoryTree{
		Id:       parent.Id.Hex(),
		ParentId: parent.ParentId.Hex(),
		Name:     parent.Name,
		Type:     parent.Type,
		Children: []model.CategoryTree{},
	}

	// Get children with user ID and type filter
	var children []model.CategoryEntity
	var err error

	if categoryType == "" {
		children, err = s.categoryMapper.GetCategoriesByParentIdAndUser(parent.Id, userId)
	} else {
		children, err = s.categoryMapper.GetCategoriesByParentIdUserAndType(parent.Id, userId, categoryType)
	}

	if err != nil {
		fmt.Printf("Error getting children for category %s: %v", parent.Id.Hex(), err)
		return categoryTree
	}

	// Recursively build children nodes
	for _, child := range children {
		childTree := s.buildUserCategoryTree(child, userId, categoryType)
		categoryTree.Children = append(categoryTree.Children, childTree)
	}

	return categoryTree
}
