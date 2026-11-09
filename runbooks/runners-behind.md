---
id: runners-behind
title: Runners run another release than the control plane
applies_to:
  alerts: [runners-behind]
  jobs: []
  events: []
  dashboards: [fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**Runners Are On Another Release Than The Control Plane** fires when connected runners report a
release (`hosts_by_version`, by `version`) that the control plane does not run
(`build_info{job="spin-controlplane"}`), for 30 minutes. The `version` label is the runners'.

Usually seen after an update: a runner the update did not reach, or one that came back on an old
image or document.

## Impact

An installation runs one release in every half. A control plane serves a host of its own release
and refuses others, so a runner left behind takes no new work and its workspaces cannot be started
or stopped through it. The fleet has that much less room.

## Diagnose

1. See what runs where:

   ```
   spin admin versions
   ```

   It lists the control plane's release and, per host, the runner's release, its machine's and
   whether this control plane serves it (`SERVED`, with why not). *Releases: what each process
   runs, and the connected runners by release* on [Admin → Metrics](/admin/metrics?dashboard=fleet)
   shows the same over time.
2. For each host behind: is it a machine of an autoscaling group, or one registered with a token?
   `spin admin runner list` shows its class; the cloud's console shows its image.
3. A runner runs the release its machine's image carries. A machine started after the update from
   an old image (an old launch template, an old document) comes back behind.

## Fix

- **An autoscaling group's runner**: make sure the group now starts machines of the control
  plane's release (its launch template names the new image), then give the old one back to its
  group, which drains it, terminates it and starts a new one:

  ```
  spin admin runner remove <runner-id>
  ```

  A runner that is the only home of a workspace's disk is refused, naming them; their owners stop
  or move them first.
- **A runner registered with a token** (self-hosted): drain it with `--suspend` so its workspaces
  move, install the control plane's release on the machine, and join it again with a new token
  (`spin admin token create`). See [lost-runner](lost-runner.md) for taking the old host out.
- **The control plane is the one behind** (an update stopped half way): finish the update of the
  control plane.

## Verify

`spin admin versions` shows every host on the control plane's release and served. The alert
resolves.

## Escalate

If a runner of the right release is still reported on another, or the control plane refuses a
runner of its own release, open an issue at https://github.com/spin-stack/spin/issues with
`spin admin versions` and `spin version`.
