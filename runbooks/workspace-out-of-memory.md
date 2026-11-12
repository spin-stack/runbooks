---
id: workspace-out-of-memory
title: A workspace ran out of memory
applies_to:
  alerts: []
  jobs: []
  events: [OUT_OF_MEMORY]
  dashboards: [fleet]
severity: info
owner: operator
last_reviewed: 2026-11-12
---

## Symptoms

The owner is emailed (on by default) and sees on [Settings → Notifications](/settings/notifications)
one of:

> Workspace `<workspace>` ran out of memory

> Workspace `<workspace>` is out of memory at its ceiling

> Workspace `<workspace>` ran out of memory and its host stopped it

with what it held, its own memory and how far it may be lent, and the command
`spin workspace grant <workspace> --memory <MB>`. `spin workspace history <workspace>` has the
event, `OUT_OF_MEMORY`, with the processes the guest's kernel killed. One is recorded for each of
the three causes at most every ten minutes, however often it runs out in between.

## Impact

- **Its kernel killed processes** (cause `guest`): a build, a language server or an editor's
  server ended. The workspace runs on; what was killed is gone until it is started again.
- **Short at its ceiling** (cause `ceiling`): its host has lent it all it may have for a minute
  and it is still short. A kill is likely next.
- **Its host killed the machine** (cause `host`): the whole workspace stopped, as after a power
  cut. Its disk is kept.

Nothing is resized on its own: the event is the owner's to act on.

## Diagnose

1. Which cause? `spin workspace history <workspace>` shows it, with what was killed.
2. What did it use? The workspace's page in the dashboard shows its last hour of memory against
   what it held, while its host still has the hour.
3. Many workspaces on one runner killed by their host at once is the runner's memory, not the
   workspaces': *Containment and DNS* on [Admin → Metrics](/admin/metrics?dashboard=fleet) counts
   the host's kills.

## Fix

- **One workspace that needs more**: its owner runs the command in the notice, or picks a larger
  preset where the notice says it is granted all its preset allows.
- **A runner killing machines**: follow [host-not-reporting](host-not-reporting.md) for the runner,
  and see what it holds on [Admin → Runners](/admin/runners).

## Verify

After a grant, `spin workspace get <workspace>` shows the memory it holds, and no new
`OUT_OF_MEMORY` appears in its history under the same work.

## Escalate

If a workspace with memory to spare is reported out of memory, open an issue at
https://github.com/spin-stack/spin/issues with the workspace id, its history, the output of
`spin workspace status <workspace>` and `spin admin versions`.
