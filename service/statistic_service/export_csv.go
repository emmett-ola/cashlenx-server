package statistic_service

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExportToCSVForUser exports the user's cash flow data to CSV file
// Only exports data belonging to the specified user
func ExportToCSVForUser(fromDateInString, toDateInString, filePath, userId string) error {
	if filePath == "" {
		filePath = "./export.csv"
	}

	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Create CSV file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{"ID", "Category ID", "Category Name", "Date", "Type", "Amount", "Description", "Remark"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Determine if we're exporting all data or a date range
	isExportAll := fromDateInString == "" && toDateInString == ""

	if isExportAll {
		// Export all data for this user using pagination
		util.Logger.Infof("Exporting all cash flow data to CSV for user %s", userId)

		limit := 100
		offset := 0

		for {
			cashFlows := cash_flow_mapper.INSTANCE.GetAllCashFlowsByUser(userObjectId, limit, offset)
			if len(cashFlows) == 0 {
				break
			}

			// Write each cash flow to CSV
			for _, cashFlow := range cashFlows {
				categoryName := "Unknown"
				categoryType := ""
				categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userObjectId)
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
					categoryType = categoryEntity.Type
				}

				dateStr := util.FormatDateToStringWithoutDash(cashFlow.BelongsDate)
				record := []string{
					cashFlow.Id.Hex(),
					cashFlow.CategoryId.Hex(),
					categoryName,
					dateStr,
					categoryType,
					strconv.FormatFloat(cashFlow.Amount, 'f', 2, 64),
					cashFlow.Description,
					cashFlow.Remark,
				}

				if err := writer.Write(record); err != nil {
					return err
				}
			}

			offset += limit
		}
	} else {
		// Export by date range for this user
		queryDateCurrent := util.FormatDateFromStringWithoutDash(fromDateInString)
		queryDateEnded := util.FormatDateFromStringWithoutDash(toDateInString).AddDate(0, 0, 1)

		for queryDateEnded.After(queryDateCurrent) {
			cashFlowArray := cash_flow_mapper.INSTANCE.GetCashFlowsByBelongsDateAndUser(queryDateCurrent, userObjectId)

			for _, cashFlow := range cashFlowArray {
				categoryName := "Unknown"
				categoryType := ""
				categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userObjectId)
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
					categoryType = categoryEntity.Type
				}

				dateStr := util.FormatDateToStringWithoutDash(queryDateCurrent)
				record := []string{
					cashFlow.Id.Hex(),
					cashFlow.CategoryId.Hex(),
					categoryName,
					dateStr,
					categoryType,
					strconv.FormatFloat(cashFlow.Amount, 'f', 2, 64),
					cashFlow.Description,
					cashFlow.Remark,
				}

				if err := writer.Write(record); err != nil {
					return err
				}
			}

			queryDateCurrent = queryDateCurrent.AddDate(0, 0, 1)
		}
	}

	return nil
}
