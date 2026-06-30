USE `cashlenx`;

-- -------------------
-- Create table `cash`
-- -------------------
DROP TABLE IF EXISTS cash_flows;
CREATE TABLE `cash_flows`
(
    `id`              VARCHAR(24)  NOT NULL,
    `belongs_user_id` VARCHAR(24)  NOT NULL,
    `category_id`     VARCHAR(24)  NOT NULL,
    `belongs_date`    TIMESTAMP    NOT NULL,
    `amount`          DECIMAL      NOT NULL,
    `description`     VARCHAR(200) NOT NULL,
    `remark`          VARCHAR(200)          DEFAULT NULL COMMENT 'KEEP EMPTY',
    `create_user_id`  VARCHAR(24)  NOT NULL,
    `create_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`  VARCHAR(24)  NOT NULL,
    `update_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`  VARCHAR(24)           DEFAULT NULL,
    `delete_time`     TIMESTAMP             DEFAULT NULL,
    `is_delete`       BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Cash Flow Table';

CREATE INDEX cash_flows_belongs_user_id_index ON cash_flows (belongs_user_id);
CREATE INDEX cash_flows_category_id_index ON cash_flows (category_id);
CREATE INDEX cash_flows_belongs_date_index ON cash_flows (belongs_date);
