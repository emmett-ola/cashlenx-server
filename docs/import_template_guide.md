# Import Template Guide

This document provides a guide for using the import template to bulk import cash flow data into the system.

## Template Structure

### Excel File Format
The import template should be an Excel file (.xlsx) with the following structure:

### Worksheet Requirements
- Each worksheet should represent a year-month group (e.g., "202501" for January 2025)
- The first row must contain the exact column headers as specified below
- Data rows follow the header row

### Required Column Headers
| Column Name | Description | Format | Required |
|-------------|-------------|--------|----------|
| Id | Unique identifier (will be generated if empty) | MongoDB ObjectID format | No |
| CategoryId | Category ID | MongoDB ObjectID format | No |
| CategoryName | Category name | Text (max 50 chars) | Yes (if CategoryId not provided) |
| BelongsDate | Transaction date | YYYYMMDD (e.g., 20250115) | Yes |
| FlowType | Transaction type | "income" or "expense" | Yes |
| Amount | Transaction amount | Numeric (e.g., 100.50) | Yes |
| Description | Transaction description | Text (max 200 chars) | No |

## Template Example

### Worksheet Name: 202501
| Id | CategoryId | CategoryName | BelongsDate | FlowType | Amount | Description |
|----|------------|--------------|-------------|----------|--------|-------------|
| | | Salary | 20250115 | income | 5000.00 | January salary |
| | | Rent | 20250101 | expense | 1200.00 | Monthly rent payment |
| | | Groceries | 20250110 | expense | 150.75 | Weekly grocery shopping |
| | | Utilities | 20250120 | expense | 85.20 | Electricity bill |
| | | Bonus | 20250125 | income | 1000.00 | Performance bonus |

## Import Process

### Using the API
1. Prepare your Excel file according to the template structure
2. Send a POST request to `/api/manage/import` with the file attached as `form-data` with key `file`
3. The system will validate and process each row
4. You will receive a response with import results

### Using the CLI
```bash
# Import data from Excel file
cashlenx import --file path/to/your/file.xlsx
```

## Validation Rules

1. **Required Fields**: BelongsDate, FlowType, and Amount must be provided
2. **Date Format**: BelongsDate must be in YYYYMMDD format
3. **FlowType**: Must be either "income" or "expense"
4. **Amount**: Must be a valid numeric value
5. **Category**: Either CategoryId or CategoryName must be provided
   - If CategoryName is provided and doesn't exist, a new category will be created
6. **Duplicate Check**: If Id is provided and exists in the system, the row will be ignored

## Import Results

After importing, you will receive a summary of the import process:

- **Succeed Rows**: Number of rows successfully imported
- **Ignored Rows**: Number of rows ignored (e.g., duplicate IDs)
- **Failed Rows**: Number of rows that failed validation

## Best Practices

1. **Test with Small Batch**: Start with a small number of rows to verify the template works correctly
2. **Use Consistent Categories**: Maintain consistent category names to avoid creating duplicate categories
3. **Backup Before Importing**: Create a database backup before importing large datasets
4. **Validate Data**: Ensure all required fields are filled correctly before importing
5. **Check Results**: Review the import results and logs to identify any issues

## Troubleshooting

### Common Issues

1. **"File too large" error**: Ensure your file is under 10MB
2. **"Category not satisfied" error**: Provide either a valid CategoryId or a CategoryName
3. **"Required field not satisfied" error**: Check that all required fields (BelongsDate, FlowType, Amount) are filled
4. **"cash_flow existed, ignored import"**: The Id provided already exists in the system
5. **"sheet title un-expected, parse failed"**: Ensure the column headers match exactly as specified

### Logs

Detailed import logs are available in the server logs, including:
- Sheet processing information
- Row-by-row import results
- Error messages for failed rows

## Export and Import Workflow

1. **Export Existing Data**: Use the export functionality to get a copy of your current data
2. **Modify in Excel**: Edit the exported file or create a new file using the same template
3. **Import Back**: Use the import functionality to add or update data

## Notes

- Excel files with multiple worksheets are supported
- Each worksheet should contain data for a single year-month group
- The system will automatically create categories if they don't exist
- Large datasets are processed efficiently using batch operations
- Imported data will be merged with existing data, not replace it

## Examples

### Example 1: Basic Income Transaction
| BelongsDate | FlowType | Amount | CategoryName |
|-------------|----------|--------|--------------|
| 20250201 | income | 3500.00 | Freelance Work |

### Example 2: Expense with Description
| BelongsDate | FlowType | Amount | CategoryName | Description |
|-------------|----------|--------|--------------|-------------|
| 20250205 | expense | 45.99 | Dining | Dinner with friends |

### Example 3: Using CategoryId
| BelongsDate | FlowType | Amount | CategoryId | Description |
|-------------|----------|--------|------------|-------------|
| 20250210 | expense | 20.00 | 60d5ec4a5a8b8c4a3d2e1f0a | Office supplies |