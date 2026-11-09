---
id: api-latency
title: API calls are slow
applies_to:
  alerts: [api-latency]
  jobs: []
  events: []
  dashboards: [overview, spin-api-performance, spin-db-performance]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**High API Latency (P99)** fires when the 99th percentile of every API call's duration over the
last 5 minutes (`rpc_server_call_duration_seconds`, all methods together) is above 1 second, for 5
minutes.

People see pages that hang before they load, `spin` commands that take seconds, a workspace that
takes long to open.

## Impact

Everyone using the installation, a little; and the work that waits on the API, such as starting
and stopping workspaces, more. Nothing is lost, but slow calls that time out become errors
([api-errors](api-errors.md)).

## Diagnose

1. Open [Admin → Metrics](/admin/metrics?dashboard=spin-api-performance). *p95 latency of the 10
   slowest methods* names the slow ones. One method slow is that method's work; every method
   slow at once is something they share, usually the database.
2. For a slow method, open its calls in [Admin → Traces](/admin/traces). The spans under a call say
   where the time went: a database statement, a call to a runner, the object store.
3. Check the database: [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance). A pool near
   its maximum, or *Waiting for a connection* climbing, is [database-slow](database-slow.md).
4. Check the control plane's own process: *Memory each service's Go runtime uses* on
   [Admin → Metrics](/admin/metrics?dashboard=overview). A process near its limits is
   [control-plane-resources](control-plane-resources.md).

## Fix

- **The database**: follow [database-slow](database-slow.md).
- **Calls to runners** (a start, a stop, a connect) waiting on a host: follow
  [host-not-reporting](host-not-reporting.md) for the host the trace names.
- **The object store** answering slowly: the storage dashboard's *How long the store takes to
  answer, by request* ([Admin → Metrics](/admin/metrics?dashboard=storage)); the store's own
  status page is the cloud's.
- **One method slow by design** (a long listing, a search over a month of logs): nothing to fix
  in the installation; the alert resolves when those calls stop.

## Verify

*API p99 latency* on [Admin → Metrics](/admin/metrics?dashboard=overview) is back under one second,
and the alert resolves.

## Escalate

If one method is slow every time and its trace shows the time inside the control plane (not the
database, a runner or the store), open an issue at https://github.com/spin-stack/spin/issues with
the method, two or three trace ids from Admin → Traces, and `spin version`.
