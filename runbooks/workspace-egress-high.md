---
id: workspace-egress-high
title: A workspace sent a lot today
applies_to:
  alerts: [workspace-egress-high]
  jobs: []
  events: []
  dashboards: [fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**A Workspace Sent A Lot Today** fires for each workspace whose machine has sent more than 100 GiB
since midnight UTC (`workspace_sent_today_bytes`, which exists only past that bound). The alert
names the workspace and its id. It resolves at midnight UTC, when the day's total starts again.

## Impact

Egress is billed by the byte: 100 GiB a day is a visible line on the bill. It can be legitimate
work (pushing images, syncing data), a mistake (a loop uploading), or abuse (the workspace used to
send traffic somewhere it should not).

## Diagnose

1. See it among others: *Workspaces that sent a lot today* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet).
2. See where it is sending. The workspace's owner runs this; it lists each destination with bytes
   sent and received, while the workspace runs:

   ```
   spin workspace egress <workspace>
   ```

   TODO: an administrator cannot read where a workspace they do not own sends ([spin#1045](https://github.com/spin-stack/spin/issues/1045))
3. See what it has cost this month:

   ```
   spin admin costs by workspace
   ```

4. Ask the owner what the workspace is doing.

## Fix

- **Legitimate**: nothing to do. If the cost matters, agree a limit with the owner.
- **A mistake, or more than it should**: limit what its network carries each way, in megabits a
  second. A running machine takes it at once; open connections stay open:

  ```
  spin workspace bandwidth <workspace-id> 50
  ```

  `0` lifts the limit. Only an administrator can set it.
- **Abuse**: quarantine the workspace, which stops it and keeps its disk as it is for review
  ([quarantine-workspace](quarantine-workspace.md)):

  ```
  spin workspace quarantine <workspace-id> --reason "sent 100 GiB to an unknown destination"
  ```

- **A network policy would have stopped it**: see [egress-denied](egress-denied.md) for how policies
  work, and [Admin → Network](/admin/network).

## Verify

The workspace's rate on *Workspaces that sent a lot today* flattens. The alert itself stays until
midnight UTC, because the day's total does not go down.

## Escalate

If the traffic looks like an attack from or on the installation, report it privately at
https://github.com/spin-stack/spin/security/advisories/new. If the counted bytes do not match what
`spin workspace egress` shows, open an issue at https://github.com/spin-stack/spin/issues with
both and `spin version`.
