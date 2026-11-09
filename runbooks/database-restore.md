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
last_reviewed: 2026-11-09
---

## Symptoms

Read this when a release replaces the control plane's machine, when a new control plane machine
waits a long time before it serves, or when the database will not start.

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

## Impact

While the new machine hands over or restores, the control plane does not start: the API and the
dashboard are down. Running workspaces keep running. A restore after a machine died loses what
that machine wrote and had not archived: at most about five minutes of writes, the archive's
timeout.

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

TODO: there is no command to restore the database to an earlier point in time or to a chosen
base backup; a restore always takes the newest ([spin#1052](https://github.com/spin-stack/spin/issues/1052)).

## Verify

`spin admin health` shows the database and its archive ok and exits 0, the dashboard loads, and
*Since the last archived WAL* on the database dashboard is under five minutes.

## Escalate

Stop at any error line above, at a restore that has not finished replaying after an hour, or
at anything that would mean deleting from the archive. Bring the `spin-postgres.service`
journal from both machines, `spin admin health` output, the release you moved from and to
(`spin admin versions` before and after), and open an issue at
https://github.com/spin-stack/spin/issues.
