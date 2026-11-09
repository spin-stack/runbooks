---
id: start-failures-unclassified
title: Workspace starts failed where the runner named no part
applies_to:
  alerts: [start-failure-unclassified]
  jobs: []
  events: []
  dashboards: [fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**Unclassified Start Failures** fires as soon as any workspace became `FAILED` while creating or
starting, in the last hour, with no part named for the failure
(`workspace_failures_total{phase=~"creating|starting",failure="unclassified"}`).

The owner sees a failed workspace with a summary of what went wrong, but without the advice a
known failure gets (which part failed, what to try).

## Impact

The owner's workspace failed to start. Every other failure names the part that failed (the
volume, the guest, resources, no host...), which is what tells the owner and you what to do; an
unclassified one is a place in Spinbox where that name is missing. That is a defect to report,
even when the workspace starts the second time.

## Diagnose

1. See how many, and whether they come with known failures: *Failures by phase and part* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet).
2. Find the failures in the logs, from the control plane and the runners:

   ```
   spin admin logs 'severity_text:error' --since 1h
   ```

3. The workspace's owner sees what failed, the host's detail and the trace id:

   ```
   spin workspace get <workspace>
   ```

   Open that trace in [Admin → Traces](/admin/traces): the start is one trace from the request to
   the guest answering, and the failing span says where.

## Fix

- Find the cause in the trace and the logs and fix it as for any start failure: the host
  ([host-not-reporting](host-not-reporting.md)), the volume
  ([volume-not-served](volume-not-served.md)), the image or the network.
- The owner starts the workspace again (`spin workspace start <workspace>`), or deletes it if it
  never ran.
- Report the missing name either way (Escalate).

## Verify

New failures on *Failures by phase and part* name a part, and the alert resolves after an hour
with no unclassified failure.

## Escalate

Always open an issue at https://github.com/spin-stack/spin/issues, so the failure gets its name:
bring the failure summary and reason the workspace shows, the trace id from Admin → Traces, the
log lines, and `spin version`.
