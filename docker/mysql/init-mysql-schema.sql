-- MySQL initialization script for CashLenX - SCHEMA ONLY
-- This script creates tables with basic default categories
-- Demo/test data is available in init-mysql-demo.sql (import manually via CLI: cashlenx manage import)

USE cashlenx;

-- Create categories table
-- Aligned with Go CategoryEntity model:
-- - id: UUID primary key
-- - parent_id: for hierarchical categories (nullable)
-- - name: category name
-- - remark: additional notes
-- - create_time: creation timestamp
-- - modify_time: last modification timestamp
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY COMMENT 'UUID identifier',
    parent_id VARCHAR(36) COMMENT 'Parent category ID for hierarchical structure',
    name VARCHAR(100) NOT NULL COMMENT 'Category name',
    remark TEXT COMMENT 'Additional remarks or notes',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    modify_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification timestamp',
    INDEX idx_parent_id (parent_id),
    INDEX idx_name (name),
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Transaction categories';

-- Create cash_flows table
-- Aligned with Go CashFlowEntity model:
-- - id: UUID primary key
-- - category_id: foreign key to categories
-- - belongs_date: date of transaction
-- - flow_type: 'income' or 'expense'
-- - amount: decimal amount with precision
-- - description: transaction description
-- - remark: additional notes
-- - create_time: creation timestamp
-- - modify_time: last modification timestamp
CREATE TABLE IF NOT EXISTS cash_flows (
    id VARCHAR(36) PRIMARY KEY COMMENT 'UUID identifier',
    category_id VARCHAR(36) NOT NULL COMMENT 'Foreign key to categories table',
    belongs_date DATE NOT NULL COMMENT 'Date of transaction',
    flow_type VARCHAR(20) NOT NULL COMMENT 'Transaction type: income or expense',
    amount DECIMAL(19, 4) NOT NULL COMMENT 'Transaction amount with 4 decimal places',
    description TEXT COMMENT 'Transaction description',
    remark TEXT COMMENT 'Additional remarks',
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    modify_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification timestamp',
    INDEX idx_belongs_date (belongs_date),
    INDEX idx_flow_type (flow_type),
    INDEX idx_category_id (category_id),
    INDEX idx_date_type (belongs_date, flow_type),
    INDEX idx_date_category (belongs_date, category_id),
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Cash flow transactions';

-- Insert basic default categories (auto-loaded on init)
-- These categories are available for all users by default
INSERT INTO categories (id, parent_id, name, remark) VALUES
(UUID(), NULL, 'Salary', 'Income from employment'),
(UUID(), NULL, 'Freelance', 'Income from freelance work'),
(UUID(), NULL, 'Investment', 'Income from investments and dividends'),
(UUID(), NULL, 'Other Income', 'Other income sources'),
(UUID(), NULL, 'Food & Dining', 'Restaurants, groceries, food delivery'),
(UUID(), NULL, 'Transportation', 'Gas, public transport, car maintenance'),
(UUID(), NULL, 'Shopping', 'Retail purchases, online shopping'),
(UUID(), NULL, 'Entertainment', 'Movies, games, hobbies'),
(UUID(), NULL, 'Healthcare', 'Medical expenses, pharmacy, fitness'),
(UUID(), NULL, 'Utilities', 'Electricity, water, internet, phone');

-- Print initialization summary
SELECT 
    'Schema initialized successfully!' AS message,
    (SELECT COUNT(*) FROM categories) AS default_categories,
    (SELECT COUNT(*) FROM cash_flows) AS initial_transactions,
    NOW() AS initialized_at;
