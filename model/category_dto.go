package model

type CategoryDTO struct {
	ParentId string `json:"parent_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Remark   string `json:"remark"`
}

// CreateCategoryRequest defines the request structure for creating a new category
type CreateCategoryRequest struct {
	ParentId string `json:"parent_id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Remark   string `json:"remark"`
}
