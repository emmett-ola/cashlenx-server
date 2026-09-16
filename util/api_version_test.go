package util

import "testing"

func TestSupportedAPIVersionsKeepsV0CompatibilityForV1(t *testing.T) {
	original := GetConfigByKey("api.version")
	t.Cleanup(func() { SetConfigByKey("api.version", original) })
	SetConfigByKey("api.version", "v1")

	versions := SupportedAPIVersions()
	if len(versions) != 2 || versions[0] != "v1" || versions[1] != "v0" {
		t.Fatalf("SupportedAPIVersions() = %#v, want [v1 v0]", versions)
	}
	if got := CanonicalAPIPath("/api/v0/user/profile"); got != "/api/v1/user/profile" {
		t.Fatalf("CanonicalAPIPath() = %q, want /api/v1/user/profile", got)
	}
	if !IsPublicAPIPath("/api/v0/open/auth/login") || !IsAdminAPIPath("/api/v1/admin/user") {
		t.Fatal("expected v1 and v0 compatibility paths to be recognized")
	}
}

func TestCustomAPIVersionDoesNotEnableCompatibilityAlias(t *testing.T) {
	original := GetConfigByKey("api.version")
	t.Cleanup(func() { SetConfigByKey("api.version", original) })
	SetConfigByKey("api.version", "vtest")

	versions := SupportedAPIVersions()
	if len(versions) != 1 || versions[0] != "vtest" {
		t.Fatalf("SupportedAPIVersions() = %#v, want [vtest]", versions)
	}
	if got := CanonicalAPIPath("/api/v0/user/profile"); got != "/api/v0/user/profile" {
		t.Fatalf("custom version unexpectedly rewrote path to %q", got)
	}
}
