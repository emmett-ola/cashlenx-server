package statistic_service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/macar-x/cashlenx-server/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ExportToPDFForUser exports the user's cash flow data to PDF file
// Only exports data belonging to the specified user
func ExportToPDFForUser(fromDateInString, toDateInString, filePath, userId string) error {
	return defaultStatisticService().ExportToPDFForUser(fromDateInString, toDateInString, filePath, userId)
}

func (s *StatisticService) ExportToPDFForUser(fromDateInString, toDateInString, filePath, userId string) error {
	if filePath == "" {
		filePath = "./export.pdf"
	}

	// Convert userId string to ObjectID
	userObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Create PDF
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape orientation for better table fit
	pdf.AddPage()

	// Set title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "CashLenX - Cash Flow Export")
	pdf.Ln(12)

	// Add export info
	pdf.SetFont("Arial", "", 10)
	if fromDateInString != "" && toDateInString != "" {
		pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", fromDateInString, toDateInString))
	} else {
		pdf.Cell(0, 6, "Period: All Time")
	}
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(200, 200, 200)

	// Column widths (total should be ~277mm for landscape A4)
	colWidths := []float64{20, 45, 20, 30, 25, 40, 60}
	headers := []string{"Date", "Category", "Type", "Amount", "ID", "Category ID", "Description"}

	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 8)
	rowCount := 0
	totalIncome := 0.0
	totalExpense := 0.0

	// Determine if we're exporting all data or a date range
	isExportAll := fromDateInString == "" && toDateInString == ""

	if isExportAll {
		// Export all data for this user using pagination
		util.Logger.Infof("Exporting all cash flow data to PDF for user %s", userId)

		limit := 100
		offset := 0

		for {
			cashFlows := s.cashFlowMapper.GetAllCashFlowsByUser(userObjectId, limit, offset)
			if len(cashFlows) == 0 {
				break
			}

			for _, cashFlow := range cashFlows {
				// Create new page if needed
				if rowCount > 0 && rowCount%35 == 0 {
					pdf.AddPage()
					// Re-add header on new page
					pdf.SetFont("Arial", "B", 9)
					pdf.SetFillColor(200, 200, 200)
					for i, header := range headers {
						pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
					}
					pdf.Ln(-1)
					pdf.SetFont("Arial", "", 8)
				}

				categoryName := "Unknown"
				categoryType := ""
				categoryEntity := s.categoryMapper.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userObjectId)
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
					categoryType = categoryEntity.Type
				}

				dateStr := util.FormatDateToStringWithDash(cashFlow.BelongsDate)
				amountStr := fmt.Sprintf("%.2f", cashFlow.Amount)

				// Truncate long text
				desc := cashFlow.Description
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				if len(categoryName) > 20 {
					categoryName = categoryName[:17] + "..."
				}

				pdf.CellFormat(colWidths[0], 6, dateStr, "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[1], 6, categoryName, "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[2], 6, categoryType, "1", 0, "C", false, 0, "")
				pdf.CellFormat(colWidths[3], 6, amountStr, "1", 0, "R", false, 0, "")
				pdf.CellFormat(colWidths[4], 6, cashFlow.Id.Hex()[:8]+"...", "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[5], 6, cashFlow.CategoryId.Hex()[:8]+"...", "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[6], 6, desc, "1", 0, "L", false, 0, "")
				pdf.Ln(-1)

				if strings.EqualFold(categoryType, "income") {
					totalIncome += cashFlow.Amount
				} else if strings.EqualFold(categoryType, "expense") {
					totalExpense += cashFlow.Amount
				}

				rowCount++
			}

			offset += limit
		}
	} else {
		// Export by date range for this user
		queryDateCurrent := util.FormatDateFromStringWithoutDash(fromDateInString)
		queryDateEnded := util.FormatDateFromStringWithoutDash(toDateInString).AddDate(0, 0, 1)

		for queryDateEnded.After(queryDateCurrent) {
			cashFlowArray := s.cashFlowMapper.GetCashFlowsByBelongsDateAndUser(queryDateCurrent, userObjectId)

			for _, cashFlow := range cashFlowArray {
				// Create new page if needed
				if rowCount > 0 && rowCount%35 == 0 {
					pdf.AddPage()
					// Re-add header on new page
					pdf.SetFont("Arial", "B", 9)
					pdf.SetFillColor(200, 200, 200)
					for i, header := range headers {
						pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
					}
					pdf.Ln(-1)
					pdf.SetFont("Arial", "", 8)
				}

				categoryName := "Unknown"
				categoryType := ""
				categoryEntity := s.categoryMapper.GetCategoryByObjectIdAndUser(cashFlow.CategoryId.Hex(), userObjectId)
				if !categoryEntity.IsEmpty() {
					categoryName = categoryEntity.Name
					categoryType = categoryEntity.Type
				}

				dateStr := util.FormatDateToStringWithDash(queryDateCurrent)
				amountStr := fmt.Sprintf("%.2f", cashFlow.Amount)

				// Truncate long text
				desc := cashFlow.Description
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				if len(categoryName) > 20 {
					categoryName = categoryName[:17] + "..."
				}

				pdf.CellFormat(colWidths[0], 6, dateStr, "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[1], 6, categoryName, "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[2], 6, categoryType, "1", 0, "C", false, 0, "")
				pdf.CellFormat(colWidths[3], 6, amountStr, "1", 0, "R", false, 0, "")
				pdf.CellFormat(colWidths[4], 6, cashFlow.Id.Hex()[:8]+"...", "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[5], 6, cashFlow.CategoryId.Hex()[:8]+"...", "1", 0, "L", false, 0, "")
				pdf.CellFormat(colWidths[6], 6, desc, "1", 0, "L", false, 0, "")
				pdf.Ln(-1)

				if strings.EqualFold(categoryType, "income") {
					totalIncome += cashFlow.Amount
				} else if strings.EqualFold(categoryType, "expense") {
					totalExpense += cashFlow.Amount
				}

				rowCount++
			}

			queryDateCurrent = queryDateCurrent.AddDate(0, 0, 1)
		}
	}

	// Add summary at the end
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Total Income:  %.2f", totalIncome))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total Expense: %.2f", totalExpense))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Balance:       %.2f", totalIncome-totalExpense))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total Records: %d", rowCount))

	// Save PDF
	err = pdf.OutputFileAndClose(filePath)
	if err != nil {
		return err
	}

	return nil
}
