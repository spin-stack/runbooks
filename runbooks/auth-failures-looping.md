---
id: auth-failures-looping
title: A credential is refused again and again
applies_to:
  alerts: [auth-failures-looping]
  jobs: []
  events: []
  dashboards: [fleet, spin-api-performance]
severity: warning
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

**Credentials Refused In A Loop** fires when the control plane refuses presented credentials more
than once every ten seconds, averaged over 5 minutes, for one kind (`auth_failures_total`, by
`kind`), and keeps doing so for 15 minutes. The `kind` label says why:

| `kind` | Means |
|---|---|
| `invalid_credential` | No session, API key, runner token or proxy token has this value: expired, revoked, or guessed. |
| `user_disabled` | The credential is real and its user is disabled. |
| `scope_denied` | An API key used for a call its scopes do not cover. |
| `lookup_failed` | The database could not say either way. |

## Impact

A person gives up after a few tries; something that keeps trying for 15 minutes is a program: a
runner revoked or replaced whose machine still runs, a script with a revoked key, a CLI of a
disabled user, or someone guessing. The first ones only waste requests. Guessing is an attack.

## Diagnose

1. Which kind, and since when: *Refused credentials and rate-limited requests* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet).
2. `lookup_failed`: not a credential problem; the database is failing. Go to
   [database-slow](database-slow.md).
3. Find the caller. Refused calls are `UNAUTHENTICATED` or `PERMISSION_DENIED` on
   [Admin → Metrics](/admin/metrics?dashboard=spin-api-performance), *Methods that fail most*; the
   method says what is calling:
   - `spin.agent.v1.*` methods (a heartbeat, a volume report): a runner. Its log, or the control
     plane's, names the host:

     ```
     spin admin logs 'body.message:"runner token"' --since 1h
     ```

     Compare with `spin admin runner list`: a revoked host still trying is this.
   - Calls a CLI or a script makes: an API key or a session. [Admin → Audit](/admin/audit) shows
     recent key revocations and users disabled.
4. Rate limiting at the same time on `auth` ([rate-limited](rate-limited.md)) with addresses you
   do not know is guessing.

## Fix

- **A revoked runner still running**: terminate its machine; it can never connect again. See
  [lost-runner](lost-runner.md).
- **A script or tool with a revoked or expired key**: its owner gives it a new key
  (`spin api-key create <name>`) or stops it.
- **A disabled user's client**: nothing to fix in the installation; tell the person, or leave it.
- **`scope_denied`**: the key's owner makes a key with the scopes the tool needs.
- **Guessing**: block the source at the network edge. Keep the `auth` rate limit on
  ([Admin → Security](/admin/security)). If any guess may have succeeded, end sessions and revoke
  keys of the accounts it targeted ([key-rotation](key-rotation.md)).

## Verify

The refusals for that kind fall under one every ten seconds on the fleet dashboard and the alert
resolves.

## Escalate

Guessing that succeeded, or that you cannot stop at the edge, is a security incident: report it
privately at https://github.com/spin-stack/spin/security/advisories/new with the kind, the methods,
the times and `spin version`. A refusal loop with no client to explain it goes to
https://github.com/spin-stack/spin/issues.
