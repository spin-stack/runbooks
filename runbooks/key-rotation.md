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
last_reviewed: 2026-11-12
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
spin admin volume rekey <workspace-id>
```

It takes the workspace's id, not its name, and works for anybody's workspace. The id is in
`spin admin volume rekeys --json` (`workspace_id`) for a marked disk, or on the workspace's page.

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
and archive keys are derived from it). It has versions: the newest seals, every version still
named opens what was sealed under it. A suspected leak of this key is a security incident:
escalate, and rotate it as below. Check it first, on the control plane's machine:

```
sudo spin-controlplane crypto status
sudo spin-controlplane crypto verify
```

1. **Add the next version where the installation names its key.** On AWS, append it to the
   Terraform module's `encryption_key_versions` (`[1]` becomes `[1, 2]`) and apply. Each version is
   a key of its own agreed with the same KMS key; nothing new is made in KMS. The document changes,
   so the apply replaces the control plane's machine with one that seals under the new version.
   Elsewhere, make a key with `sudo spin-controlplane crypto generate-key` and give the control
   plane `SPIN_CP_ENCRYPTION_KEY="v1:<old key>,v2:<new key>"` where it reads its key, then restart
   it.
2. **Re-seal**, on the control plane's machine that now serves:

   ```
   sudo spin-controlplane crypto rotate --dry-run
   sudo spin-controlplane crypto rotate
   ```

   It refuses while the serving control plane does not seal under the newest version. That
   happens when the replacement has not finished starting, or when you ran it on the old machine.
   Wait for the new machine to serve and run it there. It also refuses a database holding values
   under a version newer than its machine names. Every run ends with *Left under older versions*:
   what is still sealed under each older version, by kind. A run with `FAILED` lines exits
   non-zero: keep the old version and escalate.
3. **Retire the old version** only when both are true:
   - a run reports `v1: none` and no failures;
   - the database's archive keeps no base backup taken before the rotation. Seven are kept, one a
     day, so wait seven days.

   A restore takes only the newest base backup. If the old version must go sooner, first take one
   with `sudo systemctl start spin-postgres-backup.service`. Older base backups then cannot be
   opened ([database-restore](database-restore.md)). Then remove the version (`[2]`) and apply,
   or drop it from `SPIN_CP_ENCRYPTION_KEY` and restart. A version removed while anything is still
   sealed under it is that data lost.

**The installation's identity key** (signs the tokens and certificates workspaces present outside
Spinbox). Rotate it with three steps:

1. Run `sudo spin-controlplane identity rotate` on the control plane's machine.
2. Apply the `identity_keys` list it prints. Old and new keys overlap: both are published, and
   new tokens are signed with the new one.
3. Run `sudo spin-controlplane identity retire <old kid>` and apply the list it prints.

Each step is in [identity-audit-failing](identity-audit-failing.md). It needs a key service: an
installation that keeps its identity key itself, in its database, refuses the rotation.

## Verify

[Admin → Audit](/admin/audit) shows each revoke. `spin admin volume rekeys` empties as the job
finishes; the job's last run on Admin → Jobs succeeded. `spin admin token list` and
`spin api-key list` no longer show what was revoked. After an encryption key rotation,
`sudo spin-controlplane crypto status` marks the new version `<- current` and says every record
is under it, and `sudo spin-controlplane crypto verify` passes after the old version is retired.

## Escalate

A leaked installation key, a runner that reconnects after a revoke, or **workspace-rekeys**
failing again and again are security problems: report them privately through
https://github.com/spin-stack/spin/security/policy with the audit entries, the job's error from
`spin admin jobs list`, and `spin admin versions`.
