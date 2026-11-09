# Contributing

A runbook is what an operator of a Spinbox installation reads when something is wrong: an alert
fired, a job keeps failing, a person was told their workspace is waiting for them, or someone
asks for a procedure (a host shell, a restore). It says what is seen, who it hurts, how to find
out, what to do, how to know it worked and whom to call. Short, plain English, the commands as
they are typed.

This repository is public. **Never** write a secret, a hostname or account id of an installation,
a person's name or email, or anything a customer did. Write `<your domain>`, `<workspace>`,
`<host>` in code, for any installation.

## A runbook

`runbooks/<id>.md`, named for its id:

```markdown
---
id: host-not-reporting
title: A host stopped reporting
applies_to:
  alerts: [hosts-disconnected, host-heartbeat-stale]
  jobs: []
  events: []
  dashboards: [fleet]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

What fires, what a person sees.

## Impact

Who and what it hurts, and what is lost if nothing is done.

## Diagnose

The commands (`spin ...`, `controlplane ...`) and the pages (Admin → Runners) that say why.

## Fix

What to do, in order, for each cause the diagnosis finds.

## Verify

How to know it worked: the alert resolves, the page reads healthy.

## Escalate

When to stop and whom to call, with what to bring.
```

- **`id`** is stable: spin's alert rules and pages link it (`/admin/runbooks/<id>`), and spin's
  tests fail on a link to one that does not exist. Renaming one is a change in spin too.
- **`applies_to`** is what sends an operator here, and what the app links from: `alerts` are
  Grafana alert rule uids, `jobs` control plane job names (Admin → Jobs), `events` the
  notices a workspace's owner gets (`HELD`, `STOP_BACKUP_FAILED`, `UNRESPONSIVE`, `REHOMED`),
  `dashboards` Grafana dashboard uids. A procedure nobody is paged for applies to nothing and is
  `severity: info`.
- **`severity`**: `critical` (someone is losing work or access now), `warning`, `info`.
- **`owner`**: the role that keeps it true: `operator`, `security`.
- **`last_reviewed`**: the date someone last held it against the release it describes.

## What the app renders

Markdown, and nothing else: raw HTML is refused here and dropped by the app. Links are to
another runbook (`[held](workspace-held.md)`, which stays in the app), to a path in the app
(`/admin/jobs`), or `https://`. No images.

Ground every command and page in what spin has: a flag that does not exist is worse than no
runbook. Where spin has no way yet, write `TODO` and link the issue that will add one.

## Check

```
go tool task check
```

Fails on a field missing, an id that is not its file's name or is used twice, a section missing
or out of order, raw HTML, a link to a runbook that does not exist, an AWS account id, a private
key or an installation's hostname. CI runs the same (Gate).

## How a change reaches an installation

spin pins this repository at a commit (`versions.yaml`, `runbooks`). A merged change here reaches
the app when spin is bumped (`task bump NAME=runbooks`, `task runbooks`) and released. The app's
**Open in repo** link is the file at the commit the release pins: propose the change on `main`.
