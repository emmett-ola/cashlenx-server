USE `cashlenx`;

ALTER TABLE `users`
    ADD COLUMN `phone_number` VARCHAR(32) NULL AFTER `gender`,
    ADD COLUMN `location` VARCHAR(200) NULL AFTER `phone_number`,
    ADD COLUMN `birth_date` CHAR(10) NULL AFTER `location`;
