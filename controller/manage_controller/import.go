package manage_controller

import (
	"io"
	"net/http"
	"os"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ImportData handles the import of cash flow data from Excel files
func ImportData(w http.ResponseWriter, r *http.Request) {
	// Verify ADMIN_TOKEN for dangerous operation
	if err := util.VerifyAdminTokenFromRequest(r); err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Parse the multipart form with a 10 MB file size limit
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "File too large, max 10MB",
		})
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "Error retrieving file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Create a temporary file to store the uploaded file
	tempFile, err := os.CreateTemp("", "cashlenx_import_*.xlsx")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Error creating temporary file: " + err.Error(),
		})
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name()) // Clean up after processing
	}()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Error saving file: " + err.Error(),
		})
		return
	}

	// Import the data
	err = manage_service.ImportService(tempFile.Name())
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Import failed: " + err.Error(),
		})
		return
	}

	// Return success response
	util.ComposeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "File imported successfully: " + handler.Filename,
	})
}
