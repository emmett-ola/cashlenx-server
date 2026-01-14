package statistic_service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	defaultSheetName = "report"
	defaultRowTitle  = []string{"Id", "CategoryId", "CategoryName", "BelongsDate", "FlowType", "Amount", "Description"}
)

// ExportForUser exports the user's cash flow data to Excel
// Only exports data belonging to the specified user
func ExportForUser(fromDateInString, toDateInString, filePath, userId string) error {
	if filePath == "" {
		filePath = "./export.xlsx"
	}

	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("invalid user ID")
	}

	fromDate := util.FormatDateFromStringWithoutDash(fromDateInString)
	toDate := util.FormatDateFromStringWithoutDash(toDateInString)
	if err := isExportRequiredFieldSatisfied(fromDate, toDate, filePath); err != nil {
		return err
	}

	file := createExcelFile()
	exportDataForUser(file, fromDateInString, toDateInString, userObjectId)
	saveExcelFile(file, filePath)
	return nil
}

func isExportRequiredFieldSatisfied(fromDate, toDate time.Time, filePath string) error {
	if !util.IsDateTimeEmpty(fromDate) && !util.IsDateTimeEmpty(toDate) && fromDate.After(toDate) {
		return errors.New("from_date should be before to_date")
	}
	if !strings.HasSuffix(filePath, ".xlsx") {
		return errors.New("file_path should end with '.xlsx'")
	}
	return nil
}

func createExcelFile() *excelize.File {
	file := excelize.NewFile()
	// Create a worksheet
	index, _ := file.NewSheet(defaultSheetName)
	// Set the default worksheet of the workbook
	file.SetActiveSheet(index)
	// Set the value of the cell
	writeExcelRow(file, defaultSheetName, "A1", "Start Time")
	writeExcelRow(file, defaultSheetName, "B1", time.Now())
	// Delete the default Sheet1 worksheet
	_ = file.DeleteSheet("Sheet1")

	return file
}

func saveExcelFile(file *excelize.File, filePath string) {
	// Save the workbook according to the specified path
	writeExcelRow(file, defaultSheetName, "A2", "Ended Time")
	writeExcelRow(file, defaultSheetName, "B2", time.Now())
	if err := file.SaveAs(filePath); err != nil {
		util.Logger.Errorln(err)
	}
}

