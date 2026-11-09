---
id: object-store-requests-runaway
title: A host keeps asking the object store
applies_to:
  alerts: [objectstore-requests-runaway]
  jobs: []
  events: []
  dashboards: [storage]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**A Host Keeps Asking The Object Store** fires when one runner (`instance`, its host id) makes
more than 10 requests a second to the object store, averaged over 10 minutes
(`objectstore_request_duration_seconds_count{job="spin-runner"}`), for 15 minutes.

A busy host makes about 1.4 a second at most and none at rest. The one storm seen was a restore
retried every cycle: over 25 a second from one host.

## Impact

Every request is on the cloud bill. Nothing is lost, but a storm left alone costs money every
hour, and is usually a symptom of a volume the host cannot serve.

## Diagnose

1. Open [Admin → Metrics](/admin/metrics?dashboard=storage):
   - *Requests to the cloud, by service and operation* and *How long the store takes to answer, by
     request* say which operation (`op`) the host repeats.
   - *Is the store answering at all*: requests that fail and are retried.
   - *A host's reconciliation: how long a cycle takes, and cycles that failed*: a host that fails
     every cycle and tries again.
2. Read that host's warnings in the runners' log, and look for one volume named over and over:

   ```
   spin admin logs 'service_name:spin-runner AND severity_text:warn' --since 1h
   ```

3. A volume restore that keeps failing is the known cause: check for
   [volume-not-served](volume-not-served.md) firing too.

## Fix

- **One volume retried** (a restore, a publish): fix that volume
  ([volume-not-served](volume-not-served.md)). Once it is served or moved, the requests stop.
- **The store refusing** (access denied, throttling): the host's credentials or the bucket's
  policy changed. Check the host's role in the cloud and *Credentials hosts asked for, and what the
  store's STS minted* on the storage dashboard.
- **Nothing explains it**: stop the host taking new work (`spin admin runner drain <runner-id>`)
  and look at it on a [host-shell](host-shell.md). Draining with `--suspend` moves its running
  workspaces elsewhere.

## Verify

The host's request rate on the storage dashboard drops back under one or two a second and the
alert resolves. Check the next day's cost on [Admin → Costs](/admin/costs).

## Escalate

A host that repeats a request with no failing volume or store error to explain it is a defect.
Open an issue at https://github.com/spin-stack/spin/issues with the host id, the operation, the
dashboard panels and the runner's log lines, and `spin version`.
