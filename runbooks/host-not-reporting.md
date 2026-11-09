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

- **Active Hosts Are Not Connected** fires when the database counts more hosts `active` than are
  holding a connection to the control plane (`hosts{state="active"}` minus `hosts_connected`),
  for 5 minutes: a partition, or a host that died without being marked.
- **A Host Stopped Reporting** fires when a connected host (`host`) has sent no heartbeat for over
  a minute (`host_heartbeat_age_seconds > 60`), for 2 minutes.

Owners of workspaces on that host see them stop answering, or fail to start or stop. Their owners
may be told their workspace is unresponsive ([workspace-unresponsive](workspace-unresponsive.md)).

## Impact

Placement skips a host that is not connected, so the fleet has less room. Workspaces on it cannot
be started, stopped or reached through the control plane. What they wrote and had not published is
on that host alone. Believe nothing else the host reports while its heartbeat is old: its capacity
and its volumes' ages are what it last said.

## Diagnose

1. See the host as the control plane does:

   ```
   spin admin runner list
   spin admin health
   ```

   A host that is not connected shows `(not connected: <why>, <how long> ago)`; the reason is
   why its stream ended. The same on [Admin → Runners](/admin/runners), with the host's own page.
2. *Per host: heartbeat age* and *Hosts: in the database, and connected* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet).
3. Read what it logged before it went quiet:

   ```
   spin admin logs 'service_name:spin-runner AND severity_text:error' --since 1h
   ```

4. Look at the machine itself in the cloud's console: running, stopped, terminated, failing its
   status checks. On AWS, an autoscaling group may have replaced it.
5. If the machine runs, on it:

   ```
   systemctl status spin-runner.service
   journalctl -u spin-runner.service --since "1 hour ago"
   ```

   Runners have no SSH; on AWS they are reached through Session Manager.

## Fix

- **The runner process stopped**: `systemctl restart spin-runner.service` on the machine. A runner
  that restarts comes back as the same host and its workspaces are waited for, nothing lost.
- **The network between them** (the machine runs, the runner logs it cannot reach the control
  plane): fix the route, the security group or the DNS name. The host reconnects on its own.
- **The runner was revoked** (its log says its token is invalid): it will never connect again.
  Replace the machine ([lost-runner](lost-runner.md)).
- **The machine is gone for good**: follow [lost-runner](lost-runner.md). Its workspaces move to
  another host either by their owners (`spin workspace recover <workspace>`) or by the control
  plane on its own, past the installation's rehome threshold
  ([workspace-rehomed](workspace-rehomed.md)).
- **Cannot tell yet**: stop placing on it until you know: `spin admin runner drain <runner-id>`.

## Verify

`spin admin runner list` shows the host connected, or the replacement connected and the old one
gone. *Per host: heartbeat age* drops under a minute and both alerts resolve.

## Escalate

If the machine and its network are healthy and the runner still does not connect, or disconnects
again and again, open an issue at https://github.com/spin-stack/spin/issues with the host id, the
stream-end reason from `spin admin runner list`, the runner's journal, the control plane's log
lines that name the host, and `spin version` and `spin admin versions`.
