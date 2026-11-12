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
last_reviewed: 2026-11-12
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
  3. Rotate the identity key, after step 1: the new key is sealed under the same key service key,
     so whoever opened the old one could open the new one the same way.
     1. On the control plane's machine:

        ```
        sudo spin-controlplane identity rotate
        ```

        It makes the next key and prints its kid and the `identity_keys` list to apply. Nothing
        publishes, accepts or signs with that key until the list is applied. Run it again and it
        prints the same key, without making another.
     2. Set the Terraform module's `identity_keys` to that list and apply. The control plane and
        the proxy are replaced onto it. From then on, `https://id.<your domain>` publishes both
        keys, and every host signs new tokens and certificates with the new key once it
        reconnects. Tokens the old key signed still verify.
     3. On the new control plane's machine:

        ```
        sudo spin-controlplane identity retire <old kid>
        ```

        The old kid is the one the first step did not print, or the first `kid` in the issuer's
        `/.well-known/jwks.json`. It prints the list with the old key retired an hour from now,
        once everything it signed has ended. Apply that list. After that time the old key is gone
        from the key set and the SPIFFE bundle, and the control plane refuses its tokens.
        `allowed_signers` keeps it, valid before that time, so commits signed earlier still
        verify. In a compromise, an earlier time ends the old key sooner; tokens it signed stop
        verifying at once.
  4. Tell the relying parties that trust this installation's identity tokens what happened, and
     that the old kid is retired. A relying party that copied the key set or the SPIFFE bundle
     instead of reading `https://id.<your domain>` must take the new one.
- **The audit fails**: restore the control plane's permission to read the key service's log and
  the bucket, then run it again:

  ```
  spin admin jobs run identity-audit
  ```

  Each run audits every settled hour since the last one it audited, oldest first, at most 24 a
  run. An hour it cannot read stops the run and is the next run's first. On
  [Admin → Jobs](/admin/jobs?job=identity-audit), `hours` is the hours audited this run, `behind`
  the settled hours still owed, and `unaudited` the hours the key service's log no longer keeps.

  If the run fails with `audit: the key service no longer keeps its log of hours that were never
  audited`, those hours (the error names how many and between when) are the only ones to check
  by hand: the `Decrypt` calls in whatever record of them you still have against the records
  under `identity/granted/` for the same hours.

## Verify

The next run on [Admin → Jobs](/admin/jobs?job=identity-audit) succeeds, with `openings` no more
than `grants`, and the job alerts resolve. After a rotation,
`https://id.<your domain>/.well-known/jwks.json` names the new kid. Once the old key's retired
time has passed, it no longer names the old kid. The control plane's journal says
`The installation's identity keys` with the new `signing_kid`.

## Escalate

An opening nobody granted is always reported, privately, at
https://github.com/spin-stack/spin/security/advisories/new with the job's error, the key service's
log of the hour, and `spin version`. Do not open a public issue. An audit that fails with
permissions correct goes to https://github.com/spin-stack/spin/issues.
