#!/bin/sh

# Special thanks for https://github.com/charmbracelet/gum

function func_exec_cash_income() {
  INCOME_AMOUNT=$(gum input --placeholder "input income amount here...");
  
  if [ "$INCOME_AMOUNT" -gt 0 ]; then
    echo "Income Amount: $INCOME_AMOUNT";
    EXEC_COMMAND="$EXEC_COMMAND -a $INCOME_AMOUNT";
    
    INCOME_TYPE=$(gum input --placeholder "input income type here...");
    if [ -z "$INCOME_TYPE" ]; then
      INCOME_TYPE="unknown";
    fi
    echo "Income Type: $INCOME_TYPE";
    EXEC_COMMAND="$EXEC_COMMAND -c $INCOME_TYPE";
    
    INCOME_DATE=$(gum input --placeholder "input income date here...");
    if [ -z "$INCOME_DATE" ]; then
      INCOME_DATE=$(date +%Y%m%d);
    fi
    echo "Income Date: $INCOME_DATE";
    EXEC_COMMAND="$EXEC_COMMAND -b $INCOME_DATE";
    
    INCOME_REMARK=$(gum input --placeholder "input income remark here...");
    echo "Income Remark: $INCOME_REMARK";
    EXEC_COMMAND="$EXEC_COMMAND -d '$INCOME_REMARK'";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_cash_expense() {
  EXPENSE_AMOUNT=$(gum input --placeholder "input expense amount here...");
  
  if [ "$EXPENSE_AMOUNT" -gt 0 ]; then
    echo "Expense Amount: $EXPENSE_AMOUNT";
    EXEC_COMMAND="$EXEC_COMMAND -a $EXPENSE_AMOUNT";
    
    EXPENSE_TYPE=$(gum input --placeholder "input expense type here...");
    if [ -z "$EXPENSE_TYPE" ]; then
      EXPENSE_TYPE="unknown";
    fi
    echo "Expense Type: $EXPENSE_TYPE";
    EXEC_COMMAND="$EXEC_COMMAND -c $EXPENSE_TYPE";
    
    EXPENSE_DATE=$(gum input --placeholder "input expense date here...");
    if [ -z "$EXPENSE_DATE" ]; then
      EXPENSE_DATE=$(date +%Y%m%d);
    fi
    echo "Expense Date: $EXPENSE_DATE";
    EXEC_COMMAND="$EXEC_COMMAND -b $EXPENSE_DATE";
    
    EXPENSE_REMARK=$(gum input --placeholder "input expense remark here...");
    echo "Expense Remark: $EXPENSE_REMARK";
    # Has some problems when there is a blankspace in remark.
    EXEC_COMMAND="$EXEC_COMMAND -d '$EXPENSE_REMARK'";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_cash_query() {
  echo "Please choose one of the query type below:";
  QUERY_TYPE=$(gum choose "date");
  echo "Query Type: $QUERY_TYPE";
  
  if [ "$QUERY_TYPE" = "date" ]; then
    EXEC_COMMAND="$EXEC_COMMAND date";
    QUERY_DATE=$(gum input --placeholder "input query date here...");
    if [ -z "$QUERY_DATE" ]; then
      QUERY_DATE=$(date +%Y%m%d);
    fi
    
    echo "Query Date: $QUERY_DATE";
    EXEC_COMMAND="$EXEC_COMMAND -b $QUERY_DATE";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_cash_delete() {
  echo "Please choose one of the delete type below:";
  DELETE_TYPE=$(gum choose "id");
  echo "Delete Type: $DELETE_TYPE";
  
  if [ "$DELETE_TYPE" = "id" ]; then
    EXEC_COMMAND="$EXEC_COMMAND id";
    DELETE_ID=$(gum input --placeholder "input object id here...");
    echo "Delete Id: $DELETE_ID";
    if [ -n "$DELETE_ID" ]; then
      EXEC_COMMAND="$EXEC_COMMAND -i $DELETE_ID";
    fi
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_cash_list() {
  PAGE=$(gum input --placeholder "input page number (default: 1)...");
  if [ -z "$PAGE" ]; then
    PAGE=1;
  fi
  
  SIZE=$(gum input --placeholder "input page size (default: 10)...");
  if [ -z "$SIZE" ]; then
    SIZE=10;
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -p $PAGE -s $SIZE";
}

function func_exec_cash_range() {
  FROM_DATE=$(gum input --placeholder "input start date (YYYYMMDD)...");
  if [ -z "$FROM_DATE" ]; then
    FROM_DATE=$(date +%Y%m%d --date="-30 days");
  fi
  
  TO_DATE=$(gum input --placeholder "input end date (YYYYMMDD)...");
  if [ -z "$TO_DATE" ]; then
    TO_DATE=$(date +%Y%m%d);
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -f $FROM_DATE -t $TO_DATE";
}

function func_exec_category_create() {
  CATEGORY_NAME=$(gum input --placeholder "input category name here...");
  if [ -z "$CATEGORY_NAME" ]; then
    echo "Category name cannot be empty";
    exit -2;
  fi
  
  echo "Category Name: $CATEGORY_NAME";
  EXEC_COMMAND="$EXEC_COMMAND -n $CATEGORY_NAME";
  
  PARENT_ID=$(gum input --placeholder "input parent category id (optional)...");
  if [ -n "$PARENT_ID" ]; then
    echo "Parent ID: $PARENT_ID";
    EXEC_COMMAND="$EXEC_COMMAND -p $PARENT_ID";
  fi
  
  REMARK=$(gum input --placeholder "input remark (optional)...");
  if [ -n "$REMARK" ]; then
    echo "Remark: $REMARK";
    EXEC_COMMAND="$EXEC_COMMAND -r '$REMARK'";
  fi
}

function func_exec_category_delete() {
  CATEGORY_ID=$(gum input --placeholder "input category id here...");
  if [ -z "$CATEGORY_ID" ]; then
    echo "Category id cannot be empty";
    exit -2;
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -i $CATEGORY_ID";
}

function func_exec_category_update() {
  CATEGORY_ID=$(gum input --placeholder "input category id to update...");
  if [ -z "$CATEGORY_ID" ]; then
    echo "Category id cannot be empty";
    exit -2;
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -i $CATEGORY_ID";
  
  NEW_NAME=$(gum input --placeholder "input new category name (optional)...");
  if [ -n "$NEW_NAME" ]; then
    EXEC_COMMAND="$EXEC_COMMAND -n $NEW_NAME";
  fi
  
  NEW_PARENT=$(gum input --placeholder "input new parent id (optional)...");
  if [ -n "$NEW_PARENT" ]; then
    EXEC_COMMAND="$EXEC_COMMAND -p $NEW_PARENT";
  fi
}

function func_exec_manage_export() {
  FROM_DATE=$(gum input --placeholder "input start date (YYYYMMDD)...");
  if [ -z "$FROM_DATE" ]; then
    FROM_DATE=$(date +%Y%m%d --date="-30 days");
  fi
  
  TO_DATE=$(gum input --placeholder "input end date (YYYYMMDD)...");
  if [ -z "$TO_DATE" ]; then
    TO_DATE=$(date +%Y%m%d);
  fi
  
  FILE_PATH=$(gum input --placeholder "input output file path (default: ./export.xlsx)...");
  if [ -z "$FILE_PATH" ]; then
    FILE_PATH="./export.xlsx";
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -f $FROM_DATE -t $TO_DATE -o $FILE_PATH";
}

function func_exec_manage_import() {
  FILE_PATH=$(gum input --placeholder "input input file path...");
  if [ -z "$FILE_PATH" ]; then
    echo "File path cannot be empty";
    exit -2;
  fi
  
  EXEC_COMMAND="$EXEC_COMMAND -i $FILE_PATH";
}

function func_exec_cash_subcommand() {
  CASH_SUBTYPE=$(gum choose "income" "expense" "update" "delete" "query" "list" "range" "summary");
  echo "Cash Subtype: $CASH_SUBTYPE";
  
  if [ "$CASH_SUBTYPE" = "income" ]; then
    EXEC_COMMAND="$EXEC_COMMAND income";
    func_exec_cash_income;
  elif [ "$CASH_SUBTYPE" = "expense" ]; then
    EXEC_COMMAND="$EXEC_COMMAND expense";
    func_exec_cash_expense;
  elif [ "$CASH_SUBTYPE" = "query" ]; then
    EXEC_COMMAND="$EXEC_COMMAND query";
    func_exec_cash_query;
  elif [ "$CASH_SUBTYPE" = "delete" ]; then
    EXEC_COMMAND="$EXEC_COMMAND delete";
    func_exec_cash_delete;
  elif [ "$CASH_SUBTYPE" = "list" ]; then
    EXEC_COMMAND="$EXEC_COMMAND list";
    func_exec_cash_list;
  elif [ "$CASH_SUBTYPE" = "range" ]; then
    EXEC_COMMAND="$EXEC_COMMAND range";
    func_exec_cash_range;
  elif [ "$CASH_SUBTYPE" = "update" ] || [ "$CASH_SUBTYPE" = "summary" ]; then
    EXEC_COMMAND="$EXEC_COMMAND $CASH_SUBTYPE";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_category_subcommand() {
  CATEGORY_SUBTYPE=$(gum choose "create" "update" "delete" "query" "list");
  echo "Category Subtype: $CATEGORY_SUBTYPE";
  
  if [ "$CATEGORY_SUBTYPE" = "create" ]; then
    EXEC_COMMAND="$EXEC_COMMAND create";
    func_exec_category_create;
  elif [ "$CATEGORY_SUBTYPE" = "update" ]; then
    EXEC_COMMAND="$EXEC_COMMAND update";
    func_exec_category_update;
  elif [ "$CATEGORY_SUBTYPE" = "delete" ]; then
    EXEC_COMMAND="$EXEC_COMMAND delete";
    func_exec_category_delete;
  elif [ "$CATEGORY_SUBTYPE" = "query" ] || [ "$CATEGORY_SUBTYPE" = "list" ]; then
    EXEC_COMMAND="$EXEC_COMMAND $CATEGORY_SUBTYPE";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_manage_subcommand() {
  MANAGE_SUBTYPE=$(gum choose "export" "import" "backup" "restore" "init" "reset" "stats" "indexes");
  echo "Manage Subtype: $MANAGE_SUBTYPE";
  
  if [ "$MANAGE_SUBTYPE" = "export" ]; then
    EXEC_COMMAND="$EXEC_COMMAND export";
    func_exec_manage_export;
  elif [ "$MANAGE_SUBTYPE" = "import" ]; then
    EXEC_COMMAND="$EXEC_COMMAND import";
    func_exec_manage_import;
  elif [ "$MANAGE_SUBTYPE" = "backup" ] || [ "$MANAGE_SUBTYPE" = "restore" ] || \
       [ "$MANAGE_SUBTYPE" = "init" ] || [ "$MANAGE_SUBTYPE" = "reset" ] || \
       [ "$MANAGE_SUBTYPE" = "stats" ] || [ "$MANAGE_SUBTYPE" = "indexes" ]; then
    EXEC_COMMAND="$EXEC_COMMAND $MANAGE_SUBTYPE";
  else
    echo "Error occur...";
    exit -2;
  fi
}

function func_exec_db_subcommand() {
  DB_SUBTYPE=$(gum choose "connect" "seed");
  echo "DB Subtype: $DB_SUBTYPE";
  EXEC_COMMAND="$EXEC_COMMAND $DB_SUBTYPE";
}


# Main Function
EXEC_COMMAND="";

echo "";
echo "--- Welcome to CashLenX Interactive Mode ---";
echo "";
echo "Please choose one of the operation below:";
MAIN_TYPE=$(gum choose "cash" "category" "manage" "db");
echo "Main Operation Type: $MAIN_TYPE";

if [ "$MAIN_TYPE" = "cash" ]; then
  EXEC_COMMAND="go run main.go cash";
  func_exec_cash_subcommand;
elif [ "$MAIN_TYPE" = "category" ]; then
  EXEC_COMMAND="go run main.go category";
  func_exec_category_subcommand;
elif [ "$MAIN_TYPE" = "manage" ]; then
  EXEC_COMMAND="go run main.go manage";
  func_exec_manage_subcommand;
elif [ "$MAIN_TYPE" = "db" ]; then
  EXEC_COMMAND="go run main.go db";
  func_exec_db_subcommand;
else
  echo "Error occur...";
  exit -1;
fi

if [ -z "$EXEC_COMMAND" ]; then
  # No need to execute command, just exit in normally.
  exit 0;
fi

gum confirm "Execute Command: $EXEC_COMMAND" && gum spin --spinner dot --title "Executing..." -- sleep 0.5 && $EXEC_COMMAND;
