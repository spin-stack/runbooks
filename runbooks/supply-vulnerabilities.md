---
id: supply-vulnerabilities
title: A vulnerability is known in what this installation runs
applies_to:
  alerts: [supply-vulnerabilities]
  jobs: [supply-vulnerabilities]
  events: []
  dashboards: [fleet]
severity: warning
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

- **A Vulnerability Is Known In What This Installation Runs** fires when the last successful run of
  the `supply-vulnerabilities` job counted one or more vulnerabilities
  (`job_last_success_count{job="supply-vulnerabilities",count="vulnerabilities"} > 0`), for an
  hour.
- The job itself failing shows on [Admin → Jobs](/admin/jobs?job=supply-vulnerabilities) and fires
  the generic job alerts ([job-failing](job-failing.md)).

Once a day the job reads the software bill of materials (SBOM) of every package the release's image
carries and asks OSV (`api.osv.dev`) which vulnerabilities are known in those versions.

## Impact

A vulnerability published after the release was built, which the release's own checks could not
have caught. Whether it matters depends on the package and on whether Spinbox uses the affected
part. While the job fails, a new vulnerability goes unreported.

## Diagnose

1. See which ones. The control plane logs each finding, with the vulnerability's id and the
   package:

   ```
   spin admin logs 'service_name:spin-controlplane AND body.message:supplyvuln' --since 48h
   ```

2. The job's page on [Admin → Jobs](/admin/jobs?job=supply-vulnerabilities) shows what its last run
   counted: SBOMs, components and vulnerabilities. So does *What each job's last successful run
   found* on [Admin → Metrics](/admin/metrics?dashboard=fleet).
3. Read each vulnerability at OSV (`https://osv.dev/vulnerability/` and its id): its severity, the
   fixed version, and the affected functions.
4. Check whether a newer Spinbox release carries the fixed version: its release notes, or a
   finding that is gone after updating.
5. If the job fails instead: the run's error on Admin → Jobs. Usually the control plane's machine
   cannot reach `api.osv.dev`. A machine with no SBOM (one not installed from a release image)
   logs `no SBOM on this machine` and counts nothing.

## Fix

- **A newer release carries the fix**: update the installation to it.
- **No release carries it yet**: report it (Escalate). Until a release has it, judge the
  exposure from the advisory: what an attacker needs, and whether that part of the package is
  reachable in your installation.
- **The job fails to reach OSV**: allow the control plane's machine HTTPS to `api.osv.dev`, then
  run it again:

  ```
  spin admin jobs run supply-vulnerabilities
  ```

## Verify

After an update, the job's next run (it also runs when the control plane starts) counts 0
vulnerabilities, and the alert resolves after an hour.

## Escalate

Tell the maintainers about a known vulnerability no release fixes yet. If it is exploitable in
Spinbox, report it privately at https://github.com/spin-stack/spin/security/advisories/new with the
id, the package, why you think it is reachable, and `spin version`. Otherwise open an issue at
https://github.com/spin-stack/spin/issues asking for the bump.
