USE `cashlenx`;

ALTER TABLE `cash_flows`
    MODIFY COLUMN `amount` DECIMAL(18, 2) NOT NULL;
