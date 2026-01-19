USE `cashlenx`;

-- -------------------
-- Create table `password_reset_tokens`
-- -------------------
DROP TABLE IF EXISTS password_reset_tokens;
CREATE TABLE `password_reset_tokens`
(
    `id`              VARCHAR(24)  NOT NULL,
    `belongs_user_id` VARCHAR(24)  NOT NULL,
    `token`           VARCHAR(255) NOT NULL,
    `expires_at`      TIMESTAMP    NOT NULL,
    `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `used_at`         TIMESTAMP    NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX password_reset_tokens_token_unique_index ON password_reset_tokens (token),
    INDEX password_reset_tokens_belongs_user_id_index ON password_reset_tokens (belongs_user_id),
    INDEX password_reset_tokens_expires_at_index ON password_reset_tokens (expires_at),
    CONSTRAINT password_reset_tokens_belongs_user_id_fk FOREIGN KEY (belongs_user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='Password Reset Tokens';
