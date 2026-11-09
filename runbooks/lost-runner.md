---
id: lost-runner
title: A runner is gone for good, or should be forgotten
applies_to:
  alerts: []
  jobs: [forget-gone-hosts]
  events: []
  dashboards: [fleet]
severity: info
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

[Admin → Runners](/admin/runners) shows a runner *Not connected* that will not come back: its
machine was terminated, replaced or lost. Or the list keeps machines an autoscaling group started
long ago. Also read this when the **forget-gone-hosts** job fails on [Admin → Jobs](/admin/jobs).

## Impact

A runner that is gone still counts in the list and may still be the only home of some workspaces'
disks: those workspaces cannot start until they are recovered. While **forget-gone-hosts** fails,
the runners list keeps every machine an autoscaling group ever started.

How a runner leaves depends on how it came:

- **An autoscaling group's machine** is forgotten on its own by **forget-gone-hosts** once it has
  been silent for `host_forget_after_hours` (24 by default), and no workspace's disk lives only
  on it. Its other volumes are detached first.
- **A runner registered with a token** (a machine that stays) is never forgotten on its own: a
  silent one is somebody's to look at, and it leaves only by revoke.

## Diagnose

```
spin admin runner list
```

Open the runner on [Admin → Runners](/admin/runners): its row's menu offers the one way it can
leave (*Revoke runner*, *Give back to the* `<class>` *group*, or *Forget runner*), or says why it is refused,
naming the workspaces whose disks are only there. Whether it is just slow to report, rather than
gone: [host-not-reporting](host-not-reporting.md).

## Fix

1. **Workspaces whose disk is only on it**: their owners recover them, giving up what the runner
   had not published ([workspace-held](workspace-held.md)):

   ```
   spin workspace recover <workspace> --accept-data-loss
   ```

2. **An autoscaling group's machine**: give it back, or forget it now instead of waiting
   `host_forget_after_hours`. A connected one is cordoned and given back (its group terminates
   it and shrinks by one); one that is not connected is given back if the group still has it and
   forgotten at once:

   ```
   spin admin runner remove <runner-id>
   ```

3. **A runner registered with a token**: revoke it. That machine can never connect again, and
   the disks it held are marked to be re-keyed ([key-rotation](key-rotation.md)):

   ```
   spin admin runner revoke <runner-id>
   ```

   If workspaces are still on it and their owners will recover them, `--leave-workspaces`
   revokes anyway.
4. **The job fails**: `spin admin jobs list` shows its error; it needs the database. Run it
   again with `spin admin jobs run forget-gone-hosts`.

## Verify

The runner is gone from `spin admin runner list` (or shown revoked). [Admin → Audit](/admin/audit)
has `host_given_back`, `host_forgotten` or `runner_revoked` with your user. The fleet dashboard's
*Hosts: in the database, and connected* on [Admin → Metrics](/admin/metrics?dashboard=fleet)
shows the two counts close together.

## Escalate

If `remove` or `revoke` is refused for a runner no workspace names, or the cloud refuses to give
the machine back, stop and open an issue at https://github.com/spin-stack/spin/issues with the
runner id, the refusal's text, the trace id from [Admin → Traces](/admin/traces) and
`spin admin versions`. A runner that reconnects after a revoke is a security problem: follow
https://github.com/spin-stack/spin/security/policy.
