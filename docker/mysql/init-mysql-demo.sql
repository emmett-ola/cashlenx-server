-- MySQL Demo Data for CashLenX - FOR TESTING ONLY
-- This script inserts sample transactions for development and testing
-- 
-- To use this data:
-- Option 1 (Manual import to database):
--   mysql -u cashlenx -p cashlenx123 cashlenx < init-mysql-demo.sql
--
-- Option 2 (Via CLI - Recommended):
--   1. Create Excel file from this data
--   2. Run: cashlenx manage import -i demo-data.xlsx
--
-- This keeps demo data separate from production initialization

USE cashlenx;

-- Get category IDs (assuming they exist from schema initialization)
-- We'll use hardcoded inserts but should match categories from init-mysql-schema.sql

SET @salary_id = (SELECT id FROM categories WHERE name = 'Salary' LIMIT 1);
SET @freelance_id = (SELECT id FROM categories WHERE name = 'Freelance' LIMIT 1);
SET @investment_id = (SELECT id FROM categories WHERE name = 'Investment' LIMIT 1);
SET @food_id = (SELECT id FROM categories WHERE name = 'Food & Dining' LIMIT 1);
SET @transport_id = (SELECT id FROM categories WHERE name = 'Transportation' LIMIT 1);
SET @shopping_id = (SELECT id FROM categories WHERE name = 'Shopping' LIMIT 1);
SET @entertainment_id = (SELECT id FROM categories WHERE name = 'Entertainment' LIMIT 1);
SET @healthcare_id = (SELECT id FROM categories WHERE name = 'Healthcare' LIMIT 1);
SET @utilities_id = (SELECT id FROM categories WHERE name = 'Utilities' LIMIT 1);

-- Demo Transactions: This Week
-- Today
INSERT INTO cash_flows (id, category_id, belongs_date, flow_type, amount, description, remark) VALUES
(UUID(), @food_id, CURDATE(), 'expense', 45.50, 'Lunch at Italian restaurant', 'Great pasta'),
(UUID(), @transport_id, CURDATE(), 'expense', 12.00, 'Uber to office', 'Morning commute'),
(UUID(), @salary_id, CURDATE(), 'income', 3500.00, 'Monthly salary', 'Regular income');

-- Yesterday
INSERT INTO cash_flows (id, category_id, belongs_date, flow_type, amount, description, remark) VALUES
(UUID(), @shopping_id, DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'expense', 89.99, 'New shoes', 'Sports shoes'),
(UUID(), @entertainment_id, DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'expense', 25.00, 'Movie tickets', 'Watched latest film');

-- This week (within last 7 days)
INSERT INTO cash_flows (id, category_id, belongs_date, flow_type, amount, description, remark) VALUES
(UUID(), @utilities_id, DATE_SUB(CURDATE(), INTERVAL 3 DAY), 'expense', 150.00, 'Electricity bill', 'Monthly bill'),
(UUID(), @healthcare_id, DATE_SUB(CURDATE(), INTERVAL 4 DAY), 'expense', 65.00, 'Pharmacy', 'Prescription refill'),
(UUID(), @investment_id, DATE_SUB(CURDATE(), INTERVAL 5 DAY), 'income', 200.00, 'Dividend payment', 'Stock dividend');

-- Earlier this month (8-30 days ago)
INSERT INTO cash_flows (id, category_id, belongs_date, flow_type, amount, description, remark) VALUES
(UUID(), @food_id, DATE_SUB(CURDATE(), INTERVAL 10 DAY), 'expense', 120.00, 'Grocery shopping', 'Weekly groceries'),
(UUID(), @transport_id, DATE_SUB(CURDATE(), INTERVAL 12 DAY), 'expense', 45.00, 'Gas station', 'Fill up tank'),
(UUID(), @food_id, DATE_SUB(CURDATE(), INTERVAL 15 DAY), 'expense', 75.50, 'Dinner with friends', 'Birthday celebration'),
(UUID(), @entertainment_id, DATE_SUB(CURDATE(), INTERVAL 18 DAY), 'expense', 30.00, 'Concert tickets', 'Live music event'),
(UUID(), @shopping_id, DATE_SUB(CURDATE(), INTERVAL 20 DAY), 'expense', 250.00, 'Clothing', 'New wardrobe'),
(UUID(), @freelance_id, DATE_SUB(CURDATE(), INTERVAL 22 DAY), 'income', 500.00, 'Freelance project', 'Web design project'),
(UUID(), @utilities_id, DATE_SUB(CURDATE(), INTERVAL 25 DAY), 'expense', 80.00, 'Internet bill', 'Monthly service');

-- Print summary
SELECT 
    'Demo data loaded successfully!' AS message,
    (SELECT COUNT(*) FROM cash_flows WHERE flow_type = 'income') AS total_income_transactions,
    (SELECT COUNT(*) FROM cash_flows WHERE flow_type = 'expense') AS total_expense_transactions,
    (SELECT COUNT(*) FROM cash_flows) AS total_transactions,
    (SELECT SUM(amount) FROM cash_flows WHERE flow_type = 'income') AS total_income_amount,
    (SELECT SUM(amount) FROM cash_flows WHERE flow_type = 'expense') AS total_expense_amount,
    (SELECT SUM(CASE WHEN flow_type = 'income' THEN amount ELSE -amount END) FROM cash_flows) AS balance;
