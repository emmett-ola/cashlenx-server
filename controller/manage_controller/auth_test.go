package manage_controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserDataHandlersRequireAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name    string
		handler http.HandlerFunc
		method  string
		path    string
	}{
		{"export user data", ExportUserData, http.MethodGet, "/user/database/backup"},
		{"import user data", ImportUserData, http.MethodPost, "/user/database/restore"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
		})
	}
}

func TestRestoreDatabaseRejectsMissingUpload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/admin/database/restore", nil)
	rec := httptest.NewRecorder()

	RestoreDatabase(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestImportUserDataRejectsMissingUpload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v0/user/database/restore", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-id"))
	rec := httptest.NewRecorder()

	ImportUserData(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
