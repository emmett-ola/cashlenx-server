package statistic_service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/macar-x/cashlenx-server/mapper/cash_flow_mapper"
	"github.com/macar-x/cashlenx-server/mapper/category_mapper"
	"github.com/macar-x/cashlenx-server/model"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	sheetRowNumberLabel        = "row_num"
	requiredRowFieldList       = []string{"BelongsDate", "FlowType", "Amount"}
	importFailedRowNumberList  []int
	importIgnoredRowNumberList []int
	importSucceedRowNumberList []int
)

// ImportForUser imports cash flow data from Excel to the user's account
// All imported records will be associated with the specified user
func ImportForUser(filePath, userId string) error {
	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Open and read the target file
	file := readExcelFile(filePath)
	if file == nil {
		return errors.New("can not read data from file")
	}
	// Remember to close the file after execution
	defer func() {
		if err := file.Close(); err != nil {
			util.Logger.Error(err.Error())
		}
	}()

	// Get the list of worksheets and iterate through them to read data
	sheetNameList := file.GetSheetList()
	for _, currentSheetName := range sheetNameList {

		importSucceedRowNumberList = []int{}
		importIgnoredRowNumberList = []int{}
		importFailedRowNumberList = []int{}

		// Report sheet is not a data sheet, skip it
		if currentSheetName == defaultSheetName {
			continue
		}

		rows, err := file.Rows(currentSheetName)
		if err != nil {
			util.Logger.Errorw("read sheet rows failed", "error", err)
		}
		util.Logger.Infof("processing sheet %s", currentSheetName)
		cashFlowMapByDate := readSheetDataForUser(rows, userObjectId)
		for date, cashFlowMapByColumnList := range cashFlowMapByDate {
			saveIntoDBForUser(cashFlowMapByColumnList, userObjectId)
			util.Logger.Debugf("%s of %s's flows imported for user %s", util.FormatDateToStringWithoutDash(date), currentSheetName, userObjectId.Hex())
		}
		util.Logger.Infow("sheet has been imported",
			"sheet_name", currentSheetName,
			"user_id", userObjectId.Hex(),
			"succeed_row", importSucceedRowNumberList,
			"ignored_row", importIgnoredRowNumberList,
			"failed_row", importFailedRowNumberList)
	}
	return nil
}

func readExcelFile(fileName string) *excelize.File {
	file, err := excelize.OpenFile(fileName)
	if err != nil {
		util.Logger.Error(err.Error())
	}
	return file
}

/**
 * Read worksheet data and organize cashFlows with date as key
 * Only processes data for the specified user
 */
func readSheetDataForUser(sheetRowCursor *excelize.Rows, userId primitive.ObjectID) map[time.Time][]map[string]string {
	cashFlowMapByDate := make(map[time.Time][]map[string]string)

	// First row is title row, verify if the format is correct
	sheetRowCursor.Next()
	currentRowNumber := 1
	rowColumnList, err := sheetRowCursor.Columns()
	if err != nil {
		util.Logger.Error(err.Error())
	}
	if !isSheetTitleVerified(rowColumnList) {
		return cashFlowMapByDate
	}

	// Iterate through each row of data to assemble CashFlow
	for sheetRowCursor.Next() {
		// Update current row number
		currentRowNumber++

		rowColumnList, err = sheetRowCursor.Columns()
		if err != nil {
			util.Logger.Error(err.Error())
		}

		// Assemble each row of data in order to form a title-value Map
		cashFlowMapByColumn := map[string]string{}
		for index, colCell := range rowColumnList {
			cashFlowMapByColumn[defaultRowTitle[index]] = colCell
		}
		cashFlowMapByColumn[sheetRowNumberLabel] = strconv.Itoa(currentRowNumber)

		// Check category info and get the correct id for this user
		newCategoryId := handleCategoryInfoForUser(
			cashFlowMapByColumn["CategoryId"], cashFlowMapByColumn["CategoryName"], userId)
		if newCategoryId == "" {
			fmt.Println("failed: row " + strconv.Itoa(currentRowNumber) + ": category not satisfied")
			importFailedRowNumberList = append(importFailedRowNumberList, currentRowNumber)
			continue
		}
		cashFlowMapByColumn["CategoryId"] = newCategoryId

		// Required field validation
		if !isRequiredFieldSatisfied(currentRowNumber, cashFlowMapByColumn) {
			fmt.Println("failed: row " + strconv.Itoa(currentRowNumber) + ": required field not satisfied")
			importFailedRowNumberList = append(importFailedRowNumberList, currentRowNumber)
			continue
		}

		cashFlowDate := util.FormatDateFromStringWithoutDash(cashFlowMapByColumn["BelongsDate"])
		cashFlowMapByDate[cashFlowDate] = append(cashFlowMapByDate[cashFlowDate], cashFlowMapByColumn)
	}

	if err = sheetRowCursor.Close(); err != nil {
		util.Logger.Error(err.Error())
	}

	return cashFlowMapByDate
}

