---
id: workspace-held
title: A workspace waits for its owner after its host stopped answering
applies_to:
  alerts: []
  jobs: []
  events: [HELD]
  dashboards: [fleet, storage]
severity: critical
owner: operator
last_reviewed: 2026-11-12
---

## Symptoms

The owner is emailed (on by default) and sees on their notifications page
([Settings → Notifications](/settings/notifications)) and in the workspace's history:

> Workspace `<workspace>` is waiting for you: its host stopped answering

with what moving it would give up and the command:

```
spin workspace recover <workspace> --accept-data-loss
```

The workspace cannot start: a workspace starts only on its own host, and that host is silent.

## Impact

The owner has no workspace until either its host comes back or they give up the work the host
had not published to the object store. That work is on the silent host's disk and nowhere else.

The control plane moves a workspace off a silent host on its own only when the installation
says so: `rehome_after_seconds` set, and the host's unpublished bytes at or under
`rehome_max_unpublished_bytes`. HELD means one of these: the cost is over that limit, or the
host never reported what it had not published. Both settings unset (the default) means nothing
is ever moved on its own, and the owner must recover by hand.

## Diagnose

1. Is the host coming back? Open [Admin → Runners](/admin/runners), or:

   ```
   spin admin runner list
   ```

   A runner that is not connected, with its last heartbeat old, is the cause. If many runners
   went at once, follow [host-not-reporting](host-not-reporting.md) first.
2. What would moving cost? *What losing a host would cost* on
   [Admin → Metrics](/admin/metrics?dashboard=storage) shows the RPO and unpublished bytes per
   host. The owner sees the same for their workspace by running `recover` without the flag:

   ```
   spin workspace recover <workspace>
   ```

   which changes nothing and prints the cost.
3. `spin workspace status <workspace>` (run by the owner) shows what the control plane, the
   supervisor, the host and the volume each believe.

## Fix

- **The host can be brought back** (a reboot, a network fault): bring it back. The workspace runs
  on it again with nothing lost. Nobody needs to recover anything.
- **The host is gone for good**: tell the owner. The owner recovers it with
  `spin workspace recover <workspace> --accept-data-loss`. It takes the disk from the dead host
  and places it on a live one; then they start it. Afterwards retire the host
  ([lost-runner](lost-runner.md)).
- **The owner cannot recover it, and you recover it for them**: run it first without
  `--accept-data-loss` to see the work that is given up; nothing changes until you pass it.
  `--reason` is required:

  ```
  spin admin workspace recover <workspace-id> --reason "<why>"
  spin admin workspace recover <workspace-id> --reason "<why>" --accept-data-loss
  ```

  It is audited, and written to the workspace's history with your reason, which its owner sees
  on Activity.
- **The installation should move such workspaces on its own** next time: set
  `rehome_after_seconds` and `rehome_max_unpublished_bytes` in the installation's configuration
  file, and apply the change on [Admin → Settings](/admin/settings). Giving up unpublished work
  without asking is a policy decision; make it on purpose.

## Verify

`spin workspace get <workspace>` shows a live host, and the workspace starts. A recovery is in
[Admin → Audit](/admin/audit) as `workspace_recovered`, with the bytes given up.

## Escalate

Stop if `recover --accept-data-loss` fails with no host able to take the disk
(`RESOURCE_EXHAUSTED`) while runners have room, or if the host is connected and the workspace
still will not start. Bring the workspace id, the runner id, the output of
`spin workspace status <workspace>`, the trace id of the failed call from
[Admin → Traces](/admin/traces) and `spin version`, and open an issue at
https://github.com/spin-stack/spin/issues.
