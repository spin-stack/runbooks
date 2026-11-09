---
id: quarantine-workspace
title: Quarantine a workspace, read its denials, release it
applies_to:
  alerts: []
  jobs: []
  events: []
  dashboards: []
severity: info
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

A workspace is held in quarantine, or should be:

- **The control plane held it.** The workspace's guard refused something an agent in it tried
  (a denial). The supervisor reports each denial, and the first one stops the workspace by force
  and holds it. Its owner sees "this workspace is held in quarantine until an installation
  administrator releases it" on every start, move or delete.
- **An administrator decides to hold it**: it is suspected of abuse or compromise, and its disk
  must be kept as it is.

## Impact

While held, the workspace cannot be started, moved or deleted by its owner, and its disk is kept
as it was. Nothing else of the owner's is affected. Releasing it puts it back to stopped; the
record of what it attempted stays.

## Diagnose

All of these are an administrator's and take the workspace's id (not its name): it is on the
workspace's page address (`/workspaces/<id>`), in the owner's `spin workspace list -o json`, and
on its entries in [Admin → Audit](/admin/audit).

What the guard refused, newest first: target, which user and process, and why:

```
spin workspace denials <id> --limit 100
```

[Admin → Audit](/admin/audit) shows `workspace_quarantined` and `workspace_unquarantined`, with
who and the reason.

## Fix

- **Hold a workspace.** A running one is stopped first, forced, and held once its runner says it
  stopped. The reason is kept on the audit entry:

  ```
  spin workspace quarantine <id> --reason "<why>"
  ```

- **If its disk key may be taken** (the runner it ran on is suspect), re-key the disk while it is
  held: see [key-rotation](key-rotation.md).
- **Release it** once the denials are understood (an agent tool that needed something the
  policy refuses, not an escape):

  ```
  spin workspace unquarantine <id>
  ```

  It goes back to stopped, with its disk as it was. A quarantine whose stop ended failed is
  released as failed, and no longer held.
- **The denials show an attempt to escape**: keep it held, and escalate as a security problem.

## Verify

`spin workspace denials <id>` still answers (the record stays). After a release the owner can
start the workspace; the audit has the `workspace_unquarantined` entry.

## Escalate

Denials that show an attempt to get past the guard, or a workspace that keeps running after
`quarantine`, are security problems: report them privately through
https://github.com/spin-stack/spin/security/policy, with the workspace id, the denials, the
audit entries and `spin admin versions`. Never paste denials into a public issue: they can name
what the agent was after.
