package statistic_controller

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/macar-x/cashlenx-server/errors"
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

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("failed to parse uploaded file"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusBadRequest, errors.NewInvalidInputError("file is required"))
		return
	}
	defer file.Close()

	extension := filepath.Ext(header.Filename)
	tempFile, err := os.CreateTemp("", "cashlenx-statistic-import-*"+extension)
	if err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to create temporary import file", err))
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := io.Copy(tempFile, file); err != nil {
		tempFile.Close()
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to save uploaded file", err))
		return
	}
	if err := tempFile.Close(); err != nil {
		util.ComposeJSONResponse(w, http.StatusInternalServerError, errors.NewInternalError("failed to finalize uploaded file", err))
		return
	}

	// Call service to import data for this user
	err = importStatisticForUser(tempPath, userId)
	if err != nil {
		util.ComposeErrorResponse(w, r, err)
		return
	}

	response := map[string]interface{}{
		"message":  "Data imported successfully",
		"filename": header.Filename,
		"user_id":  userId,
	}
	util.ComposeJSONResponse(w, http.StatusOK, response)
}
