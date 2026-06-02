package manage_controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/macar-x/cashlenx-server/service/manage_service"
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

func TestDumpDatabaseReturnsGeneratedFile(t *testing.T) {
	original := adminDumpDatabase
	adminDumpDatabase = func(filePath string) (manage_service.OperationStats, error) {
		if err := os.WriteFile(filePath, []byte(`{"ok":true}`), 0600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		return manage_service.OperationStats{Users: manage_service.EntityStats{Success: 1}}, nil
	}
	t.Cleanup(func() { adminDumpDatabase = original })

	req := httptest.NewRequest(http.MethodGet, "/api/v0/admin/database/backup", nil)
	rec := httptest.NewRecorder()

	DumpDatabase(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("Content-Disposition") != "attachment; filename=dump.json" {
		t.Fatalf("content disposition = %q", rec.Header().Get("Content-Disposition"))
	}
	if rec.Body.String() != `{"ok":true}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestDumpDatabaseReturnsServiceError(t *testing.T) {
	original := adminDumpDatabase
	adminDumpDatabase = func(filePath string) (manage_service.OperationStats, error) {
		return manage_service.OperationStats{}, errors.New("dump failed")
	}
	t.Cleanup(func() { adminDumpDatabase = original })

	req := httptest.NewRequest(http.MethodGet, "/api/v0/admin/database/backup", nil)
	rec := httptest.NewRecorder()

	DumpDatabase(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestExportUserDataReturnsGeneratedFile(t *testing.T) {
	const userID = "user-id"
	original := userExportData
	userExportData = func(serviceUserID string, filePath string) (manage_service.OperationStats, error) {
		if serviceUserID != userID {
			t.Fatalf("service user id = %q, want %q", serviceUserID, userID)
		}
		if err := os.WriteFile(filePath, []byte(`{"user":true}`), 0600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		return manage_service.OperationStats{CashFlows: manage_service.EntityStats{Success: 2}}, nil
	}
	t.Cleanup(func() { userExportData = original })

	req := httptest.NewRequest(http.MethodGet, "/api/v0/user/database/backup", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	rec := httptest.NewRecorder()

	ExportUserData(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get("Content-Disposition") != "attachment; filename=export_user-id.json" {
		t.Fatalf("content disposition = %q", rec.Header().Get("Content-Disposition"))
	}
	if rec.Body.String() != `{"user":true}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRestoreDatabasePassesUploadedFileToService(t *testing.T) {
	var capturedPath string
	original := adminRestoreDatabase
	adminRestoreDatabase = func(filePath string) (manage_service.OperationStats, error) {
		capturedPath = filePath
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != `{"backup":true}` {
			t.Fatalf("uploaded content = %q", data)
		}
		return manage_service.OperationStats{Users: manage_service.EntityStats{Success: 1}}, nil
	}
	t.Cleanup(func() { adminRestoreDatabase = original })

	req := multipartRequest(t, "/api/v0/admin/database/restore", "backup.json", `{"backup":true}`)
	rec := httptest.NewRecorder()

	RestoreDatabase(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if capturedPath == "" {
		t.Fatal("expected service to receive temp file path")
	}
	var response struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("response JSON decode failed: %v", err)
	}
	if response.Message != "Database restored successfully from file: backup.json" {
		t.Fatalf("message = %q", response.Message)
	}
}

func TestImportUserDataPassesUserAndUploadedFileToService(t *testing.T) {
	const userID = "user-id"
	var capturedUserID string
	original := userImportData
	userImportData = func(serviceUserID string, filePath string) (manage_service.OperationStats, error) {
		capturedUserID = serviceUserID
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if string(data) != `{"backup":true}` {
			t.Fatalf("uploaded content = %q", data)
		}
		return manage_service.OperationStats{Categories: manage_service.EntityStats{Success: 3}}, nil
	}
	t.Cleanup(func() { userImportData = original })

	req := multipartRequest(t, "/api/v0/user/database/restore", "user-backup.json", `{"backup":true}`)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	rec := httptest.NewRecorder()

	ImportUserData(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if capturedUserID != userID {
		t.Fatalf("service user id = %q, want %q", capturedUserID, userID)
	}
}

func multipartRequest(t *testing.T, path string, filename string, content string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile failed: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("part Write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer Close failed: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
