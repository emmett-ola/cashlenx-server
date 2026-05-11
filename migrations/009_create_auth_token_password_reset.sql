CREATE TABLE IF NOT EXISTS auth_token_password_reset (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    operation_type VARCHAR(50) NOT NULL DEFAULT 'password_reset',
    payload TEXT,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    ip_address VARCHAR(45),
    
    -- BaseEntity columns
    create_user_id VARCHAR(36) NOT NULL,
    create_time DATETIME NOT NULL,
    update_user_id VARCHAR(36) NOT NULL,
    update_time DATETIME NOT NULL,
    delete_user_id VARCHAR(36),
    delete_time DATETIME,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_operation_type (operation_type),
    INDEX idx_token (token)
);
