---
id: tls-ask-unreachable
title: The edge cannot ask the control plane about certificates
applies_to:
  alerts: [tls-ask-unreachable]
  jobs: []
  events: []
  dashboards: [fleet]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**The Edge Cannot Ask About Certificates** fires as soon as the proxy failed to ask the control
plane whether a name may have a certificate, in the last 10 minutes
(`tls_ask_total{answer="control_plane_unreachable"}`).

The proxy asks before it gets a certificate for a workspace's name it has none for. When it cannot
ask, it hands out none: people opening a new workspace, or a newly exposed port, get a TLS error in
the browser.

## Impact

No new workspace name gets a certificate on that edge until the proxy can ask again. Names that
already have one keep working.

## Diagnose

1. *Certificates asked for* and *The edge's questions, as the control plane answered* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet).
2. Is the control plane up? If [api-errors](api-errors.md) is firing too, start there.
3. Read why the proxy could not ask. On the proxy's machine, its journal has a line for each
   failure, with the name it was asked for:

   ```
   systemctl status spin-proxy.service
   journalctl -u spin-proxy.service --since "1 hour ago" | grep "could not ask the control plane"
   ```

   The error says which: a connection refused or timed out (network), or an unauthenticated
   answer (the proxy's token).

## Fix

- **The control plane is down**: bring it back (`systemctl status spin-controlplane.service` on its
  machine); the proxy asks again on the next handshake.
- **The network between proxy and control plane**: fix the route, security group or DNS name.
- **The proxy's token is refused**: an AWS installation's proxy proves itself and renews its token
  every half hour; check its journal for why it could not. A proxy given a token by hand needs a
  new one, made on the control plane's machine:

  ```
  spin-controlplane proxy-token --name <proxy-name> --expires 720h
  ```

  Install it where the proxy's document says it is kept (`install.join.token_at`), readable by root
  alone; the proxy reads the file again when it changes.

## Verify

No new `control_plane_unreachable` answers in 10 minutes and the alert resolves. A new workspace
opens in the browser without a certificate error.

## Escalate

If the proxy cannot ask while the control plane answers everything else, open an issue at
https://github.com/spin-stack/spin/issues with the proxy's log lines, the control plane's log lines
for the same minutes, and `spin version`.
