// MongoDB category presentation migration.
// Runtime execution is implemented by the Go migration handler; this immutable
// asset defines the ordered migration identity and checksum.

db.categories.updateMany(
  { $or: [{ emoji: { $exists: false } }, { emoji: '' }] },
  { $set: { emoji: '🙂' } }
);
db.categories.updateMany(
  { $or: [{ bg_color: { $exists: false } }, { bg_color: '' }] },
  { $set: { bg_color: '#E5E7EB' } }
);
