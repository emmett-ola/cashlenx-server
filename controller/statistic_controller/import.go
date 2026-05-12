package statistic_controller

import (
	"net/http"

	"github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/service/statistic_service"
	"github.com/macar-x/cashlenx-server/util"
)

// ImportData imports cash flow data from Excel file to user's account
func ImportData(w http.ResponseWriter, r *http.Request) {
	// Get user ID from request context
	userId, ok := r.Context().Value("user_id").(string)
	if !ok || userId == "" {
		util.ComposeJSONResponse(w, http.StatusUnauthorized, errors.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse query parameters
	filePath := r.URL.Query().Get("file_path")

	if filePath == "" {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("file_path is required"))
		return
	}

	// Call service to import data for this user
	err := statistic_service.ImportForUser(filePath, userId)
	if err != nil {
		util.ComposeErrorResponse(w, err)
		return
	}

	response := map[string]interface{}{
		"message":   "Data imported successfully",
		"file_path": filePath,
		"user_id":   userId,
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
