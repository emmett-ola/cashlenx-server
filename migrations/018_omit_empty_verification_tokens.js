// MongoDB verification-token cleanup migration.
// Runtime execution is implemented by the Go migration handler; this immutable
// asset defines the ordered migration identity and checksum.

db.operation_confirm_codes.updateMany(
  { verification_token: '' },
  { $unset: { verification_token: '' } }
);
