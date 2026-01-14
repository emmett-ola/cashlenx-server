#!/usr/bin/env pwsh

<#
.SYNOPSIS
CashLenX Interactive Mode

.DESCRIPTION
Interactive command-line interface for CashLenX, allowing you to manage cash flows, categories, and application data.

.REQUIREMENTS
- PowerShell 7.0+
- gum (https://github.com/charmbracelet/gum)

.EXAMPLE
./interactive.ps1

#>

function Func-Exec-Cash-Income {
    $incomeAmount = gum input --placeholder "input income amount here..."
    
    try {
        $incomeAmount = [decimal]$incomeAmount
        if ($incomeAmount -gt 0) {
            Write-Host "Income Amount: $incomeAmount"
            $script:ExecCommand = "$script:ExecCommand -a $incomeAmount"
            
            $incomeType = gum input --placeholder "input income type here..."
            if ([string]::IsNullOrEmpty($incomeType)) {
                $incomeType = "unknown"
            }
            Write-Host "Income Type: $incomeType"
            $script:ExecCommand = "$script:ExecCommand -c $incomeType"
            
            $incomeDate = gum input --placeholder "input income date here..."
            if ([string]::IsNullOrEmpty($incomeDate)) {
                $incomeDate = Get-Date -Format "yyyyMMdd"
            }
            Write-Host "Income Date: $incomeDate"
            $script:ExecCommand = "$script:ExecCommand -b $incomeDate"
            
            $incomeRemark = gum input --placeholder "input income remark here..."
            Write-Host "Income Remark: $incomeRemark"
            $script:ExecCommand = "$script:ExecCommand -d '$incomeRemark'"
        } else {
            Write-Host "Error: Amount must be greater than 0" -ForegroundColor Red
            exit -2
        }
    } catch {
        Write-Host "Error: Invalid amount format" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Cash-Expense {
    $expenseAmount = gum input --placeholder "input expense amount here..."
    
    try {
        $expenseAmount = [decimal]$expenseAmount
        if ($expenseAmount -gt 0) {
            Write-Host "Expense Amount: $expenseAmount"
            $script:ExecCommand = "$script:ExecCommand -a $expenseAmount"
            
            $expenseType = gum input --placeholder "input expense type here..."
            if ([string]::IsNullOrEmpty($expenseType)) {
                $expenseType = "unknown"
            }
            Write-Host "Expense Type: $expenseType"
            $script:ExecCommand = "$script:ExecCommand -c $expenseType"
            
            $expenseDate = gum input --placeholder "input expense date here..."
            if ([string]::IsNullOrEmpty($expenseDate)) {
                $expenseDate = Get-Date -Format "yyyyMMdd"
            }
            Write-Host "Expense Date: $expenseDate"
            $script:ExecCommand = "$script:ExecCommand -b $expenseDate"
            
            $expenseRemark = gum input --placeholder "input expense remark here..."
            Write-Host "Expense Remark: $expenseRemark"
            $script:ExecCommand = "$script:ExecCommand -d '$expenseRemark'"
        } else {
            Write-Host "Error: Amount must be greater than 0" -ForegroundColor Red
            exit -2
        }
    } catch {
        Write-Host "Error: Invalid amount format" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Cash-Query {
    Write-Host "Please choose one of the query type below:"
    $queryType = gum choose "date"
    Write-Host "Query Type: $queryType"
    
    if ($queryType -eq "date") {
        $script:ExecCommand = "$script:ExecCommand date"
        $queryDate = gum input --placeholder "input query date here..."
        if ([string]::IsNullOrEmpty($queryDate)) {
            $queryDate = Get-Date -Format "yyyyMMdd"
        }
        
        Write-Host "Query Date: $queryDate"
        $script:ExecCommand = "$script:ExecCommand -b $queryDate"
    } else {
        Write-Host "Error: Invalid query type" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Cash-Delete {
    Write-Host "Please choose one of the delete type below:"
    $deleteType = gum choose "id"
    Write-Host "Delete Type: $deleteType"
    
    if ($deleteType -eq "id") {
        $script:ExecCommand = "$script:ExecCommand id"
        $deleteId = gum input --placeholder "input object id here..."
        Write-Host "Delete Id: $deleteId"
        if (-not [string]::IsNullOrEmpty($deleteId)) {
            $script:ExecCommand = "$script:ExecCommand -i $deleteId"
        }
    } else {
        Write-Host "Error: Invalid delete type" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Cash-List {
    $page = gum input --placeholder "input page number (default: 1)..."
    if ([string]::IsNullOrEmpty($page)) {
        $page = 1
    }
    
    $size = gum input --placeholder "input page size (default: 10)..."
    if ([string]::IsNullOrEmpty($size)) {
        $size = 10
    }
    
    $script:ExecCommand = "$script:ExecCommand -p $page -s $size"
}

function Func-Exec-Cash-Range {
    $fromDate = gum input --placeholder "input start date (YYYYMMDD)..."
    if ([string]::IsNullOrEmpty($fromDate)) {
        $fromDate = (Get-Date).AddDays(-30).ToString("yyyyMMdd")
    }
    
    $toDate = gum input --placeholder "input end date (YYYYMMDD)..."
    if ([string]::IsNullOrEmpty($toDate)) {
        $toDate = (Get-Date).ToString("yyyyMMdd")
    }
    
    $script:ExecCommand = "$script:ExecCommand -f $fromDate -t $toDate"
}

function Func-Exec-Category-Create {
    $categoryName = gum input --placeholder "input category name here..."
    if ([string]::IsNullOrEmpty($categoryName)) {
        Write-Host "Category name cannot be empty" -ForegroundColor Red
        exit -2
    }
    
    Write-Host "Category Name: $categoryName"
    $script:ExecCommand = "$script:ExecCommand -n $categoryName"
    
    $parentId = gum input --placeholder "input parent category id (optional)..."
    if (-not [string]::IsNullOrEmpty($parentId)) {
        Write-Host "Parent ID: $parentId"
        $script:ExecCommand = "$script:ExecCommand -p $parentId"
    }
    
    $remark = gum input --placeholder "input remark (optional)..."
    if (-not [string]::IsNullOrEmpty($remark)) {
        Write-Host "Remark: $remark"
        $script:ExecCommand = "$script:ExecCommand -r '$remark'"
    }
}

function Func-Exec-Category-Delete {
    $categoryId = gum input --placeholder "input category id here..."
    if ([string]::IsNullOrEmpty($categoryId)) {
        Write-Host "Category id cannot be empty" -ForegroundColor Red
        exit -2
    }
    
    $script:ExecCommand = "$script:ExecCommand -i $categoryId"
}

function Func-Exec-Category-Update {
    $categoryId = gum input --placeholder "input category id to update..."
    if ([string]::IsNullOrEmpty($categoryId)) {
        Write-Host "Category id cannot be empty" -ForegroundColor Red
        exit -2
    }
    
    $script:ExecCommand = "$script:ExecCommand -i $categoryId"
    
    $newName = gum input --placeholder "input new category name (optional)..."
    if (-not [string]::IsNullOrEmpty($newName)) {
        $script:ExecCommand = "$script:ExecCommand -n $newName"
    }
    
    $newParent = gum input --placeholder "input new parent id (optional)..."
    if (-not [string]::IsNullOrEmpty($newParent)) {
        $script:ExecCommand = "$script:ExecCommand -p $newParent"
    }
}

function Func-Exec-Manage-Export {
    $fromDate = gum input --placeholder "input start date (YYYYMMDD)..."
    if ([string]::IsNullOrEmpty($fromDate)) {
        $fromDate = (Get-Date).AddDays(-30).ToString("yyyyMMdd")
    }
    
    $toDate = gum input --placeholder "input end date (YYYYMMDD)..."
    if ([string]::IsNullOrEmpty($toDate)) {
        $toDate = (Get-Date).ToString("yyyyMMdd")
    }
    
    $filePath = gum input --placeholder "input output file path (default: ./export.xlsx)..."
    if ([string]::IsNullOrEmpty($filePath)) {
        $filePath = "./export.xlsx"
    }
    
    $script:ExecCommand = "$script:ExecCommand -f $fromDate -t $toDate -o $filePath"
}

function Func-Exec-Manage-Import {
    $filePath = gum input --placeholder "input input file path..."
    if ([string]::IsNullOrEmpty($filePath)) {
        Write-Host "File path cannot be empty" -ForegroundColor Red
        exit -2
    }
    
    $script:ExecCommand = "$script:ExecCommand -i $filePath"
}

function Func-Exec-Cash-Subcommand {
    $cashSubtype = gum choose "income" "expense" "update" "delete" "query" "list" "range" "summary"
    Write-Host "Cash Subtype: $cashSubtype"
    
    if ($cashSubtype -eq "income") {
        $script:ExecCommand = "$script:ExecCommand income"
        Func-Exec-Cash-Income
    } elseif ($cashSubtype -eq "expense") {
        $script:ExecCommand = "$script:ExecCommand expense"
        Func-Exec-Cash-Expense
    } elseif ($cashSubtype -eq "query") {
        $script:ExecCommand = "$script:ExecCommand query"
        Func-Exec-Cash-Query
    } elseif ($cashSubtype -eq "delete") {
        $script:ExecCommand = "$script:ExecCommand delete"
        Func-Exec-Cash-Delete
    } elseif ($cashSubtype -eq "list") {
        $script:ExecCommand = "$script:ExecCommand list"
        Func-Exec-Cash-List
    } elseif ($cashSubtype -eq "range") {
        $script:ExecCommand = "$script:ExecCommand range"
        Func-Exec-Cash-Range
    } elseif ($cashSubtype -eq "update" -or $cashSubtype -eq "summary") {
        $script:ExecCommand = "$script:ExecCommand $cashSubtype"
    } else {
        Write-Host "Error: Invalid cash subtype" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Category-Subcommand {
    $categorySubtype = gum choose "create" "update" "delete" "query" "list"
    Write-Host "Category Subtype: $categorySubtype"
    
    if ($categorySubtype -eq "create") {
        $script:ExecCommand = "$script:ExecCommand create"
        Func-Exec-Category-Create
    } elseif ($categorySubtype -eq "update") {
        $script:ExecCommand = "$script:ExecCommand update"
        Func-Exec-Category-Update
    } elseif ($categorySubtype -eq "delete") {
        $script:ExecCommand = "$script:ExecCommand delete"
        Func-Exec-Category-Delete
    } elseif ($categorySubtype -eq "query" -or $categorySubtype -eq "list") {
        $script:ExecCommand = "$script:ExecCommand $categorySubtype"
    } else {
        Write-Host "Error: Invalid category subtype" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Manage-Subcommand {
    $manageSubtype = gum choose "export" "import" "backup" "restore" "init" "reset" "stats" "indexes"
    Write-Host "Manage Subtype: $manageSubtype"
    
    if ($manageSubtype -eq "export") {
        $script:ExecCommand = "$script:ExecCommand export"
        Func-Exec-Manage-Export
    } elseif ($manageSubtype -eq "import") {
        $script:ExecCommand = "$script:ExecCommand import"
        Func-Exec-Manage-Import
    } elseif ($manageSubtype -eq "backup" -or $manageSubtype -eq "restore" -or 
              $manageSubtype -eq "init" -or $manageSubtype -eq "reset" -or 
              $manageSubtype -eq "stats" -or $manageSubtype -eq "indexes") {
        $script:ExecCommand = "$script:ExecCommand $manageSubtype"
    } else {
        Write-Host "Error: Invalid manage subtype" -ForegroundColor Red
        exit -2
    }
}

function Func-Exec-Db-Subcommand {
    $dbSubtype = gum choose "connect" "seed"
    Write-Host "DB Subtype: $dbSubtype"
    $script:ExecCommand = "$script:ExecCommand $dbSubtype"
}

# Main Script
$ExecCommand = ""

Write-Host ""
Write-Host "--- Welcome to CashLenX Interactive Mode ---"
Write-Host ""
Write-Host "Please choose one of the operation below:"
$mainType = gum choose "cash" "category" "manage" "db"
Write-Host "Main Operation Type: $mainType"

if ($mainType -eq "cash") {
    $ExecCommand = "go run main.go cash"
    Func-Exec-Cash-Subcommand
} elseif ($mainType -eq "category") {
    $ExecCommand = "go run main.go category"
    Func-Exec-Category-Subcommand
} elseif ($mainType -eq "manage") {
    $ExecCommand = "go run main.go manage"
    Func-Exec-Manage-Subcommand
} elseif ($mainType -eq "db") {
    $ExecCommand = "go run main.go db"
    Func-Exec-Db-Subcommand
} else {
    Write-Host "Error: Invalid operation type" -ForegroundColor Red
    exit -1
}

if (-not [string]::IsNullOrEmpty($ExecCommand)) {
    if (gum confirm "Execute Command: $ExecCommand") {
        gum spin --spinner dot --title "Executing..." -- sleep 0.5
        Invoke-Expression $ExecCommand
    }
}
