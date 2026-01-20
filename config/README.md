# Configuration Files

This directory contains configuration files for the CashLenX application.

## default_categories.json

This file defines the default categories that are automatically created for each new user when they register.

### Structure

```json
{
  "categories": [
    {
      "name": "Category Name",
      "type": "INCOME or EXPENSE",
      "remark": "Description of the category"
    }
  ]
}
```

### Fields

- **name** (string, required): The display name of the category
- **type** (string, required): Either "INCOME" or "EXPENSE"
- **remark** (string, optional): A description or note about the category

### Behavior

- When a new user is created (via registration or admin creation), the application reads this file
- Each category in the list is automatically created for the new user with `belongs_user_id` set to their user ID
- Users can then manage, edit, or delete these categories as needed
- If the file cannot be read, the application falls back to hardcoded default categories

### Customization

To customize default categories for your deployment:

1. Edit `config/default_categories.json` directly, or
2. Create a custom file and set the path via environment variable:
   ```bash
   DEFAULT_CATEGORIES_PATH=/path/to/your/categories.json
   ```

### Example Categories Provided

The default configuration includes:

**Income Categories:**
- Salary
- Freelance
- Investment
- Other Income

**Expense Categories:**
- Food & Dining
- Transportation
- Shopping
- Entertainment
- Healthcare
- Utilities

### Notes

- Categories are user-specific - each user gets their own copy
- Changes to this file only affect newly created users
- Existing users retain their current categories
- Users can add, edit, or delete categories independently
