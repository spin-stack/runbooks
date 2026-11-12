---
id: workspace-oversized
title: A workspace held far more than it used
applies_to:
  alerts: []
  jobs: [oversized]
  events: [OVERSIZED]
  dashboards: []
severity: info
owner: operator
last_reviewed: 2026-11-12
---

## Symptoms

The owner sees on [Settings → Notifications](/settings/notifications), and is emailed only where
they turned it on (it is off by default):

> Workspace `<workspace>` holds far more than it used this week

with the most memory it used over the last seven days, its cores' use 95% of the time, what it
holds, and the smaller preset that holds what it used with room to spare. `spin workspace history
<workspace>` has the event, `OVERSIZED`. The daily `oversized` job writes it at most once a week
for each workspace.

## Impact

None on the workspace: nothing is resized on its own. Its host sets aside what it holds, so a
workspace that holds twice what it uses keeps another from being placed there, and its share of
the runners' bill (Admin → Costs) is what it holds.

## Diagnose

1. Is it right? The event says it from what the workspace's hosts read every heartbeat: the most
   its machine's processes used of memory each day, and the rate its cores were at or under 95%
   of the day. A workspace used hard once a month is not seen in one week.
2. Why that preset? The smallest enabled preset of its host class that holds its disk, less memory
   than it holds, no more cores, and half as much again as it used of both. None fitting is no
   event.
3. The job failing: Admin → Jobs shows `oversized`'s last runs and their error; while it fails
   nobody is told, and nothing else changes.

## Fix

- **One holding a grant above its preset**: its owner gives it back with
  `spin workspace release <workspace>`, the command the notice gives.
- **One at its preset's size**: a workspace of the smaller preset, made from a snapshot of this one
  (`spin workspace create --preset <preset> --from-snapshot <snapshot>`), keeps its disk.
- **An owner who does not want these**: they turn it off on Settings → Notifications.

## Verify

After a release, `spin workspace get <workspace>` shows the memory it holds; after a move to a
smaller preset, no `OVERSIZED` appears in the new workspace's history the next week.

## Escalate

If a workspace that ran out of memory in the same week is told it is oversized, open an issue at
https://github.com/spin-stack/spin/issues with the workspace id, its history and the output of
`spin admin versions`.
