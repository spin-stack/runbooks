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
last_reviewed: 2026-11-22
---

## Symptoms

**Requests Are Being Rate Limited** fires as soon as the control plane refused more than 50
requests for the rate limit in 10 minutes, for one kind of limit
(`rate_limit_rejected_total`, by `limit_type`):

- `auth`: the login calls, which have their own lower limit and are counted per address, whoever
  makes them.
- `api`: every other call. A person who is logged in has a budget of their own, shared by all their
  sessions and API keys. A request nobody is logged in for (none, a credential refused, a runner's or
  the proxy's) is counted against its address.

The client refused gets `RESOURCE_EXHAUSTED`: "Rate limit exceeded. Please wait before retrying."

## Impact

Either the limit is too low for how the installation is used (one person's agents and keys that
call the API in bursts together, or one address that many callers who are not logged in share), and real
people are refused; or someone is guessing credentials or
hammering the API, and the limit is doing its job. Both want a look.

## Diagnose

1. *Refused credentials and rate-limited requests* on
   [Admin → Metrics](/admin/metrics?dashboard=fleet): which limit, and whether credentials are
   refused at the same time (a brute force: [auth-failures-looping](auth-failures-looping.md)).
2. Find who: the control plane logs each refusal with the client's address, the method and, when
   the caller was logged in, their `user_id`.

   ```
   spin admin logs 'service_name:spin-controlplane AND body.message:"Rate limit exceeded"' --since 1h
   ```

3. A `user_id` in the lines means one person's own budget ran out: who is that, and is the
   method one that one of their tools calls in a loop? No `user_id` means an address's budget:
   is it one you know (a CI runner, the installation's own proxy)?
4. The limits in force are on [Admin → Security](/admin/security), *Rate limits and sessions*.

## Fix

- **A known person or address, a limit too low**: raise the limit (requests a minute and burst) on
  [Admin → Security](/admin/security). It applies at once, without a restart.
- **A tool looping**: fix the tool. Its owner can see it in their own use; an API key it holds can
  be revoked by its owner (`spin api-key revoke <key-id>`). Every session and key of one person
  spends the same budget, so a tool that loops refuses that person's other work too.
- **Refusals of callers who are not logged in all show one address** (the proxy's, or a load balancer's): the control
  plane is not seeing the caller's real address, and they share one budget. It takes
  `X-Forwarded-For` only from a trusted proxy (loopback always, and whatever
  [Admin → Security](/admin/security) lists as *Trusted proxies*); add what sits in front of it.
  People who are logged in are not affected: their budget is their own, whatever address they come from.
- **Guessing or abuse from an unknown address**: leave the limit as it is; block the address at
  the network edge (security group, firewall). Treat a login brute force as a security event
  ([auth-failures-looping](auth-failures-looping.md)).

## Verify

No more than 50 refusals in 10 minutes for that limit, and the alert resolves. The people who were
refused can work.

## Escalate

If the refusals come from addresses that should not be limited and raising the limit does not stop
them, open an issue at https://github.com/spin-stack/spin/issues with the log lines (remove the
addresses first: the tracker is public) and `spin version`. Abuse you cannot stop at the edge:
report it privately at https://github.com/spin-stack/spin/security/advisories/new.
