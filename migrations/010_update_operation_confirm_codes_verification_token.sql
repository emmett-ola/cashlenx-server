ALTER TABLE operation_confirm_codes
    DROP FOREIGN KEY operation_confirm_codes_user_id_fk,
    DROP INDEX operation_confirm_codes_token_unique_index,
    MODIFY user_id VARCHAR(24) NULL,
    MODIFY create_user_id VARCHAR(24) NULL,
    MODIFY update_user_id VARCHAR(24) NULL,
    ADD COLUMN verification_token VARCHAR(255) NULL AFTER token,
    ADD INDEX operation_confirm_codes_token_index (token),
    ADD UNIQUE INDEX operation_confirm_codes_verification_token_unique_index (verification_token),
    ADD INDEX operation_confirm_codes_purpose_payload_index (operation_type, payload(255));
