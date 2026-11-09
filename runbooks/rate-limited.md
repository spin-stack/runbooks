---
id: rate-limited
title: Requests are being rate limited
applies_to:
  alerts: [rate-limited]
  jobs: []
  events: []
  dashboards: [fleet, spin-api-performance]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**Requests Are Being Rate Limited** fires as soon as the control plane refused more than 50
requests for the per-client rate limit in 10 minutes, for one kind of limit
(`rate_limit_rejected_total`, by `limit_type`):

- `auth`: the sign-in calls, which have their own lower limit.
- `api`: every other call.

The client refused gets `RESOURCE_EXHAUSTED`: "Rate limit exceeded. Please wait before retrying."

## Impact

Either the limit is too low for how the installation is used (a team behind one address, an agent
that calls the API in bursts), and real people are refused; or someone is guessing credentials or
hammering the API, and the limit is doing its job. Both want a look.

## Diagnose

1. *Refused credentials and rate-limited requests* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet): which limit, and whether credentials are
   refused at the same time (a brute force: [auth-failures-looping](auth-failures-looping.md)).
2. Find who: the control plane logs each refusal with the client's address and the method.

   ```
   spin admin logs 'service_name:spin-controlplane AND body.message:"Rate limit exceeded"' --since 1h
   ```

3. Is the address one you know: your office, a CI runner, the installation's own proxy? Is the
   method one that a tool calls in a loop?
4. The limits in force are on [Admin → Security](/admin/security), *Rate limits and sessions*.

## Fix

- **A known client, a limit too low**: raise the limit (requests a minute and burst) on
  [Admin → Security](/admin/security). It applies at once, without a restart.
- **A tool looping**: fix the tool. Its owner can see it in their own use; an API key it holds can
  be revoked by its owner (`spin api-key revoke <key-id>`).
- **Every client shows one address** (the proxy's, or a load balancer's): the control plane is not
  seeing the client's real address, and everyone shares one budget. It trusts `X-Forwarded-For`
  only from a local peer; check what sits in front of it.
- **Guessing or abuse from an unknown address**: leave the limit as it is; block the address at
  the network edge (security group, firewall). Treat a sign-in brute force as a security event
  ([auth-failures-looping](auth-failures-looping.md)).

## Verify

No more than 50 refusals in 10 minutes for that limit, and the alert resolves. The people who were
refused can work.

## Escalate

If the refusals come from addresses that should not be limited and raising the limit does not stop
them, open an issue at https://github.com/spin-stack/spin/issues with the log lines (remove the
addresses first: the tracker is public) and `spin version`. Abuse you cannot stop at the edge:
report it privately at https://github.com/spin-stack/spin/security/advisories/new.
