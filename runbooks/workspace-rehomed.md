---
id: workspace-rehomed
title: A workspace was moved off a host that stopped answering
applies_to:
  alerts: []
  jobs: []
  events: [REHOMED]
  dashboards: [fleet, storage]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

The owner sees on [Settings → Notifications](/settings/notifications) (email is off by default
for this one):

> Workspace `<workspace>` was moved off a host that stopped answering

with how much work its host had not published, and `spin workspace start <workspace>`.

## Impact

The host stopped answering for longer than `rehome_after_seconds`, and what it had not
published was at or under `rehome_max_unpublished_bytes`, so the control plane moved the
workspace to another host on its own, as the installation's settings allow. The owner lost the
writes that host had not published (the amount is in the notice). The workspace is stopped on
its new host until they start it.

Many such notices at once mean a host, or several, went away.

## Diagnose

1. Which host went silent? [Admin → Audit](/admin/audit) has a `workspace_recovered` entry per
   move, with the bytes given up; *Per host: heartbeat age* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet) shows which runner went quiet.
2. Why? Follow [host-not-reporting](host-not-reporting.md).
3. Was the loss larger than expected? *What losing a host would cost* on
   [Admin → Metrics](/admin/metrics?dashboard=storage) shows how far behind each host's
   publishing was. A host that publishes far behind its RPO loses more when it goes.

## Fix

- The owner starts the workspace: `spin workspace start <workspace>`.
- Deal with the host that went: bring it back, or retire it ([lost-runner](lost-runner.md)).
- If the loss is more than the installation is willing to give up without asking, lower
  `rehome_max_unpublished_bytes` (or unset `rehome_after_seconds` to move nothing on its own)
  in the installation's configuration file and apply the change on
  [Admin → Settings](/admin/settings). Workspaces over the limit then wait for their owner
  instead ([workspace-held](workspace-held.md)).

## Verify

The workspace runs on its new host. The silent host is back, or gone from
[Admin → Runners](/admin/runners).

## Escalate

If workspaces are moved while their host is connected and heartbeating, stop: that is a defect.
Open an issue at https://github.com/spin-stack/spin/issues with the audit entries, the runner's
id and heartbeat age at the time, and `spin admin versions`.
