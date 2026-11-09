---
id: key-rotation
title: Rotate or revoke an installation's keys and credentials
applies_to:
  alerts: []
  jobs: [workspace-rekeys]
  events: []
  dashboards: []
severity: info
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

A key or credential is suspected taken (a leaked token, a runner believed compromised, a person
who left), or is due for rotation. Also read this when the **workspace-rekeys** job fails on
[Admin → Jobs](/admin/jobs).

## Impact

A taken credential acts as whoever it belongs to until it is revoked. A runner's keys open the
disks it held until those disks are re-keyed. While **workspace-rekeys** fails, the revoked
runner's keys still open those disks, and their workspaces wait to start.

## Diagnose

Find what the credential reaches, then pick its section under Fix:

- [Admin → Audit](/admin/audit): what was done with it, from where.
- `spin admin runner list`, `spin admin token list --include-used`: runners and registration
  tokens.
- `spin api-key list` (your own); [Admin → Users](/admin/users) lists each person's API keys.
- `spin admin volume rekeys`: the disks being re-keyed, and those a revoked runner marked.

## Fix

**A runner believed taken** (registered with a token): revoke it. Its credentials are deleted,
its stream is cut, and every disk it held the key of is marked to be re-keyed:

```
spin admin runner revoke <runner-id>
```

It refuses while workspaces' disks are on it; `--leave-workspaces` revokes anyway, and each of
those runs again only once its owner recovers it ([workspace-held](workspace-held.md)). An
autoscaling group's machine is given back instead ([lost-runner](lost-runner.md)); that does not
mark its disks for re-keying. A runner someone had a root shell on is revoked this way on its own
([host-shell](host-shell.md)).

**A workspace's disk key** (re-keyed: copied under a new key, and the old disk and every snapshot
of it destroyed). The **workspace-rekeys** job does this for the disks a revoke marked, once each
workspace is stopped or quarantined. Watch it with `spin admin volume rekeys`. By hand, for a
stopped or quarantined workspace:

```
spin admin volume rekey <workspace>
```

TODO: the CLI looks the workspace up as the caller's own before it calls the admin API, so it
reaches only an administrator's own workspaces; there is no command for another person's
([spin#1046](https://github.com/spin-stack/spin/issues/1046)).

**Runner registration tokens**: `spin admin token revoke <token-id>`. A pool's token is
re-minted on the control plane's machine with `sudo spin-controlplane registration-token ...`.

**API keys**: the owner runs `spin api-key revoke <key-id>`; an administrator revokes anyone's
on [Admin → Users](/admin/users). Keys a key delegated go with it.

**Sessions and people**: disable the person's sign-in on [Admin → Users](/admin/users); they run
`spin logout` on a machine they still hold. A person who lost their passkeys: *Clear passkeys* on
the same page ([admin-lockout](admin-lockout.md) if no administrator can sign in).

**SSH keys**: the owner runs `spin ssh-key delete <key-id>`; a guest is taken out with
`spin workspace guests remove <workspace> <email>`.

**The proxy's token** (an installation whose proxy is not on AWS), on the control plane's
machine:

```
sudo spin-controlplane proxy-token --name <proxy> --revoke
sudo spin-controlplane proxy-token --name <proxy> --expires 720h
```

**GitHub App**: its private key is never kept. To rotate its client secret, delete it on
[Admin → Integrations](/admin/integrations) and create another; every connection is revoked at
GitHub and people connect again (`spin git connect`). A person gives up their own token with
`spin git disconnect github`.

**Team secrets**: replace the value at its source, then set it again from standard input:

```
spin team secret set <team> <name> --host <host>
```

**The installation's encryption key** (seals secrets in the database, and the database's password
and archive keys are derived from it): check it on the control plane's machine:

```
sudo spin-controlplane crypto status
sudo spin-controlplane crypto verify
```

TODO: no procedure or command adds a new version of the installation's key to where an
installation keeps it, before `controlplane crypto rotate` ([spin#1049](https://github.com/spin-stack/spin/issues/1049)). A suspected leak of
this key is a security incident: escalate.

## Verify

[Admin → Audit](/admin/audit) shows each revoke. `spin admin volume rekeys` empties as the job
finishes; the job's last run on Admin → Jobs succeeded. `spin admin token list` and
`spin api-key list` no longer show what was revoked.

## Escalate

A leaked installation key, a runner that reconnects after a revoke, or **workspace-rekeys**
failing again and again are security problems: report them privately through
https://github.com/spin-stack/spin/security/policy with the audit entries, the job's error from
`spin admin jobs list`, and `spin admin versions`.
