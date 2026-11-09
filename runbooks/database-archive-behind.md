---
id: database-archive-behind
title: The control plane's database is not reaching its archive
applies_to:
  alerts: [database-archive-behind]
  jobs: []
  events: []
  dashboards: [spin-db-performance]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**The Database's Archive Is Behind** fires when PostgreSQL has finished WAL files it has not sent
to the archive (`database_wal_archive_pending > 0`), or any send to the archive failed in the last
15 minutes (`database_wal_archive_failures_total`), and this lasts 12 minutes: twice the archive
timeout (5 minutes) and two more, since one file is always on its way.

A quiet database archives nothing for hours and is not behind: the age of the last archived file
alone does not fire this.

Nobody using the installation sees anything.

## Impact

The archive (base backups and WAL in the bucket the installation document names under
`database.archive`) is what a replaced control plane machine restores the database from. What has
not reached it is lost if the machine dies now, and an archive that stopped is found out only at
that restore. A release that replaces the machine needs it too.

## Diagnose

1. Ask the installation:

   ```
   spin admin health
   ```

   The `database archive` part says whether it is behind and why.
2. *WAL waiting for the archive*, *Since the last archived WAL* and *Archive failures in the last
   hour* on [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance).
3. Read why the sends fail. PostgreSQL runs the archive command under the database's unit, on the
   control plane's machine:

   ```
   journalctl -u spin-postgres.service --since "1 hour ago"
   ```

   The usual causes: the machine's role lost permission on the archive bucket or its key, the
   bucket's endpoint is unreachable, or the disk under the WAL is full.
4. Check the daily base backup: `systemctl list-timers spin-postgres-backup.timer` and
   `journalctl -u spin-postgres-backup.service`.

## Fix

- **Permission or key**: restore the control plane machine's access to the archive bucket and its
  key (on AWS, the role's policy and the KMS key policy, both from the installation's Terraform).
  PostgreSQL retries the waiting files on its own.
- **Network**: restore the route to the bucket's endpoint.
- **Disk full**: free space on the machine's data disk. PostgreSQL keeps every WAL file it has not
  archived, so an archive that stays down fills the disk; that is the next outage.
- Once it catches up, take a fresh base backup so a restore does not need the long WAL tail:

  ```
  systemctl start spin-postgres-backup.service
  ```

For restoring from the archive, or moving the database to a new machine, see
[database-restore](database-restore.md).

## Verify

`spin admin health` shows the archive ok, *WAL waiting for the archive* is 0 or 1 on the database
dashboard, and the alert resolves.

## Escalate

If the archive keeps failing with permissions and network correct, stop before replacing the
control plane's machine: a replacement now loses data. Open an issue at
https://github.com/spin-stack/spin/issues with the `spin-postgres.service` journal lines (remove
bucket names first: the tracker is public), `spin admin health`, and `spin version`.
