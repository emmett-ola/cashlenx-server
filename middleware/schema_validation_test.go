package middleware

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPISpecLoadsAndValidates(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate current test file")
	}

	specPath := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "docs", "openapi.yaml"))
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("failed to read OpenAPI spec: %v", err)
	}

	spec, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatalf("failed to load OpenAPI spec: %v", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		t.Fatalf("invalid OpenAPI spec: %v", err)
	}
}

func TestParseOpenAPIValidationErrors_ErrorAt(t *testing.T) {
	err := errors.New("request body has an error: doesn't match schema #/components/schemas/UserCreateRequest: Error at \"/username\": minimum string length is 6\nError at \"/password\": minimum string length is 6\n")
	got := parseOpenAPIValidationErrors(err)
	assertHasFieldError(t, got, "username", "minimum string length is 6")
	assertHasFieldError(t, got, "password", "minimum string length is 6")
}

func TestParseOpenAPIValidationErrors_Param(t *testing.T) {
	err := errors.New("parameter id in path has an error: value doesn't match pattern")
	got := parseOpenAPIValidationErrors(err)
	assertHasFieldError(t, got, "id", "value doesn't match pattern")
}

func assertHasFieldError(t *testing.T, got []map[string]string, field, want string) {
	t.Helper()
	for _, item := range got {
		if item["field"] == field {
			if item["message"] != want {
				t.Fatalf("expected %s error to be %q, got %q", field, want, item["message"])
			}
			return
		}
	}
	t.Fatalf("expected field %q error, got %v", field, got)
}

func TestJSONPointerToFieldPath(t *testing.T) {
	got := jsonPointerToFieldPath("/user/profile/name")
	if got != "user.profile.name" {
		t.Fatalf("expected %q, got %q", "user.profile.name", got)
	}
}
