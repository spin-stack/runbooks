---
id: host-shell
title: Get a root shell on a runner
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

A runner misbehaves in a way nothing from outside explains (the metrics, `spin admin logs`, the
runner's own checks on its page under [Admin → Runners](/admin/runners)), and someone needs root
on it. Ask for a root shell; never get onto a runner another way.

## Impact

A root shell is for a while, with a reason, and costs the runner:

- The runner is emptied first: each workspace is suspended and resumed on another runner. With
  `--incident` it is not: the shell opens on the runner as it is, its workspaces still running.
- Every session is recorded by the control plane and kept in the installation's logs bucket,
  under Object Lock.
- Once the shell ends (its minutes are up, or it is closed) the runner is **retired**: revoked,
  every disk it held the key of marked to be re-keyed, and its machine given back to its group,
  which replaces it. A runner someone had root on runs no workspace again.

## Diagnose

Before asking, know which runner and why:

```
spin admin runner list
spin admin shell list
```

`spin admin shell list <runner-id>` shows that runner's shells: one may be asked for already.

## Fix

1. Ask, and open a session when it opens. The reason is shown to whoever approves and is kept:

   ```
   spin admin shell open <runner-id> --reason "<why>" --minutes 20
   ```

   For a runner whose running workspaces are part of the problem:

   ```
   spin admin shell open <runner-id> --reason "<why>" --incident
   ```

   `--minutes` is at most the installation's `host_shell_max_minutes` (30 by default). The same
   is offered on the runner's page, under *Root shells*.
2. Where the installation needs a second administrator (`host_shell_needs_approval`), another
   administrator approves it, by its id from `spin admin shell list`:

   ```
   spin admin shell approve <shell-id>
   ```

   Only then is the runner touched. To refuse or withdraw one not yet approved:

   ```
   spin admin shell close <shell-id>
   ```

3. Running `spin admin shell open <runner-id>` again opens another session on the same open
   shell. End the shell as soon as you are done; the runner is retired either way:

   ```
   spin admin shell close <shell-id>
   ```

4. Read what was done, each session in turn, as asciicast v2:

   ```
   spin admin shell recording <shell-id> > shell.cast
   asciinema play shell.cast
   ```

TODO: `host_shell_needs_approval` and `host_shell_max_minutes` cannot be set from Admin →
Settings, the CLI or the configuration file; only the admin API takes them ([spin#1048](https://github.com/spin-stack/spin/issues/1048)).

## Verify

`spin admin shell list` shows the shell `ended`, with RETIRED `true`. The runner is gone from
[Admin → Runners](/admin/runners) or revoked, and its group has started a replacement.
`spin admin volume rekeys` lists the disks it held while they are re-keyed
([key-rotation](key-rotation.md)). [Admin → Audit](/admin/audit) has `host_shell_requested`,
`host_shell_approved`, `host_shell_session_opened` and `host_shell_session_ended`.

## Escalate

If a shell will not open after approval (stuck emptying), or the runner is not retired after it
ends, stop and open an issue at https://github.com/spin-stack/spin/issues with the shell id, the
runner id, `spin admin shell list` output and `spin admin versions`. If the shell found signs of
compromise, follow https://github.com/spin-stack/spin/security/policy and do not post it in a
public issue.
