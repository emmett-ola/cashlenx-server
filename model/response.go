package model

// NewSuccessResponse creates a successful response with data
func NewSuccessResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
	}
}

// NewSuccessResponseWithMeta creates a successful response with data and metadata
func NewSuccessResponseWithMeta(data interface{}, meta map[string]interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"data": data,
	}
	if meta != nil {
		response["meta"] = meta
	}
	return response
}

// NewErrorResponse creates an error response
func NewErrorResponse(code, message string) map[string]interface{} {
	return map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    nil,
	}
}

// NewErrorResponseWithDetails creates an error response with additional details
func NewErrorResponseWithDetails(code, message, details string) map[string]interface{} {
	return map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    nil,
		"details": details,
	}
}

// NewFieldErrorResponse creates an error response for field validation
func NewFieldErrorResponse(code, message, field string) map[string]interface{} {
	errors := []map[string]string{{
		"field":   field,
		"message": message,
	}}
	return NewValidationErrorResponse(code, message, errors)
}

// NewFieldErrorResponseWithDetails creates an error response for field validation with details
func NewFieldErrorResponseWithDetails(code, message, field, details string) map[string]interface{} {
	errors := []map[string]string{{
		"field":   field,
		"message": details,
	}}
	return NewValidationErrorResponse(code, message, errors)
}

// NewValidationErrorResponse creates an error response with validation details
func NewValidationErrorResponse(code, message string, errors []map[string]string) map[string]interface{} {
	response := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    nil,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}
	return response
}
