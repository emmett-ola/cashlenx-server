package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/macar-x/cashlenx-server/errors"
)

func TestComposeJSONResponse_FieldAlreadyExists(t *testing.T) {
	recorder := httptest.NewRecorder()

	ComposeJSONResponse(recorder, http.StatusConflict, errors.NewFieldAlreadyExistsError("username", "username is already taken"))

	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != string(errors.ErrAlreadyExists) {
		t.Fatalf("expected code %q, got %q", string(errors.ErrAlreadyExists), resp.Code)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("expected errors length 1, got %d", len(resp.Errors))
	}
	if resp.Errors[0]["field"] != "username" {
		t.Fatalf("expected errors[0].field to be %q, got %q", "username", resp.Errors[0]["field"])
	}
	if resp.Errors[0]["message"] != "username is already taken" {
		t.Fatalf("expected errors[0].message to be %q, got %q", "username is already taken", resp.Errors[0]["message"])
	}
}

func TestComposeJSONResponse_ResponseShapeWithDataAndMeta(t *testing.T) {
	recorder := httptest.NewRecorder()

	ComposeJSONResponse(recorder, http.StatusOK, map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "1"},
		},
		"meta": map[string]interface{}{
			"total_count": 1,
			"limit":       50,
			"offset":      0,
		},
	})

	var resp Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Code != "OK" {
		t.Fatalf("expected code %q, got %q", "OK", resp.Code)
	}
	if resp.Message != "" {
		t.Fatalf("expected message %q, got %q", "", resp.Message)
	}
	if len(resp.Errors) != 0 {
		t.Fatalf("expected empty errors, got %v", resp.Errors)
	}
	if resp.Meta["total_count"] != float64(1) {
		t.Fatalf("expected meta.total_count to be %v, got %v", float64(1), resp.Meta["total_count"])
	}
	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(data) != 1 {
		t.Fatalf("expected data length 1, got %d", len(data))
	}
}

func TestStatusCodeForError_AppError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "validation", err: errors.NewValidationError("field is required"), want: http.StatusBadRequest},
		{name: "unauthorized", err: errors.NewUnauthorizedError("not authenticated"), want: http.StatusUnauthorized},
		{name: "forbidden", err: errors.NewForbiddenError("admin role required"), want: http.StatusForbidden},
		{name: "not found", err: errors.NewNotFoundError("user not found"), want: http.StatusNotFound},
		{name: "conflict", err: errors.NewAlreadyExistsError("username already exists"), want: http.StatusConflict},
		{name: "internal", err: errors.NewInternalError("failed", nil), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCodeForError(tt.err); got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}

func TestStatusCodeForError_ServiceMessageFallback(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "already exists", err: stdError("category already exists"), want: http.StatusConflict},
		{name: "not found", err: stdError("category not found or access denied"), want: http.StatusNotFound},
		{name: "invalid input", err: stdError("invalid user ID"), want: http.StatusBadRequest},
		{name: "unknown failure", err: stdError("failed to create category"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusCodeForError(tt.err); got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}

type stdError string

func (e stdError) Error() string {
	return string(e)
}
