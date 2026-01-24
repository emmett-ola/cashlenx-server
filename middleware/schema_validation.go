package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
)

var (
	openapi      *openapi3.T
	routesRouter routers.Router
)

func init() {
	// Only initialize schema validation if enabled
	if util.GetConfigByKey("api.schema.validation") != "true" {
		return
	}

	_, currentFile, _, ok := runtime.Caller(0)
	specPath := "docs/openapi.yaml"
	if ok {
		baseDir := filepath.Dir(currentFile)
		specPath = filepath.Clean(filepath.Join(baseDir, "..", "docs", "openapi.yaml"))
	}

	data, err := os.ReadFile(specPath)
	if err != nil {
		util.Logger.Errorw("Failed to load OpenAPI spec", "error", err, "path", specPath)
		util.SetConfigByKey("api.schema.validation", "false")
		return
	}

	// Parse OpenAPI spec
	spec, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		util.Logger.Errorw("Failed to parse OpenAPI spec", "error", err)
		util.SetConfigByKey("api.schema.validation", "false")
		return
	}

	// Validate OpenAPI spec
	if err := spec.Validate(context.Background()); err != nil {
		util.Logger.Errorw("Invalid OpenAPI spec", "error", err)
		util.SetConfigByKey("api.schema.validation", "false")
		return
	}

	// Create router for path matching
	openapi = spec
	routesRouter, err = gorillamux.NewRouter(openapi)
	if err != nil {
		util.Logger.Errorw("Failed to create OpenAPI router", "error", err)
		util.SetConfigByKey("api.schema.validation", "false")
		openapi = nil
		routesRouter = nil
		return
	}
}

// SchemaValidation middleware to validate requests against OpenAPI schema
func SchemaValidation(next http.Handler) http.Handler {
	// Return original handler if schema validation is disabled
	if util.GetConfigByKey("api.schema.validation") != "true" {
		return next
	}
	if openapi == nil || routesRouter == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip health and version endpoints
		// Get API prefix from config to skip health/version check correctly
		apiVersion := util.GetConfigByKey("api.version")
		if apiVersion == "" {
			apiVersion = "v0"
		}
		apiPrefix := "/api/" + apiVersion

		if strings.HasPrefix(r.URL.Path, apiPrefix+"/open/health") || strings.HasPrefix(r.URL.Path, apiPrefix+"/open/version") {
			next.ServeHTTP(w, r)
			return
		}

		// Validate request against schema
		if err := validateRequest(r); err != nil {
			// Log the detailed error for debugging
			util.Logger.Errorw("Schema validation failed", "error", err, "path", r.URL.Path, "method", r.Method)

			validationErrors := parseOpenAPIValidationErrors(err)

			response := model.NewValidationErrorResponse("VALIDATION_ERROR", "Request validation failed", validationErrors)
			util.ComposeJSONResponse(w, http.StatusBadRequest, response)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

var (
	reErrorAt = regexp.MustCompile(`Error at "([^"]+)":\s*([^\n]+)`)
	reParam   = regexp.MustCompile(`parameter\s+([^\s]+)\s+in\s+(query|path|header|cookie)\s+has an error:\s*([^\n]+)`)
)

func parseOpenAPIValidationErrors(err error) []map[string]string {
	result := map[string]string{}

	errs := []error{err}
	var multiErr openapi3.MultiError
	if errors.As(err, &multiErr) {
		errs = multiErr
	}

	texts := make([]string, 0, len(errs))
	for _, e := range errs {
		if e == nil {
			continue
		}
		texts = append(texts, e.Error())

		var validationErr *openapi3filter.ValidationError
		if errors.As(e, &validationErr) {
			msg := strings.TrimSpace(validationErr.Detail)
			if msg == "" {
				msg = strings.TrimSpace(validationErr.Error())
			}
			if i := strings.IndexByte(msg, '\n'); i >= 0 {
				msg = strings.TrimSpace(msg[:i])
			}

			if validationErr.Source != nil {
				if validationErr.Source.Parameter != "" {
					if _, exists := result[validationErr.Source.Parameter]; !exists && msg != "" {
						result[validationErr.Source.Parameter] = msg
					}
				}
				if validationErr.Source.Pointer != "" {
					field := jsonPointerToFieldPath(validationErr.Source.Pointer)
					if field == "" {
						field = "body"
					}
					if _, exists := result[field]; !exists && msg != "" {
						result[field] = msg
					}
				}
			}

			if validationErr.Detail != "" {
				texts = append(texts, validationErr.Detail)
			}
		}
	}

	for _, text := range texts {
		for _, match := range reErrorAt.FindAllStringSubmatch(text, -1) {
			if len(match) < 3 {
				continue
			}
			field := jsonPointerToFieldPath(match[1])
			msg := strings.TrimSpace(match[2])
			if field == "" {
				field = "body"
			}
			if _, exists := result[field]; !exists && msg != "" {
				result[field] = msg
			}
		}
	}

	if len(result) > 0 {
		return fieldErrorMapToList(result)
	}

	for _, text := range texts {
		match := reParam.FindStringSubmatch(text)
		if len(match) >= 4 {
			param := strings.TrimSpace(match[1])
			msg := strings.TrimSpace(match[3])
			if param != "" && msg != "" {
				result[param] = msg
				return fieldErrorMapToList(result)
			}
		}
	}

	fallback := err.Error()
	if idx := strings.Index(fallback, "has an error:"); idx >= 0 {
		fallback = strings.TrimSpace(fallback[idx+len("has an error:"):])
	}
	if idx := strings.IndexByte(fallback, '\n'); idx >= 0 {
		fallback = strings.TrimSpace(fallback[:idx])
	}
	if fallback == "" {
		fallback = "request validation failed"
	}
	result["body"] = fallback
	return fieldErrorMapToList(result)
}

func fieldErrorMapToList(m map[string]string) []map[string]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]map[string]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, map[string]string{
			"field":   k,
			"message": m[k],
		})
	}
	return out
}

