---
id: identity-audit-failing
title: The installation's identity key was opened more than granted, or its audit fails
applies_to:
  alerts: []
  jobs: [identity-audit]
  events: []
  dashboards: []
severity: critical
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

The `identity-audit` job failed on [Admin → Jobs](/admin/jobs?job=identity-audit), and the generic
job alerts fire for it ([job-failing](job-failing.md)). Its error says which of two things:

- **`the identity key was opened where no grant answers for it`**: in some hour, the cloud's key
  service opened the identity key more times than the control plane granted it to hosts. The
  error gives how many, of how many, and the hour. The control plane also logs `The identity key
  was opened more often than the installation granted it`.
- **Anything else**: the audit could not read the key service's log or the grant records, and
  nothing was checked for that hour.

The job runs every hour and checks the hour that ended 30 minutes before, since the key service's
log arrives late. It exists only where the identity key is kept by a cloud's key service (on AWS,
KMS, with CloudTrail as its log).

## Impact

The identity key signs the tokens a workspace presents to clouds, vaults and clusters outside
Spinbox, which let it in by what it is. Each host that needs it is granted it by the control plane,
and every grant is recorded in the bucket where it cannot be taken back.

- **Opened more than granted**: someone opened the key outside Spinbox's grants, or used a grant
  twice. Whoever holds it can sign tokens as any workspace, and every relying party that trusts
  this installation trusts them. Treat it as a compromise.
- **The audit fails**: a use of the key nobody granted goes unnoticed for as long as it fails.

## Diagnose

1. Read the error and the counts of its last runs (`openings`, `grants`, `unrecorded`) on
   [Admin → Jobs](/admin/jobs?job=identity-audit).
2. **Opened more than granted**: in the cloud's key service log (CloudTrail on AWS), list the
   `Decrypt` calls on the identity key in that hour, and who made each: which role, from which
   machine. Compare with the hosts that joined or restarted then (`spin admin runner list`).
   The grant records are under `identity/granted/` in the installation's bucket, by day.
3. **The audit fails**: the error names what it could not read: the key service's log (the control
   plane's permission to read CloudTrail) or the bucket (`identity/granted/`).

## Fix

- **Opened more than granted**:
  1. Find and stop who opened it: remove the permission that let them (the key policy, the role
     they assumed), and terminate a machine you do not know.
  2. If a runner did it, revoke it (`spin admin runner revoke <runner-id>`); on an autoscaling
     group, `spin admin runner remove <runner-id>`. See
     [host-wrote-foreign-volume](host-wrote-foreign-volume.md) for what else a taken host held.
  3. Tell the relying parties that trust this installation's identity tokens to stop trusting them
     until the key is replaced.

  TODO: the identity key cannot be rotated; it is made once and every relying party trusts its one public key ([spin#1050](https://github.com/spin-stack/spin/issues/1050))
- **The audit fails**: restore the control plane's permission to read the key service's log and
  the bucket, then run it again:

  ```
  spin admin jobs run identity-audit
  ```

  Each run checks only its own hour; hours missed while it failed are not checked again by the job.
  Check them by hand: the `Decrypt` calls in the key service's log against the records under
  `identity/granted/` for the same hours.

  TODO: the audit does not go back over hours it missed while failing ([spin#1051](https://github.com/spin-stack/spin/issues/1051))

## Verify

The next run on [Admin → Jobs](/admin/jobs?job=identity-audit) succeeds, with `openings` no more
than `grants`, and the job alerts resolve.

## Escalate

An opening nobody granted is always reported, privately, at
https://github.com/spin-stack/spin/security/advisories/new with the job's error, the key service's
log of the hour, and `spin version`. Do not open a public issue. An audit that fails with
permissions correct goes to https://github.com/spin-stack/spin/issues.
