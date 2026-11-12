---
id: database-slow
title: The control plane's database is slow or its pool is full
applies_to:
  alerts: [db-pool-saturation, db-query-latency]
  jobs: []
  events: []
  dashboards: [spin-db-performance, overview]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

- **Database Connection Pool Saturation** fires when more than 80% of the control plane's
  connection pool is in use (`db_client_connections_usage{state="used"}` over
  `db_client_connection_max`) for 5 minutes.
- **High Database Query Latency** fires when the 99th percentile of statement durations over 5
  minutes (`db_client_operation_duration_seconds`) is above 0.5 seconds, for 5 minutes.

People see a slow dashboard and slow commands; [api-latency](api-latency.md) often fires with it.

## Impact

PostgreSQL holds the whole installation and every API call waits on it, as does the work that
starts and stops workspaces. A full pool makes every caller queue for a connection; long enough,
calls time out and fail ([api-errors](api-errors.md)).

## Diagnose

1. Open [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance).
   - *Statement latency by operation* and *Statements a second by operation*: is one kind of
     statement slow, or are there simply many more?
   - *The pool over time* and *Waiting for a connection*: how full, and how long callers wait. The
     pool holds at most `database.max_conns` connections (*Pool maximum*): 10 unless the control
     plane's document says otherwise, and never more than 20.
2. Find the slow callers in [Admin → Traces](/admin/traces): their spans carry each statement and
   how long it took.
3. On the control plane's machine, check the database itself:

   ```
   systemctl status spin-postgres.service
   journalctl -u spin-postgres.service --since "1 hour ago"
   ```

4. Check the machine: CPU, memory and disk I/O of the control plane's machine (the cloud's console,
   or `top` and `iostat` on the machine). The database shares it with the control plane and its
   stores.

## Fix

- **One statement slow** (the same operation in every slow trace): that is a defect in Spinbox,
  not something to tune; escalate with the trace.
- **A burst of callers** (statements a second climbing with requests a second): find who on
  [Admin → Metrics](/admin/metrics?dashboard=spin-api-performance), *The 10 busiest methods*. A
  client looping on one method is stopped at the client; a limit set too high is lowered on
  [Admin → Security](/admin/security) (*Rate limits and sessions*).
- **The machine is out of CPU, memory or I/O**: give the control plane's machine more, which is a
  change of its instance type and a replacement of the machine. Read
  [database-restore](database-restore.md) before replacing it: the database moves by handover.
- **Long-held connections** with few statements: restart the control plane, which closes its pool:
  `systemctl restart spin-controlplane.service`. Brief: the API is down while it starts.
- **The pool is full with the machine idle** (callers waiting, CPU and I/O low): raise
  `database.max_conns` (and `database.min_conns` if it starts cold) in the control plane's
  document, which the installation's Terraform writes, then restart the control plane. Above 20
  it refuses to start: the machine's PostgreSQL has 30 connections, and the rest are its one-shot
  commands', its backup's and an operator's.

## Verify

On [Admin → Metrics](/admin/metrics?dashboard=spin-db-performance), connections in use drop below
four fifths of the pool maximum and statement p99 below 0.5 s. Both alerts resolve.

## Escalate

Open an issue at https://github.com/spin-stack/spin/issues when one statement is slow on its own,
or the pool fills with no burst of callers to explain it. Bring the dashboard's panels for the
window, two or three trace ids from Admin → Traces, and `spin version`.
