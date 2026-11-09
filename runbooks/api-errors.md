---
id: api-errors
title: The API fails more than one call in twenty
applies_to:
  alerts: [api-error-rate]
  jobs: []
  events: []
  dashboards: [overview, spin-api-performance, spin-db-performance]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**High API Error Rate** fires when, over the last 5 minutes, more than 5% of the control plane's
API calls ended with any status other than `OK`, and it stays so for 5 minutes. The rule divides
`rpc_server_call_duration_seconds_count{rpc_response_status_code!="OK"}` by all calls; the
description gives the share.

People see pages that fail to load, `spin` commands that print an error, editors that cannot
connect.

## Impact

Everyone who uses the installation: the dashboard, the `spin` CLI, editor extensions and agents
all go through this API. Workspaces already running keep running, but nobody can start, stop or
open one while the calls that do it fail.

## Diagnose

1. Open [Admin → Metrics](/admin/metrics?dashboard=spin-api-performance) (API Performance).
   *Failed requests by status* says what kind of failure; *Methods that fail most* says where.
   - `UNAUTHENTICATED`, `PERMISSION_DENIED`: callers turned away. Many at once is a client with a
     dead credential, or a brute force: see [auth-failures-looping](auth-failures-looping.md).
   - `RESOURCE_EXHAUSTED`: the rate limit refusing: see [rate-limited](rate-limited.md).
   - `INTERNAL`, `UNAVAILABLE`, `DEADLINE_EXCEEDED`: the control plane's own faults.
2. Take a failing method to [Admin → Traces](/admin/traces): each failed call is kept whole, with
   the database statements and calls to runners under it.
3. Read what the control plane logged:

   ```
   spin admin logs 'service_name:spin-controlplane AND severity_text:error' --since 1h
   ```

4. Ask what the installation knows is broken:

   ```
   spin admin health
   ```

   A failed database, log store or collector shows here first. A whole service failing at once is
   usually the database: [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance).

## Fix

- **Database down or saturated**: follow [database-slow](database-slow.md). On the control
  plane's machine, `systemctl status spin-postgres.service` says whether it runs.
- **The control plane itself unwell** (memory, goroutines climbing): follow
  [control-plane-resources](control-plane-resources.md).
- **Runners not answering** (errors on start, stop, connect): follow
  [host-not-reporting](host-not-reporting.md).
- **One client failing in a loop** (one method, `UNAUTHENTICATED`): it is not the installation;
  find and stop the client ([auth-failures-looping](auth-failures-looping.md)).
- **A release just went out**: a CLI or runner of another release is refused by the control
  plane. `spin admin versions` shows what runs where; people update their CLI with `spin update`.

## Verify

The *Error rate* stat on [Admin → Metrics](/admin/metrics?dashboard=overview) falls under 5% and
the alert resolves after its next evaluation. `spin admin health` exits 0.

## Escalate

If the errors are `INTERNAL` and nothing above explains them, stop and open an issue at
https://github.com/spin-stack/spin/issues with: the alert's value, the failing methods, two or
three trace ids from Admin → Traces, the matching log lines, and `spin version` and
`spin admin versions`.
