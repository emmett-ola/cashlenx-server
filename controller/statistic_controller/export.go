package statistic_controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ExportData exports user's cash flow data to Excel file and returns it as binary download
func ExportData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse query parameters
	fromDate := r.URL.Query().Get("from_date") // YYYYMMDD or YYYY-MM-DD
	toDate := r.URL.Query().Get("to_date")     // YYYYMMDD or YYYY-MM-DD
	format := r.URL.Query().Get("format")      // xlsx, csv, pdf (default: xlsx)

	// Default format
	if format == "" {
		format = "xlsx"
	}

	// Validate format
	if format != "xlsx" && format != "csv" && format != "pdf" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("format must be xlsx, csv, or pdf"))
		return
	}

	// Create temporary directory for export files
	tempDir := filepath.Join(os.TempDir(), "cashlenx-exports")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to create temp directory", err))
		return
	}

	// Generate unique temporary file path
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("cashlenx-export-%s-%s.%s", userId[:8], timestamp, format)
	tempFilePath := filepath.Join(tempDir, filename)

	// Ensure cleanup of temporary file
	defer func() {
		os.Remove(tempFilePath)
	}()

	// Call service to export data for this user based on format
	var err error
	switch format {
	case "xlsx":
		err = statistic_service.ExportForUser(fromDate, toDate, tempFilePath, userId)
	case "csv":
		err = statistic_service.ExportToCSVForUser(fromDate, toDate, tempFilePath, userId)
	case "pdf":
		err = statistic_service.ExportToPDFForUser(fromDate, toDate, tempFilePath, userId)
	}

	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	// Read the exported file
	fileData, err := os.ReadFile(tempFilePath)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to read exported file", err))
		return
	}

	// Set appropriate headers for file download
	var contentType string
	switch format {
	case "xlsx":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		contentType = "text/csv"
	case "pdf":
		contentType = "application/pdf"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// Write file data to response
	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}
