package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/macar-x/cashlenx-server/util"
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

func TestV0CompatibilityPathValidatesAgainstV1Schema(t *testing.T) {
	original := util.GetConfigByKey("api.version")
	t.Cleanup(func() { util.SetConfigByKey("api.version", original) })
	util.SetConfigByKey("api.version", "v1")

	req := httptest.NewRequest(http.MethodPost, "/api/v0/open/auth/login", strings.NewReader(`{"username":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	if err := validateRequest(req); err != nil {
		t.Fatalf("v0 compatibility request did not validate against v1 schema: %v", err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read original request body: %v", err)
	}
	if got := string(body); got != `{"username":"alice","password":"secret123"}` {
		t.Fatalf("original request body = %q after validation", got)
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
