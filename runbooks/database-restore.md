---
id: database-restore
title: The control plane's database - archive, restore, and handover at a release
applies_to:
  alerts: []
  jobs: []
  events: []
  dashboards: [spin-db-performance]
severity: info
owner: operator
last_reviewed: 2026-11-12
---

## Symptoms

Read this when a release replaces the control plane's machine, when a new control plane machine
waits a long time before it serves, when the database will not start, or when a bad write (a
lockout, a destructive change) has to be undone by restoring the database to before it.

How it works, so the logs make sense:

- The database is PostgreSQL on the control plane's machine (`spin-postgres.service`, which runs
  `controlplane database exec`). Every WAL segment, and a base backup a day
  (`spin-postgres-backup.timer`), go sealed to the installation's archive bucket
  (`database.archive` in the installation's document). Seven base backups are kept: daily ones,
  one before every handover and one after every restore.
- **Handover, at a release.** The new machine asks, in the handover bucket
  (`database.handover`), for the database. The machine that holds it stops, takes a base backup,
  archives what it wrote and lets go; the new machine lays that backup and replays the archive.
  The old machine's unit exits and stays down.
- **Restore.** A machine with no database of its own, whose holder has not said it holds it for
  a minute, restores the newest base backup and replays every WAL segment after it.
- **Restore to a point.** An operator on the control plane's machine can restore the database
  to a time, a place in the WAL or a chosen base backup (`spin-database restore`, below). It
  replays to that point and starts a new history there: the writes after it are left out, and
  every later restore follows the new history.

## Impact

While the new machine hands over or restores, the control plane does not start: the API and the
dashboard are down. Running workspaces keep running. A restore after a machine died loses what
that machine wrote and had not archived: at most about five minutes of writes, the archive's
timeout. A restore to a point loses every write after that point, the bad one and every good one
since: pick the latest point before the bad write.

## Diagnose

1. From anywhere:

   ```
   spin admin health
   ```

   The *database* and *database archive* parts say whether it answers and whether WAL is waiting.
   [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance), section *The archive*, shows
   WAL waiting, time since the last archived WAL and archive failures.
2. On the control plane's machine (its Session Manager shell, or however the installation reaches
   it):

   ```
   sudo systemctl status spin-postgres.service
   sudo journalctl -u spin-postgres.service --since -1h
   ```

   The log lines say what it is doing:
   - `asked the machine that holds it to hand it over`, then `handed over`: a normal release.
   - `did not take up the request; restoring what the archive has` or `stopped saying so before
     it handed it over`: the old holder is gone; a restore follows.
   - `laid the archive's newest base backup`: replaying; this can take minutes.
   - `handed over without WAL the other machine could not archive`: some writes are only on the
     old machine's disk.
   - `the archive holds WAL and no complete base backup`: the database refuses to start empty.
3. What the archive holds, to choose a point to restore to (on the control plane's machine):

   ```
   sudo spin-database bases
   ```

   One line per complete base backup: its id, when it was taken (UTC), its timeline and the WAL
   segment it starts in. Each restores to any time from when it was taken to the end of the WAL,
   whose last segment and the time the bucket says it was stored are printed below the list.

## Fix

- **Before a release**, make sure the archive is current: `spin admin health` shows the archive
  healthy, and [database-archive-behind](database-archive-behind.md) is not firing. To take a
  fresh base backup now, on the control plane's machine:

  ```
  sudo systemctl start spin-postgres-backup.service
  ```

- **A handover is slow**: the new machine waits as long as the old one keeps saying it holds the
  database. Let it finish; do not stop the old machine mid-handover.
- **The old machine is gone**: nothing to do; the new machine restores from the archive on its
  own. Keep both buckets and the installation's encryption key: the archive is sealed under keys
  derived from it, and a machine without every version of the key cannot read it.
- **WAL left on the old machine** or **no complete base backup**: stop and escalate. Do not delete
  anything in the archive or handover buckets, and do not start a database by hand.

- **Undo a bad write** by restoring to a time before it. On the control plane's machine:

  1. Find when the bad write happened (the audit log, the control plane's journal) and pick a
     time just before it, in UTC: `2026-11-12T09:30:00Z`.
  2. Stop the control plane and the database. Stopping the database archives its last segment;
     its log says `stopped with everything it wrote archived`.

     ```
     sudo systemctl stop spin-controlplane.service spin-postgres.service
     ```

  3. Restore. The newest base backup taken before the time is laid and replayed to it:

     ```
     sudo spin-database restore --to 2026-11-12T09:30:00Z
     ```

     `--base <id>` (from `spin-database bases`) lays that base backup instead, and with no `--to`
     replays it to the end. `--to-lsn <lsn>` (like `0/16B3748`) ends at a place in the WAL
     instead of a time. The restore is laid beside the machine's database, which is replaced
     only once the restore has come up: a refused or failed restore leaves it as it was. It
     ends with `restored; start spin-postgres.service to serve it`.
  4. Start them again; the database serves the restored copy:

     ```
     sudo systemctl start spin-postgres.service spin-controlplane.service
     ```

  What a restore refuses, and what to do instead:
  - `the database is serving: on this machine`: step 2 was skipped; stop the database first.
  - `the database is serving: on <machine>`: another machine holds the database. Restore there,
    or stop its database first.
  - `has not happened yet`: the time is in the future. For the newest state there is nothing to
    restore; leave out `--to` to replay to the end.
  - `its oldest base backup ... was taken at ..., after ...` or `base backup ... was taken at
    ..., after ...`: the archive holds nothing that far back (seven base backups are kept), or
    the base named is newer than the time. Choose a later time or an older base backup.
  - `its WAL ends before ...`: the time is after the last write the archive holds. Leave out
    `--to` to replay to the end.
  - `holds no complete base backup <id>`: copy the id from `spin-database bases`.

## Verify

`spin admin health` shows the database and its archive ok and exits 0, the dashboard loads, and
*Since the last archived WAL* on the database dashboard is under five minutes. After a restore to
a point, what the bad write changed is back as it was, and `spin-database bases` lists a new base
backup, on a timeline one higher, taken as the restore finished.

## Escalate

Stop at any error line above, at a restore that has not finished replaying after an hour, or
at anything that would mean deleting from the archive. Bring the `spin-postgres.service`
journal from both machines, `spin admin health` output, the release you moved from and to
(`spin admin versions` before and after), and open an issue at
https://github.com/spin-stack/spin/issues.
