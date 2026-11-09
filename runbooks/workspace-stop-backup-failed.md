---
id: workspace-stop-backup-failed
title: A workspace could not be stopped because its disk could not be published
applies_to:
  alerts: []
  jobs: []
  events: [STOP_BACKUP_FAILED]
  dashboards: [storage, fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

The owner is emailed (on by default) and sees on [Settings → Notifications](/settings/notifications):

> Workspace `<workspace>` could not be stopped

with the reason the last attempt failed, and the command:

```
spin workspace stop <workspace>
```

The workspace is still running.

## Impact

On a runner that can be taken away (a preemptible or autoscaling group's machine), a stop
publishes the disk to the object store first, so the workspace can start on any runner later.
That publish was tried three times and failed every time, so the stop was given up and the
workspace runs on. Nothing is lost yet: the disk is where it was. But the workspace keeps
costing, and if the machine is reclaimed before a stop succeeds, what it had not published is
lost.

## Diagnose

1. Read the reason in the notice, or in the owner's `spin workspace history <workspace>`.
2. Is the runner connected? [Admin → Runners](/admin/runners), or `spin admin runner list`. A
   stop on such a runner is refused while the runner is not connected.
3. Is the object store answering? *The object store itself* on
   [Admin → Metrics](/admin/metrics?dashboard=storage), and *Publishing* for failed publishes.
   `spin admin health` names a failed part.
4. Search what the runner logged around the attempts:

   ```
   spin admin logs 'service_name:runner AND severity_text:error' --since 1h
   ```

## Fix

- **The runner was disconnected for a while**: once it is connected again, the owner runs
  `spin workspace stop <workspace>` again.
- **The object store refused or timed out**: fix the store first (see
  [volume-not-served](volume-not-served.md) and
  [object-store-requests-runaway](object-store-requests-runaway.md)), then stop again.
- **The runner is about to go, and the work on it does not matter**: the owner can stop without
  publishing, giving up what only that runner has:

  ```
  spin workspace stop <workspace> --force
  ```

TODO: an administrator cannot stop another person's workspace; `StopWorkspace` answers only the
owner ([spin#1045](https://github.com/spin-stack/spin/issues/1045)).

## Verify

`spin workspace get <workspace>` shows it stopped (suspended at its published disk), and a
start places it on any runner.

## Escalate

If the store answers and the runner is connected but the stop still fails, stop and open an
issue at https://github.com/spin-stack/spin/issues with the reason from the notice, the
workspace and runner ids, the failing call's trace id from [Admin → Traces](/admin/traces), the
runner's error log lines and `spin admin versions`.
