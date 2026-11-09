---
id: egress-denied
title: A workspace cannot reach something
applies_to:
  alerts: []
  jobs: []
  events: []
  dashboards: [fleet]
severity: info
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

Someone says their workspace cannot reach a host: a connection hangs or is reset, a name does not
resolve (`REFUSED`), `git` or a package manager fails, or an API through the egress proxy
answers with an error. Or the workspace was stopped and held in quarantine after an agent in it
tried something its guard refused.

## Impact

One workspace, or every workspace under one network policy. Nothing is lost; work that needs
the network waits.

How the network decides, so the diagnosis makes sense:

- A workspace runs under its **network policies** (tags). A workspace with any tag reaches only
  what its tags allow, as a union. A create that names none runs under the installation's
  `default_network_tags`, or else `offline`, which allows nothing.
- The runner enforces the policy: a name a policy does not allow is answered `REFUSED`, and an
  address it does not allow is dropped.
- A **brokered** name is reached only through the runner's egress proxy, which may add a team's
  secret to the request.
- Inside a sealed workspace, the **guard** refuses an agent's attempts past its bounds; the first
  denial stops the workspace and holds it in quarantine.

## Diagnose

The owner runs these on their workspace:

1. Which policies it runs under, and which exist:

   ```
   spin workspace get <workspace>
   spin workspace policies
   ```

2. What it has reached, with the runner's own counters, the names it asked for and the
   brokered requests of the last hour:

   ```
   spin workspace egress <workspace>
   ```

   A destination missing here, with its lookup refused, is the policy.
3. To see each lookup and each first connection let through or refused as it happens, turn on
   the watch and read it from the workspace's logs:

   ```
   spin workspace watch <workspace>
   spin workspace logs <workspace> 'service_name:network' --since 1h
   spin workspace watch <workspace> --off
   ```

An administrator also checks:

- [Admin → Network](/admin/network): the policies the installation declares, what each allows,
  and how many runners have installed each. A policy declared but not installed on the runner
  is not enforced there. A team's own policies: `spin team policy list <team>`.
- *Containment and DNS* and *Egress proxy* on [Admin → Metrics](/admin/metrics?dashboard=fleet):
  egress drops, resolver outcomes, and requests the proxy refused.
- A workspace in quarantine: `spin workspace denials <id>` ([quarantine-workspace](quarantine-workspace.md)).

## Fix

- **The destination is not in any of its policies**: add a rule to the right policy on
  [Admin → Network](/admin/network), or to a team's (`spin team policy set <team> <name> --rule ...`),
  then the owner moves the workspace onto it while it runs (the tags given are the whole set):

  ```
  spin workspace retag <workspace> --network-tag <policy> --network-tag <other-policy>
  ```

- **It runs under `offline`** because it was created with no tag and the installation has no
  `default_network_tags`: retag it, and set the default on Admin → Network.
- **A brokered request fails**: the team's secret is missing or not allowed for that host
  (`spin team secret list <team>`; `spin team secret set` names the hosts it may be sent to).
- **It was held after a denial**: follow [quarantine-workspace](quarantine-workspace.md).
- **It reaches far more than expected**: see [workspace-egress-high](workspace-egress-high.md).

## Verify

`spin workspace egress <workspace>` lists the destination with packets both ways, and the
owner's command works. Turn the watch off when done.

## Escalate

If the destination is in an installed policy and is still dropped, or the proxy refuses a request
its rule allows, stop and open an issue at https://github.com/spin-stack/spin/issues with the
policy (rules only, no secrets), the `spin workspace egress` output, the watch's log lines, the
runner id from `spin workspace get` and `spin admin versions`. A workspace reaching what its
policy refuses is a security problem: follow https://github.com/spin-stack/spin/security/policy.
