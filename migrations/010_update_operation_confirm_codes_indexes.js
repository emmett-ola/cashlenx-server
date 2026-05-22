// MongoDB verification-code index update.
// Run with: mongosh <connection_string> < 010_update_operation_confirm_codes_indexes.js

use cashlenx;

try {
    db.operation_confirm_codes.dropIndex("code_1");
} catch (e) {
    print("code_1 index did not exist or was already updated");
}
db.operation_confirm_codes.createIndex({ code: 1 });
db.operation_confirm_codes.createIndex({ verification_token: 1 }, { unique: true, sparse: true });
db.operation_confirm_codes.createIndex({ operation_type: 1, payload: 1 });
