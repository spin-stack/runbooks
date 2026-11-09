---
id: workspace-unresponsive
title: A workspace stopped responding and was stopped
applies_to:
  alerts: []
  jobs: [silent-supervisors]
  events: [UNRESPONSIVE]
  dashboards: [fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

The owner sees on [Settings → Notifications](/settings/notifications) (email is off by default
for this one):

> Workspace `<workspace>` stopped responding and was stopped

with the command `spin workspace start <workspace>`.

The **silent-supervisors** job on [Admin → Jobs](/admin/jobs) counts each such stop as
`supervisor_silence_stopped`; [Admin → Audit](/admin/audit) has a `workspace_supervisor_silence`
entry with the last heartbeat. After a control plane restart or a new leader, the job reports
`still_listening` for a few minutes and stops nothing.

## Impact

The supervisor inside the workspace said nothing for longer than the control plane waits, so the
workspace was stopped by force. Its disk is kept; a start boots it again, as after a power cut:
running processes and unsaved editor state are gone.

If the **silent-supervisors** job itself fails, such a workspace runs on unwatched: its machine
may reach out while nothing reports what it does.

## Diagnose

1. One workspace, or many? Many at once on one runner is the runner, not the guests: follow
   [host-not-reporting](host-not-reporting.md).
2. Why did the guest go quiet? The owner reads its console:

   ```
   spin workspace console <workspace> --lines 300
   ```

   or, where the installation collects consoles, every run of it:

   ```
   spin workspace logs <workspace> 'panic OR "out of memory"' --since 24h
   ```

   An out-of-memory guest is the usual cause; *Containment and DNS* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet) counts OOM kills.
3. If the job fails: `spin admin jobs list` shows its last error.

## Fix

- **The guest ran out of memory**: the owner starts it again and gives it more
  (`spin workspace grant <workspace> --memory <MB>`), or uses a larger preset.
- **It keeps happening to one workspace with nothing in its console**: compare
  `spin workspace status <workspace>`, which shows what the supervisor last said, and escalate.
- **The job fails**: it fails when it cannot reach the database; see
  [database-slow](database-slow.md). Then run it again:

  ```
  spin admin jobs run silent-supervisors
  ```

## Verify

The owner's `spin workspace start <workspace>` brings it to running and `spin workspace status`
shows the supervisor reporting. The job's last run on Admin → Jobs succeeded.

## Escalate

If guests with enough memory go silent again and again, or the job fails with the database
healthy, stop and open an issue at https://github.com/spin-stack/spin/issues with the workspace
id, the console tail, the audit entries, the job's error from `spin admin jobs list` and
`spin admin versions`.
