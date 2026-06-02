package manage_controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

// DumpDatabase creates a database dump and returns it as a download
func DumpDatabase(w http.ResponseWriter, r *http.Request) {
	file, err := os.CreateTemp("", "cashlenx_dump_*.json")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("Failed to create temporary file", err))
		return
	}
	filePath := file.Name()
	file.Close()
	defer os.Remove(filePath)

	_, err = adminDumpDatabase(filePath)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
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
	outputFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer outputFile.Close()

	// Copy file content to response
	util.SendFile(w, outputFile)
}

// ExportUserData exports user data to a JSON file (renamed from DumpUserDatabase)
func ExportUserData(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("User not authenticated"))
		return
	}

	file, err := os.CreateTemp("", fmt.Sprintf("cashlenx_export_%s_*.json", userId))
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("Failed to create temporary file", err))
		return
	}
	filePath := file.Name()
	file.Close()
	defer os.Remove(filePath)

	_, err = userExportData(userId, filePath)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_%s.json", userId))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	w.Header().Set("Cache-Control", "must-revalidate")
	w.Header().Set("Pragma", "public")

	// Send the file
	w.WriteHeader(http.StatusOK)
	outputFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer outputFile.Close()

	// Copy file content to response
	util.SendFile(w, outputFile)
}
