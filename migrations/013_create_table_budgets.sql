USE `cashlenx`;

CREATE TABLE IF NOT EXISTS `budgets`
(
    `id`              VARCHAR(24)    NOT NULL,
    `belongs_user_id` VARCHAR(24)    NOT NULL,
    `category_id`     VARCHAR(24)    NOT NULL,
    `period`          CHAR(7)        NOT NULL,
    `limit_amount`    DECIMAL(18, 2) NOT NULL,
    `create_user_id`  VARCHAR(24)    NOT NULL,
    `create_time`     TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`  VARCHAR(24)    NOT NULL,
    `update_time`     TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`  VARCHAR(24)             DEFAULT NULL,
    `delete_time`     TIMESTAMP               DEFAULT NULL,
    `is_delete`       BOOLEAN        NOT NULL DEFAULT FALSE,
	`active_scope_key` VARCHAR(80) GENERATED ALWAYS AS (
		IF(`is_delete` = FALSE, CONCAT(`belongs_user_id`, '|', `period`, '|', `category_id`), NULL)
	) STORED,
    PRIMARY KEY (`id`),
	UNIQUE INDEX `budgets_active_scope_unique_index` (`active_scope_key`),
    INDEX `budgets_user_period_active_index` (`belongs_user_id`, `period`, `is_delete`),
    INDEX `budgets_user_category_index` (`belongs_user_id`, `category_id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Monthly Budget Table';