func isSheetTitleVerified(titleColumnList []string) bool {
	for index, colCell := range titleColumnList {
		if colCell != defaultRowTitle[index] {
			util.Logger.Warn("sheet title un-expected, parse failed.")
			return false
		}
	}
	return true
}

func isRequiredFieldSatisfied(currentRowNumber int, columnCellMap map[string]string) bool {
	for _, requiredRowField := range requiredRowFieldList {
		if columnCellMap[requiredRowField] == "" {
			util.Logger.Errorw("field could not be empty, import failed",
				sheetRowNumberLabel, currentRowNumber, "field", requiredRowField)
			return false
		}
	}
	return true
}

// handleCategoryInfoForUser handles category lookup/creation with user context
// Ensures categories are only matched within the user's own categories
func handleCategoryInfoForUser(categoryId, categoryName string, userId primitive.ObjectID) string {
	// Use category id to fetch first (only within user's categories)
	if categoryId != "" {
		categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectIdAndUser(categoryId, userId)
		if !categoryEntity.IsEmpty() {
			return categoryEntity.Id.Hex()
		}
		util.Logger.Warnw("category not existed for user", "category_id", categoryId, "user_id", userId.Hex())
	}

	// If category name is empty, fail it.
	if categoryName == "" {
		return ""
	}

	// Use category name to fetch correct id (only within user's categories)
	categoryEntity := category_mapper.INSTANCE.GetCategoryByNameAndUser(categoryName, userId)
	if !categoryEntity.IsEmpty() {
		return categoryEntity.Id.Hex()
	}
	util.Logger.Warnw("category not existed for user", "category_name", categoryName, "user_id", userId.Hex())

	// Create new category for this user
	plainId := category_mapper.INSTANCE.InsertCategoryByEntity(model.CategoryEntity{
		UserId: userId,
		Name:   categoryName,
		Remark: "created by import",
	})
	util.Logger.Infow("created new category for user", "category_name", categoryName, "user_id", userId.Hex())
	return plainId
}

// saveIntoDBForUser saves cash flows with user association
// Ensures all cash flows are associated with the specified user
func saveIntoDBForUser(cashFlowMapByColumnList []map[string]string, userId primitive.ObjectID) {
	for _, cashFlowMapByColumn := range cashFlowMapByColumnList {
		cashFlowEntity := model.CashFlowEntity{}.Build(cashFlowMapByColumn)

		// Set the userId for this cash flow
		cashFlowEntity.UserId = userId

		if cashFlowEntity.Id != primitive.NilObjectID {
			// Check if cash flow already exists for this user
			existedCashFlow := cash_flow_mapper.INSTANCE.GetCashFlowByObjectIdAndUser(cashFlowEntity.Id.Hex(), userId)
			if !existedCashFlow.IsEmpty() {
				util.Logger.Warnw("cash_flow existed for user, ignored import.",
					sheetRowNumberLabel, cashFlowMapByColumn[sheetRowNumberLabel],
					"objectId", cashFlowEntity.Id.Hex(),
					"user_id", userId.Hex())
				fmt.Println("ignored: row " + cashFlowMapByColumn[sheetRowNumberLabel] + ": cash_flow existed")
				importIgnoredRowNumberList = append(importIgnoredRowNumberList,
					util.ToInteger(cashFlowMapByColumn[sheetRowNumberLabel]))
				continue
			}
		}
		newPlainId := cash_flow_mapper.INSTANCE.InsertCashFlowByEntity(cashFlowEntity)
		cashFlowEntity.Id = util.Convert2ObjectId(newPlainId)
		util.Logger.Debugw("cash_flow inserted for user",
			"cash_flow", cashFlowEntity.ToString(),
			"user_id", userId.Hex())
		fmt.Println("succeed: row " + cashFlowMapByColumn[sheetRowNumberLabel] + ": cash_flow saved")
		importSucceedRowNumberList = append(importSucceedRowNumberList,
			util.ToInteger(cashFlowMapByColumn[sheetRowNumberLabel]))
	}
}