func jsonPointerToFieldPath(pointer string) string {
	pointer = strings.TrimSpace(pointer)
	if pointer == "" {
		return ""
	}
	if pointer == "/" {
		return ""
	}
	if strings.HasPrefix(pointer, "#") {
		if i := strings.Index(pointer, "/"); i >= 0 {
			pointer = pointer[i:]
		} else {
			return ""
		}
	}
	pointer = strings.TrimPrefix(pointer, "/")
	if pointer == "" {
		return ""
	}
	parts := strings.Split(pointer, "/")
	for i, p := range parts {
		p = strings.ReplaceAll(p, "~1", "/")
		p = strings.ReplaceAll(p, "~0", "~")
		parts[i] = p
	}
	return strings.Join(parts, ".")
}

// validateRequest validates incoming request against OpenAPI schema
func validateRequest(r *http.Request) error {
	// Create a copy of the request with modified URL to match OpenAPI server URL
	// This ensures validation works regardless of the actual hostname/port
	rCopy := r.Clone(context.Background())

	// Use the first server URL from the spec or default to http://localhost:8080
	var serverURL string
	if len(openapi.Servers) > 0 {
		serverURL = openapi.Servers[0].URL
	} else {
		serverURL = "http://localhost:8080"
	}

	// Adjust server URL with API version if needed
	// This is a simplified approach. Ideally, we should parse the OpenAPI server URL and replace variables.
	// But since we control the code and spec, we can assume the spec might have /api/v0 or we need to strip it for validation
	// Actually, kin-openapi should handle path matching against the spec.
	// If we changed routes in server.go to include /v0/, we MUST change openapi.yaml to include /v0/ as well.
	// Then standard validation should work.

	// Parse the server URL
	server, err := r.URL.Parse(serverURL)
	if err != nil {
		return err
	}

	// Keep the original path, query, fragment, etc.
	rCopy.URL.Scheme = server.Scheme
	rCopy.URL.Host = server.Host

	// Find matching route using the modified URL
	route, pathParams, err := routesRouter.FindRoute(rCopy)
	if err != nil {
		return err
	}

	// Create validation context
	ctx := context.Background()

	// Configure validation options
	options := &openapi3filter.Options{
		ExcludeRequestBody: false,
	}

	// Skip body validation for file upload endpoints to avoid "value must be a string" error
	// This is a known issue with kin-openapi validation for multipart/form-data
	if strings.HasSuffix(r.URL.Path, "/restore") {
		options.ExcludeRequestBody = true
	}

	// Validate request using the original request but matched route
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    options,
	}

	return openapi3filter.ValidateRequest(ctx, requestValidationInput)
}
