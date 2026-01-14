package util

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/macar-x/cashlenx-server/errors"
)

// ResponseError sends an error response using Gin context
func ResponseError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Code:    defaultCodeForStatus(statusCode),
		Message: message,
		Data:    nil,
	})
}

// ResponseSuccess sends a success response using Gin context
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    "SUCCESS",
		Data:    data,
		Message: "Operation successful",
	})
}

// ParseJSONRequest is a utility function to parse JSON requests
func ParseJSONRequest(r *http.Request, v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	return err
}

// Response defines the unified structure for all API responses
type Response struct {
	Code    string                 `json:"code"`
	Data    interface{}            `json:"data"`
	Message string                 `json:"message"`
	Meta    map[string]interface{} `json:"meta"`
	Extra   map[string]interface{} `json:"extra"`
	Errors  []map[string]string    `json:"errors"`
}

// ComposeJSONResponse is a utility function to write JSON responses with consistent format
func ComposeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Code:    "OK",
		Message: "",
		Data:    map[string]interface{}{},
		Meta:    map[string]interface{}{},
		Extra:   map[string]interface{}{},
		Errors:  []map[string]string{},
	}

	if statusCode >= 400 {
		response.Code = defaultCodeForStatus(statusCode)
	}

	// Check if data is an error
	if err, ok := data.(error); ok {
		// Check if it's an AppError from errors package
		if appErr, ok := err.(*errors.AppError); ok {
			// Set error code and message
			response.Code = string(appErr.Code)
			response.Message = appErr.Message
			errItem := map[string]string{
				"message": appErr.Message,
			}
			if appErr.Field != "" {
				errItem["field"] = appErr.Field
			}
			response.Errors = append(response.Errors, errItem)
		} else {
			response.Code = defaultCodeForStatus(statusCode)
			response.Message = err.Error()
			response.Errors = append(response.Errors, map[string]string{"message": err.Error()})
		}
	} else {
		if m, ok := data.(map[string]interface{}); ok {
			applyMapPayload(&response, statusCode, m)
		} else if m, ok := data.(map[string]string); ok {
			converted := make(map[string]interface{}, len(m))
			for k, v := range m {
				converted[k] = v
			}
			applyMapPayload(&response, statusCode, converted)
		} else if data != nil {
			response.Data = data
		}
	}

	// Encode the response as JSON and write it to the response writer
	json.NewEncoder(w).Encode(response)
}

func defaultCodeForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return string(errors.ErrInvalidInput)
	case http.StatusUnauthorized:
		return string(errors.ErrUnauthorized)
	case http.StatusForbidden:
		return string(errors.ErrForbidden)
	case http.StatusNotFound:
		return string(errors.ErrNotFound)
	case http.StatusConflict:
		return string(errors.ErrAlreadyExists)
	default:
		if statusCode >= 500 {
			return string(errors.ErrInternal)
		}
		return string(errors.ErrInternal)
	}
}

func applyMapPayload(response *Response, statusCode int, payload map[string]interface{}) {
	reserved := map[string]struct{}{
		"code":    {},
		"data":    {},
		"message": {},
		"meta":    {},
		"extra":   {},
		"errors":  {},
		"error":   {},
	}

	_, hasCode := payload["code"]
	_, hasData := payload["data"]
	_, hasMeta := payload["meta"]
	_, hasExtra := payload["extra"]
	_, hasErrors := payload["errors"]
	msg, hasMessage := payload["message"].(string)
	_, hasErrorKey := payload["error"]

	shouldTreatAsParts := hasCode || hasData || hasMeta || hasExtra || hasErrors || (hasMessage && (statusCode >= 400 || len(payload) > 1))

	if !shouldTreatAsParts {
		response.Data = payload
		return
	}

	if code, ok := payload["code"].(string); ok && code != "" {
		response.Code = code
	}
	if hasMessage {
		response.Message = msg
	}
	if hasData {
		response.Data = payload["data"]
	}
	if meta, ok := payload["meta"].(map[string]interface{}); ok {
		response.Meta = meta
	}
	if extra, ok := payload["extra"].(map[string]interface{}); ok {
		response.Extra = extra
	}
	if errorsVal, ok := payload["errors"]; ok {
		response.Errors = normalizeErrors(errorsVal)
	}

	for k, v := range payload {
		if _, ok := reserved[k]; ok {
			continue
		}
		response.Extra[k] = v
	}

	if hasErrorKey {
		errText, _ := payload["error"].(string)
		if response.Message == "" {
			response.Message = errText
		}
		if errText != "" && len(response.Errors) == 0 {
			response.Errors = append(response.Errors, map[string]string{"message": errText})
		}
		if statusCode >= 400 && (hasCode == false) {
			response.Code = defaultCodeForStatus(statusCode)
		}
	}

	if statusCode >= 400 && response.Message != "" && len(response.Errors) == 0 {
		response.Errors = append(response.Errors, map[string]string{"message": response.Message})
	}
}

func normalizeErrors(v interface{}) []map[string]string {
	if v == nil {
		return []map[string]string{}
	}

	if list, ok := v.([]map[string]string); ok {
		return list
	}

	if m, ok := v.(map[string]string); ok {
		out := make([]map[string]string, 0, len(m))
		for field, msg := range m {
			out = append(out, map[string]string{"field": field, "message": msg})
		}
		return out
	}

	if m, ok := v.(map[string]interface{}); ok {
		out := make([]map[string]string, 0, len(m))
		for field, raw := range m {
			if msg, ok := raw.(string); ok {
				out = append(out, map[string]string{"field": field, "message": msg})
			}
		}
		return out
	}

	if items, ok := v.([]interface{}); ok {
		out := make([]map[string]string, 0, len(items))
		for _, item := range items {
			if m, ok := item.(map[string]interface{}); ok {
				errItem := map[string]string{}
				if field, ok := m["field"].(string); ok {
					errItem["field"] = field
				}
				if msg, ok := m["message"].(string); ok {
					errItem["message"] = msg
				}
				if len(errItem) > 0 {
					out = append(out, errItem)
				}
			}
		}
		return out
	}

	if s, ok := v.(string); ok && s != "" {
		return []map[string]string{{"message": s}}
	}

	return []map[string]string{}
}

// SendFile sends a file as an HTTP response
func SendFile(w http.ResponseWriter, file io.Reader) {
	if _, err := io.Copy(w, file); err != nil {
		// If there's an error while sending the file, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("Failed to send file", err))
	}
}
