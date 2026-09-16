# Database-Level Data Protection

These operator scripts protect the complete selected database, including
migration state. They complement the application JSON export/import features;
they do not replace user-scoped portability or the compensating application
restore path.

## Backup

Configure a repository-local environment file from `.env.example` and a
non-empty, one-line passphrase file that is not committed or stored beside the backup.
The selected database dependency must already be running.

```bash
ENV_FILE=.env scripts/data-protection/backup.sh daily
ENV_FILE=.env scripts/data-protection/backup.sh weekly
ENV_FILE=.env scripts/data-protection/backup.sh monthly
```

The script:

- checks tools, selected database, key file, destination safety, and free space;
- creates a MongoDB archive or consistent MySQL logical dump;
- includes metadata, migration state, and an internal checksum manifest;
- encrypts the package with OpenSSL AES-256-CBC, PBKDF2, SHA-256, and 200,000
  iterations;
- writes the encrypted artifact and external SHA-256 sidecar atomically;
- retains the newest configured 7 daily, 4 weekly, or 12 monthly artifacts; and
- writes a secret-free `status/latest.json` result for scheduler monitoring.

Files are created with owner-only permissions where the host filesystem honors
POSIX modes. Dumps exist as restricted plaintext only in the per-run staging
directory and are removed on normal success, failure, or termination. Protect
the backup root from other host users and investigate any stale `.staging.*`
directory left by an abrupt host or process kill before removing it.

Schedule `daily` every day, `weekly` once per week after a successful daily
window, and `monthly` once per month after a successful daily window. Serialize
runs with the scheduler so two backup processes do not target the same root.
Alert on a non-zero exit and when `status/latest.json` has no recent success
within `BACKUP_MAX_AGE_HOURS`. Storage replication, encryption-key custody, and
notification delivery are operator-owned and should use different failure
domains from the application node.

## Disposable restore drill

```bash
ENV_FILE=.env scripts/data-protection/restore-drill.sh \
  backups/daily/cashlenx-mongodb-YYYYMMDDTHHMMSSZ.tar.gz.enc
```

The drill verifies the external checksum, decrypts the package, rejects unsafe
archive paths, verifies the internal manifest, starts an unnetworked disposable
database container, restores the dump, verifies database objects and migration
state, writes secret-free JSON evidence, and removes the container and temporary
plaintext on exit. The configured database image must already exist locally.

The drill never connects to or mutates the configured source database. Run it
quarterly and after changing database versions, migration runners, backup
scripts, encryption settings, or storage tooling.

## Failure behavior

- A capacity failure occurs before the database dump begins.
- Dump, encryption, checksum, or atomic-write failure leaves no completed
  artifact and does not prune prior backups.
- A missing checksum, checksum mismatch, wrong key, corrupt payload, unsafe
  archive path, missing migration state, or empty restore fails the drill.
- Retention only removes matching completed artifacts inside the resolved tier
  directory after a new encrypted artifact succeeds.
- A production restore is not automated here. It requires an environment-
  specific recovery plan, compatibility review, verified backup, rollback
  readiness, and explicit authorization before production data is touched.
