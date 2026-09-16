package util

import "strings"

const (
	DefaultAPIVersion = "v1"
	LegacyAPIVersion  = "v0"
)

// GetAPIVersion returns the configured canonical API version.
func GetAPIVersion() string {
	version := strings.Trim(strings.TrimSpace(GetConfigByKey("api.version")), "/")
	if version == "" {
		return DefaultAPIVersion
	}
	return version
}

// SupportedAPIVersions returns the route versions served by this process.
// The stable v1 contract keeps the frozen v0 route as a compatibility alias
// for previously shipped clients. Custom versions remain isolated.
func SupportedAPIVersions() []string {
	version := GetAPIVersion()
	if version == DefaultAPIVersion {
		return []string{DefaultAPIVersion, LegacyAPIVersion}
	}
	return []string{version}
}

func APIPrefix(version string) string {
	return "/api/" + strings.Trim(version, "/")
}

func IsPublicAPIPath(path string) bool {
	return hasVersionedPathPrefix(path, "/open/")
}

func IsAdminAPIPath(path string) bool {
	return hasVersionedPathPrefix(path, "/admin/")
}

func IsHealthOrVersionPath(path string) bool {
	for _, version := range SupportedAPIVersions() {
		prefix := APIPrefix(version) + "/open/"
		if path == prefix+"health" || path == prefix+"version" {
			return true
		}
	}
	return false
}

// CanonicalAPIPath maps a supported compatibility path to the canonical path
// used by the OpenAPI document.
func CanonicalAPIPath(path string) string {
	if GetAPIVersion() != DefaultAPIVersion {
		return path
	}
	legacyPrefix := APIPrefix(LegacyAPIVersion)
	if path == legacyPrefix {
		return APIPrefix(DefaultAPIVersion)
	}
	if strings.HasPrefix(path, legacyPrefix+"/") {
		return APIPrefix(DefaultAPIVersion) + strings.TrimPrefix(path, legacyPrefix)
	}
	return path
}

// APIVersionFromPath returns a supported version embedded in a request path.
func APIVersionFromPath(path string) string {
	for _, version := range SupportedAPIVersions() {
		prefix := APIPrefix(version)
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return version
		}
	}
	return GetAPIVersion()
}

func hasVersionedPathPrefix(path, suffix string) bool {
	for _, version := range SupportedAPIVersions() {
		if strings.HasPrefix(path, APIPrefix(version)+suffix) {
			return true
		}
	}
	return false
}
