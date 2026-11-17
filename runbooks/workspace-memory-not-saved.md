---
id: workspace-memory-not-saved
title: A workspace's memory was not saved before its host was taken away
applies_to:
  alerts: []
  jobs: []
  events: [MEMORY_NOT_SAVED]
  dashboards: [fleet, storage]
severity: warning
owner: operator
last_reviewed: 2026-11-17
---

## Symptoms

The owner is emailed, and sees on [Settings → Notifications](/settings/notifications):

> Workspace `<workspace>` lost its running processes: its host was taken away

with the publish its disk is safe up to, why its memory was not saved, and
`spin workspace start <workspace>`.

## Impact

The cloud announced it was taking the host away (a reclaimed spot instance), and the host's
evacuation could not suspend this workspace before the notice's deadline. What ran in it is lost;
its disk is safe at least up to the publish the notice names, and the host may have published more
before it went. Once the deadline passes, the workspace is moved to another host when the move
loses nothing ([workspace-rehomed](workspace-rehomed.md)), and otherwise waits for its owner
([workspace-held](workspace-held.md)).

Several at once from one host mean the host held more memory than it could save in its notice.

## Diagnose

1. Which host was taken, and when? [Admin → Audit](/admin/audit) has `host_going_away` with the
   grace the cloud gave, and `host_evacuated` with what was suspended and what failed.
2. Why was this one not saved? The notice's *Why* says: the deadline came before its turn, or its
   suspend was still running when the deadline came, or the host refused it.
3. How fast does the host save memory? *Checkpoint memory: how long a suspend's upload and a
   resume's download take* on [Admin → Metrics](/admin/metrics?dashboard=storage); a host that
   holds more memory than it saves in its grace loses some of it each time it is taken.

## Fix

- The owner starts the workspace once it is moved: `spin workspace start <workspace>`; or, if it
  waits for them, moves it as [workspace-held](workspace-held.md) says.
- If it happens often, put less memory on spot hosts: fewer or smaller workspaces per host, or
  on-demand hosts for the workspaces whose memory must survive.

## Verify

The workspace runs on another host, from the publish the notice named or a later one.

## Escalate

If a workspace's memory was not saved on a host that was not taken away (no `host_going_away` in
the audit), stop: that is a defect. Open an issue at https://github.com/spin-stack/spin/issues
with the audit entries, the runner's id, and `spin admin versions`.
