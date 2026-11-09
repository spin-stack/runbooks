---
id: job-failing
title: A control plane job keeps failing, or has not succeeded in two days
applies_to:
  alerts: [job-runs-failing, job-not-succeeding]
  jobs: [sessions, expired-data, workspace-trash, workspace-grants, prebuilds, forge-revocations, forwards, stale-workspaces, notifications, shares, usage, collect-releases, storage-gc, bill]
  events: []
  dashboards: [fleet]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

- **A Job Keeps Failing** fires when one job (`job`) failed more than 3 runs in the last hour
  (`job_run_duration_seconds_count{outcome="failed"}`).
- **A Job Has Not Succeeded In Two Days** fires when a job's last success is more than 48 hours
  old (`job_last_success_timestamp_seconds`), for 15 minutes. That is a job failing every run, or
  one that stopped running at all, which a failure count cannot show.

Some jobs have their own runbook: `supply-vulnerabilities`
([supply-vulnerabilities](supply-vulnerabilities.md)), `identity-audit`
([identity-audit-failing](identity-audit-failing.md)), `silent-supervisors`
([workspace-unresponsive](workspace-unresponsive.md)), `forget-gone-hosts`
([lost-runner](lost-runner.md)), `workspace-rekeys` ([key-rotation](key-rotation.md)),
`vscode-servers` and `agent-clis` ([third-party-terms](third-party-terms.md)). The rest are below.

## Impact

A job repairs or cleans something on a schedule. While it fails, that work stays undone; each job
below says what. Nothing breaks at once, but it piles up.

## Diagnose

1. Open the job on [Admin → Jobs](/admin/jobs): its last runs, each failure's error, what its last
   good run did, and when it runs next. Or:

   ```
   spin admin jobs list
   ```

2. Read the error. Most failures are one of three things the job reaches: the database, the
   object store, or something outside (a forge, an email server, the cloud's bill).
3. A job not run at all (no recent runs, `NEXT` in the past): the control plane is not running its
   jobs. `spin admin health` names every job whose last run failed, with its error, but not one
   that stopped running.
4. The fleet dashboard has *What the jobs did, and the runs that failed* and *Since each job last
   succeeded*: [Admin → Metrics](/admin/metrics?dashboard=fleet).

### The jobs

- **`sessions`**: deletes expired sign-in sessions. While it fails they stay in the database; an
  expired one signs nobody in. Check the database.
- **`expired-data`**: deletes expired API keys, sign-in, device and registration codes, forge
  states, and audit entries past their retention. While it fails they pile up; an expired one
  admits nobody. Check the database.
- **`workspace-trash`**: erases workspaces whose days in the trash are over, with their disks, and
  gives back what they held of their runners. While it fails, deleted workspaces keep their disks
  and their runners' room ([kept-volumes-growing](kept-volumes-growing.md)). Check the runners
  holding those disks and the object store.
- **`workspace-grants`**: gives back the cores and memory a workspace was granted for a time, once
  the time is up. While it fails, workspaces keep them and their runners offer less. Check the
  database and the runners.
- **`prebuilds`**: watches each project's prebuild refs on its forge, builds when one moves, and
  gives back old prebuilds. While it fails, new workspaces start from an older prebuild. Check the
  forge's connection on [Admin → Integrations](/admin/integrations) and the project's prebuilds
  (`spin project prebuilds <slug>`).
- **`forge-revocations`**: revokes at the forge the tokens people and apps gave up, and finishes
  deleting an app. While it fails, a token given up still works at the forge. Check that the forge
  is reachable and the app's credentials on [Admin → Integrations](/admin/integrations). This one
  matters for security: escalate if it fails for more than a day.
- **`forwards`**: asks again whether each open tunnel into a workspace is still allowed, and has its
  runner close those that are not. While it fails, a tunnel outlives a sign-out, a revoked key or
  a disabled account. Check the runners; this one matters for security too.
- **`stale-workspaces`**: fails workspaces stuck creating, starting or stopping past the deadline.
  While it fails, such a workspace stays stuck and its owner cannot leave it. Check the database.
- **`notifications`**: sends each person what happened to their workspaces, through the channels
  the installation has. While it fails nobody is told; what is due waits for the next run. Check
  the email settings on [Admin → Integrations](/admin/integrations).
- **`shares`**: forgets a snapshot's share outside its team once nothing may clone from it, which
  destroys the share's key. While it fails, keys of ended shares are kept. Check the object store
  and the key service.
- **`usage`**: records what each workspace held of its runner since the last sample, which the
  bill is shared out by. While it fails, that time is counted to nobody. Check the database.
- **`collect-releases`**: removes extension releases nothing runs and that are not their
  extension's newest. While it fails they stay in the store and are paid for. Check the object
  store; `spin admin storage list` shows what is kept.
- **`storage-gc`**: deletes objects in the bucket nothing reaches any more, and reads what the
  bucket holds. While it fails they stay, the bucket costs what they take, and
  [Admin → Storage](/admin/storage) shows an older reading. Check the object store's permissions.
- **`bill`**: reads what the cloud bills the installation, a few times a day. While it fails,
  [Admin → Costs](/admin/costs) shows the last reading. Check the control plane's permission to
  read the bill and the installation's `bill_scope`.

## Fix

1. Fix what the error names: the database ([database-slow](database-slow.md)), the object store,
   the forge or the email server.
2. Run the job again without waiting for its schedule, from its page on
   [Admin → Jobs](/admin/jobs) (**Run now**), or:

   ```
   spin admin jobs run <job>
   ```

   It is not waited for; `spin admin jobs list` says how it went. A job already running is
   refused.
3. A job that does not run at all: restart the control plane on its machine,
   `systemctl restart spin-controlplane.service`.

## Verify

The job's next run on [Admin → Jobs](/admin/jobs) succeeds. **A Job Keeps Failing** resolves once
the last hour has no more than 3 failures; **A Job Has Not Succeeded In Two Days** resolves with
the first success.

## Escalate

If the error is not about something the job reaches (a panic, a database error that names a
column, a failure that repeats with everything else healthy), it is a defect. Open an issue at
https://github.com/spin-stack/spin/issues with the job's name, the failed run's error from Admin →
Jobs, the control plane's log lines around it, and `spin version`:

```
spin admin logs 'service_name:spin-controlplane AND severity_text:error' --since 24h
```
