---
id: nothing-published
title: A volume has published nothing for 30 minutes
applies_to:
  alerts: [nothing-published]
  jobs: []
  events: []
  dashboards: [storage, fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**A Volume Has Published Nothing For 30 Minutes** fires when, for some volume a host serves, the
host last published a commit to the object store more than 30 minutes ago
(`max(last_successful_commit_age_seconds) > 1800`), for 10 minutes. The description gives the
oldest age.

Nobody sees anything: the workspace runs as usual.

## Impact

What the workspace wrote since its last commit lives only on its host. If the host is lost now,
that work is lost: this age is the recovery point (RPO). A stop or a move also waits on publishing.

## Diagnose

1. Find the volume and its host: *RPO: time since each volume last published* on
   [Admin → Metrics](/admin/metrics?dashboard=storage), with *Unpublished bytes* beside it. Then
   `spin admin storage list` says whose volume it is.
2. One volume or all of a host's?
   - **All on one host**: the host cannot publish. Check *Is the store answering at all* and *A
     host's reconciliation* on the storage dashboard, and the host on
     [Admin → Runners](/admin/runners).
   - **One volume**: check whether its host refuses it ([volume-not-served](volume-not-served.md)),
     and *Rotations* and *A publish, end to end* on the storage dashboard.
3. Read the runner's warnings:

   ```
   spin admin logs 'service_name:spin-runner AND severity_text:warn' --since 1h
   ```

## Fix

- **The store is not answering** for that host: its network or credentials. Follow
  [object-store-requests-runaway](object-store-requests-runaway.md) for the store's side and
  [host-not-reporting](host-not-reporting.md) if the host is unwell.
- **The volume is refused** (`durability_lost`, `publish_fenced`, ...): follow
  [volume-not-served](volume-not-served.md).
- **The host is stuck and the store is fine**: move its workspaces off before you lose it.
  `spin admin runner drain <runner-id> --suspend` suspends each running workspace, which publishes
  its disk, and resumes it on another runner. A workspace whose suspension fails stays where it
  is ([workspace-suspends-failing](workspace-suspends-failing.md)).

## Verify

The volume's line on the RPO panel drops back to minutes, and the alert resolves.

## Escalate

A volume that does not publish while its host, the store and the volume all look healthy is a
defect. Open an issue at https://github.com/spin-stack/spin/issues with the volume id, the storage
panels over the window, the runner's log lines and `spin version`.
