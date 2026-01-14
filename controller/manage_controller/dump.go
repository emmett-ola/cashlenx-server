package manage_controller

import (
	"net/http"
	"os"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/util"
)

// DumpDatabase creates a database dump and returns it as a download
func DumpDatabase(w http.ResponseWriter, r *http.Request) {
	// Verify ADMIN_TOKEN for dangerous operation
	if err := util.VerifyAdminTokenFromRequest(r); err != nil {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Create a temporary file for the dump
	filePath := "temp_dump.json"
	defer os.Remove(filePath) // Clean up after response

	// Create the dump
	_, err := manage_service.CreateBackup(filePath)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Disposition", "attachment; filename=dump.json")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Cache-Control", "must-revalidate")
	w.Header().Set("Pragma", "public")

	// Send the file
	w.WriteHeader(http.StatusOK)
	file, _ := os.Open(filePath)
	defer file.Close()

	// Copy file content to response
	util.SendFile(w, file)
}
