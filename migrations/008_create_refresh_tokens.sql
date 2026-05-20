CREATE TABLE IF NOT EXISTS refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    revoked_at DATETIME,
    revoked_by VARCHAR(36),
    device_id VARCHAR(100),
    device_name VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    
    -- BaseEntity columns
    create_user_id VARCHAR(36) NOT NULL,
    create_time DATETIME NOT NULL,
    update_user_id VARCHAR(36) NOT NULL,
    update_time DATETIME NOT NULL,
    delete_user_id VARCHAR(36),
    delete_time DATETIME,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_token (token)
);
