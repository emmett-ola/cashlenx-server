package manage_service

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

func ImportService(filePath string) error {
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
		cashFlowMapByDate := readSheetData(rows)
		for date, cashFlowMapByColumnList := range cashFlowMapByDate {
			saveIntoDB(cashFlowMapByColumnList)
			util.Logger.Debugf("%s of %s's flows imported", util.FormatDateToStringWithoutDash(date), currentSheetName)
		}
		util.Logger.Infow("sheet has been imported",
			"sheet_name", currentSheetName,
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
 */
func readSheetData(sheetRowCursor *excelize.Rows) map[time.Time][]map[string]string {
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
		// check category info and get the correct id
		newCategoryId := handleCategoryInfo(
			cashFlowMapByColumn["CategoryId"], cashFlowMapByColumn["CategoryName"])
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

func handleCategoryInfo(categoryId, categoryName string) string {
	// use category id to fetch first
	if categoryId != "" {
		categoryEntity := category_mapper.INSTANCE.GetCategoryByObjectId(categoryId)
		if !categoryEntity.IsEmpty() {
			return categoryEntity.Id.Hex()
		}
		util.Logger.Warnw("category not existed", "category_id", categoryId)
	}

	// if category name is empty, fail it.
	if categoryName == "" {
		return ""
	}

	// use category name to fetch correct id
	categoryEntity := category_mapper.INSTANCE.GetCategoryByName(categoryName)
	if !categoryEntity.IsEmpty() {
		return categoryEntity.Id.Hex()
	}
	util.Logger.Warnw("category not existed", "category_name", categoryName)

	// create new category for this flow
	plainId := category_mapper.INSTANCE.InsertCategoryByEntity(model.CategoryEntity{
		Name:   categoryName,
		Remark: "create by import",
	})
	return plainId
}

func saveIntoDB(cashFlowMapByColumnList []map[string]string) {
	for _, cashFlowMapByColumn := range cashFlowMapByColumnList {
		cashFlowEntity := model.CashFlowEntity{}.Build(cashFlowMapByColumn)
		if cashFlowEntity.Id != primitive.NilObjectID {
			existedCashFlow := cash_flow_mapper.INSTANCE.GetCashFlowByObjectId(cashFlowEntity.Id.Hex())
			if !existedCashFlow.IsEmpty() {
				util.Logger.Warnw("cash_flow existed, ignored import.",
					sheetRowNumberLabel, cashFlowMapByColumn[sheetRowNumberLabel],
					"objectId", cashFlowEntity.Id.Hex())
				fmt.Println("ignored: row " + cashFlowMapByColumn[sheetRowNumberLabel] + ": cash_flow existed")
				importIgnoredRowNumberList = append(importIgnoredRowNumberList,
					util.ToInteger(cashFlowMapByColumn[sheetRowNumberLabel]))
				continue
			}
		}
		newPlainId := cash_flow_mapper.INSTANCE.InsertCashFlowByEntity(cashFlowEntity)
		cashFlowEntity.Id = util.Convert2ObjectId(newPlainId)
		util.Logger.Debug("cash_flow inserted: " + cashFlowEntity.ToString())
		fmt.Println("succeed: row " + cashFlowMapByColumn[sheetRowNumberLabel] + ": cash_flow saved")
		importSucceedRowNumberList = append(importSucceedRowNumberList,
			util.ToInteger(cashFlowMapByColumn[sheetRowNumberLabel]))
	}
}
