package manage_controller

import (
	"io"
	"net/http"
	"os"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/util"
)

// RestoreDatabase restores database from a dump file uploaded via multipart form
func RestoreDatabase(w http.ResponseWriter, r *http.Request) {
	// Verify ADMIN_TOKEN for dangerous operation
	if err := util.VerifyAdminTokenFromRequest(r); err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max file size
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("Failed to parse form data"))
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("No file uploaded or invalid file"))
		return
	}
	defer file.Close()

	// Create a temporary file to save the uploaded dump
	tempFile, err := os.CreateTemp("", "restore_dump_*.json")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("Failed to create temporary file", err))
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tempFile, file); err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("Failed to save uploaded file", err))
		return
	}

	// Restore from the temporary file
	stats, err := manage_service.RestoreBackup(tempFile.Name())
	if err != nil {
		// Return error along with statistics
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]interface{}{
			"error":   err.Error(),
			"stats":   stats,
			"message": "Database restore failed",
		})
		return
	}

	// Return success response with statistics
	util.ComposeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Database restored successfully from file: " + handler.Filename,
		"stats":   stats,
	})
}