func exportDataForUser(file *excelize.File, fromDateInString, toDateInString string, userId primitive.ObjectID) {
	cashFlowRowIndex := 1

	// Determine if we're exporting all data or a date range
	isExportAll := fromDateInString == "" && toDateInString == ""

	if isExportAll {
		// Export all data for this user using pagination
		util.Logger.Infof("Exporting all cash flow data for user %s", userId.Hex())

		// Group all cash flows by year-month
		dataByYearMonth := make(map[string][]model.CashFlowEntity)

		// Use pagination to get all cash flows for this user
		limit := 100 // Fetch 100 records at a time
		offset := 0

		for {
			cashFlows := cash_flow_mapper.INSTANCE.GetAllCashFlowsByUser(userId, limit, offset)
			if len(cashFlows) == 0 {
				break // No more records
			}

			// Group by year-month
			for _, cashFlow := range cashFlows {
				yearMonth := cashFlow.BelongsDate.Format("200601")
				dataByYearMonth[yearMonth] = append(dataByYearMonth[yearMonth], cashFlow)
			}

			offset += limit
		}

		// Write each year-month group to separate sheets
		for yearMonth, cashFlows := range dataByYearMonth {
			// Create sheet for each year-month
			_, _ = file.NewSheet(yearMonth)
			cashFlowRowIndex = 1

			// Write headers
			writeExcelRow(file, yearMonth, "A1", defaultRowTitle[0])
			writeExcelRow(file, yearMonth, "B1", defaultRowTitle[1])
			writeExcelRow(file, yearMonth, "C1", defaultRowTitle[2])
			writeExcelRow(file, yearMonth, "D1", defaultRowTitle[3])
			writeExcelRow(file, yearMonth, "E1", defaultRowTitle[4])
			writeExcelRow(file, yearMonth, "F1", defaultRowTitle[5])
			writeExcelRow(file, yearMonth, "G1", defaultRowTitle[6])

			// Write data
			for _, cashFlow := range cashFlows {
				cashFlowRowIndex++
				cashFlowIndexInString := strconv.Itoa(cashFlowRowIndex)
				dateStr := cashFlow.BelongsDate.Format("20060102")
				writeExcelRow(file, yearMonth, "A"+cashFlowIndexInString, cashFlow.Id.Hex())
				writeExcelRow(file, yearMonth, "B"+cashFlowIndexInString, cashFlow.CategoryId.Hex())

				// Get category name with user context
				categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userId)
				categoryName := "Unknown"
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
				}
				writeExcelRow(file, yearMonth, "C"+cashFlowIndexInString, categoryName)

				writeExcelRow(file, yearMonth, "D"+cashFlowIndexInString, dateStr)
				writeExcelRow(file, yearMonth, "E"+cashFlowIndexInString, cashFlow.FlowType)
				writeExcelRow(file, yearMonth, "F"+cashFlowIndexInString, cashFlow.Amount)
				writeExcelRow(file, yearMonth, "G"+cashFlowIndexInString, cashFlow.Description)
			}
			util.Logger.Debugf("Exported %d records for %s", len(cashFlows), yearMonth)
		}
	} else {
		// Export by date range for this user
		queryDateCurrent := util.FormatDateFromStringWithoutDash(fromDateInString)
		queryDateEnded := util.FormatDateFromStringWithoutDash(toDateInString).AddDate(0, 0, 1)
		currentYearAndMonth := "nil"

		for queryDateEnded.After(queryDateCurrent) {
			cashFlowArray := cash_flow_mapper.INSTANCE.GetCashFlowsByBelongsDateAndUser(queryDateCurrent, userId)
			if len(cashFlowArray) == 0 {
				util.Logger.Debugf("%s's flow is empty.\n", util.FormatDateToStringWithoutDash(queryDateCurrent))
				queryDateCurrent = queryDateCurrent.AddDate(0, 0, 1)
				continue
			}

			queryDateCurrentInString := util.FormatDateToStringWithoutDash(queryDateCurrent)
			util.Logger.Debugf("%s's flow is exporting.\n", queryDateCurrentInString)

			// If the year-month changes, initialize a new sheet
			newYearAndMonth := queryDateCurrentInString[0:6]
			if newYearAndMonth != currentYearAndMonth {
				currentYearAndMonth = newYearAndMonth

				_, _ = file.NewSheet(currentYearAndMonth)

				cashFlowRowIndex = 1
				writeExcelRow(file, currentYearAndMonth, "A1", defaultRowTitle[0])
				writeExcelRow(file, currentYearAndMonth, "B1", defaultRowTitle[1])
				writeExcelRow(file, currentYearAndMonth, "C1", defaultRowTitle[2])
				writeExcelRow(file, currentYearAndMonth, "D1", defaultRowTitle[3])
				writeExcelRow(file, currentYearAndMonth, "E1", defaultRowTitle[4])
				writeExcelRow(file, currentYearAndMonth, "F1", defaultRowTitle[5])
				writeExcelRow(file, currentYearAndMonth, "G1", defaultRowTitle[6])
			}

			for _, cashFlow := range cashFlowArray {
				cashFlowRowIndex++
				cashFlowIndexInString := strconv.Itoa(cashFlowRowIndex)

				writeExcelRow(file, currentYearAndMonth, "A"+cashFlowIndexInString, cashFlow.Id.Hex())
				writeExcelRow(file, currentYearAndMonth, "B"+cashFlowIndexInString, cashFlow.CategoryId.Hex())

				// Get category name with user context
				categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userId)
				categoryName := "Unknown"
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
				}
				writeExcelRow(file, currentYearAndMonth, "C"+cashFlowIndexInString, categoryName)

				writeExcelRow(file, currentYearAndMonth, "D"+cashFlowIndexInString, queryDateCurrentInString)
				writeExcelRow(file, currentYearAndMonth, "E"+cashFlowIndexInString, cashFlow.FlowType)
				writeExcelRow(file, currentYearAndMonth, "F"+cashFlowIndexInString, cashFlow.Amount)
				writeExcelRow(file, currentYearAndMonth, "G"+cashFlowIndexInString, cashFlow.Description)
			}

			queryDateCurrent = queryDateCurrent.AddDate(0, 0, 1)
		}
	}
}

func writeExcelRow(file *excelize.File, sheetName, cellPosition string, cellValue interface{}) {
	if err := file.SetCellValue(sheetName, cellPosition, cellValue); err != nil {
		util.Logger.Errorln(err)
	}
}
