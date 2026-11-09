---
id: workspace-suspends-failing
title: A workspace could not be suspended
applies_to:
  alerts: [workspace-suspends-failing]
  jobs: []
  events: []
  dashboards: [fleet]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**A Workspace Could Not Be Suspended** fires as soon as any suspension failed in the last 30
minutes (`workspace_suspend_duration_seconds_count{outcome="failed"}`, by `reason`). The `reason`
label says who asked:

| `reason` | Who asked |
|---|---|
| `drain` | A host being drained with `--suspend`, or going away for good. |
| `owner` | The workspace's owner put it down. |
| `stop` | A stop on a host that can be taken away, which publishes the disk alone. |
| `move` | A move to a host with room for what it holds. |
| `idle` | Nobody used it for as long as the installation allows. |
| `running_cap` | It ran at a stretch for as long as the installation allows. |

## Impact

A suspension that fails leaves the workspace running where it was. For `drain`, that is a
workspace still running on a machine about to go: if the machine goes, its owner can recover it
only by giving up what it had not published. For `stop`, the owner is told the stop's backup
failed ([workspace-stop-backup-failed](workspace-stop-backup-failed.md)).

## Diagnose

1. See how many and which reason: *Workspaces suspended off their host, by who asked and how it
   ended* on [Admin → Metrics](/admin/metrics?dashboard=fleet).
2. For a drain, the workspaces are named:
   - in what `spin admin runner drain <runner-id> --suspend` printed: `Workspace ... was not
     suspended and runs on: ...`;
   - in [Admin → Audit](/admin/audit), the `host_evacuated` entry of the drain.
3. Read why it failed in the control plane's log:

   ```
   spin admin logs 'service_name:spin-controlplane AND severity_text:warn AND body.message:suspend' --since 1h
   ```

4. A suspension writes the workspace's memory and disk to the object store: check the host
   ([Admin → Runners](/admin/runners)) and the store (*Checkpoint memory* panels on
   [Admin → Metrics](/admin/metrics?dashboard=storage)).

## Fix

- **A drain**: do not take the machine away yet. Run the drain again; it suspends what still runs:

  ```
  spin admin runner drain <runner-id> --suspend
  ```

  If a workspace fails again, its owner can stop it, which publishes its disk and gives up only
  what was in memory (`spin workspace stop <workspace>`). Only then retire the machine.
- **The host cannot write to the store**: follow
  [object-store-requests-runaway](object-store-requests-runaway.md) or
  [nothing-published](nothing-published.md).
- **The host is not answering**: follow [host-not-reporting](host-not-reporting.md).
- **`idle`, `running_cap`, `owner`, `move`**: the workspace runs on where it is, nothing lost.
  Fix the cause above.

## Verify

`spin admin runner list` shows no workspaces left on a drained host. No new failures on the panel,
and the alert resolves after 30 minutes.

## Escalate

If suspensions fail with the host and the store healthy, open an issue at
https://github.com/spin-stack/spin/issues with the reason, the control plane's log lines, a trace
id from [Admin → Traces](/admin/traces) for the failed suspension, and `spin version`.
