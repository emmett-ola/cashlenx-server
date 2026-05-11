USE `cashlenx`;

-- -------------------
-- Create table `operation_confirm_codes`
-- -------------------
DROP TABLE IF EXISTS operation_confirm_codes;
CREATE TABLE `operation_confirm_codes`
(
    `id`             VARCHAR(24)  NOT NULL,
    `user_id`        VARCHAR(24)  NOT NULL,
    `token`          VARCHAR(255) NOT NULL,
    `operation_type` VARCHAR(50)  NOT NULL DEFAULT 'password_reset',
    `payload`        TEXT         NULL,
    `expires_at`     TIMESTAMP    NOT NULL,
    `used_at`        TIMESTAMP    NULL,
    `create_user_id` VARCHAR(24)  NOT NULL,
    `create_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id` VARCHAR(24)  NOT NULL,
    `update_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id` VARCHAR(24)           DEFAULT NULL,
    `delete_time`    TIMESTAMP             DEFAULT NULL,
    `is_delete`      BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`),
    UNIQUE INDEX operation_confirm_codes_token_unique_index ON operation_confirm_codes (token),
    INDEX operation_confirm_codes_user_id_index ON operation_confirm_codes (user_id),
    INDEX operation_confirm_codes_operation_type_index ON operation_confirm_codes (operation_type),
    INDEX operation_confirm_codes_expires_at_index ON operation_confirm_codes (expires_at),
    CONSTRAINT operation_confirm_codes_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='Operation Confirmation Codes';
