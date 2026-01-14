package manage_controller

import (
	"net/http"
	"os"

	"github.com/macar-x/cashlenx-server/service/manage_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ExportData exports cash flow data to Excel and returns it as a download
func ExportData(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")

	// Create a temporary file for the export
	filePath := "temp_export.xlsx"
	defer os.Remove(filePath) // Clean up after response

	// Export the data
	err := manage_service.ExportService(fromDate, toDate, filePath)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, err)
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Disposition", "attachment; filename=export.xlsx")
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
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
