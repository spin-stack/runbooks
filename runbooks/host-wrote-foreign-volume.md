---
id: host-wrote-foreign-volume
title: A host wrote a volume that is not its
applies_to:
  alerts: [host-writing-volume-not-its]
  jobs: []
  events: []
  dashboards: [storage]
severity: critical
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

**A Host Wrote A Volume That Is Not Its** fires as soon as a restore passed over a signed HEAD
whose writer did not hold the volume (`unadmitted_heads_total`, by `writer_host`, any in 15
minutes). The `writer_host` label names the host that wrote it.

Only a host writing a volume it no longer holds makes one: a host that lost the volume, or was
replaced, and writes on with the credential it kept.

## Impact

Two hosts writing one disk is what the storage design is arranged against. The restore ignored
the foreign HEAD, so the workspace's data is what its rightful writer published. But the named
host still has a credential for the bucket and the key of every disk it held, and is using them.
Treat the host as taken until proven otherwise.

## Diagnose

1. Find the host in `spin admin runner list`: its id is the alert's `writer_host`. Note its state;
   a host that should be gone (revoked, removed, its machine terminated) and still writing is the
   worst case.
2. See what it wrote: *Refused writers: compare-and-set failures* on
   [Admin → Metrics](/admin/metrics?dashboard=storage), and the runners' refusals:

   ```
   spin admin logs 'service_name:spin-runner AND body.message:"will not serve a volume"' --since 24h
   ```

   A `VOLUME_REFUSAL_PUBLISH_FENCED` refusal is the rightful writer finding another host published
   over it.
3. In the cloud, find the machine that answers to that host: on AWS, the instance whose role made
   the requests, in CloudTrail's record of the bucket.

## Fix

1. Take the host out for good.
   - Registered with a token: revoke it. Its credentials are deleted, its stream cut, and every
     team's disk whose key it was given is marked to be re-keyed:

     ```
     spin admin runner revoke <runner-id>
     ```

     If it still holds workspaces' disks the revoke is refused and names them. Move them first,
     or pass `--leave-workspaces`; their owners then recover them
     ([workspace-rehomed](workspace-rehomed.md)).
   - A machine of an autoscaling group: revoke is refused. Give it back to its group, which
     terminates it:

     ```
     spin admin runner remove <runner-id>
     ```

     This does not mark its disks for re-keying; re-key them by hand (step 3).
2. Make sure the machine is terminated in the cloud, and do not reuse its disk.
3. Re-key what it held. The `workspace-rekeys` job copies each marked disk under a new key once
   its workspace is stopped or quarantined. See what is marked and how far along:

   ```
   spin admin volume rekeys
   ```

   Re-key a disk that is not marked (its workspace stopped or quarantined first):

   ```
   spin admin volume rekey <workspace>
   ```

   A re-key deletes the old disk and every snapshot of it. More in [key-rotation](key-rotation.md).

TODO: removing an autoscaling group's runner believed taken does not mark the disks it held for re-keying, and nothing lists them ([spin#1047](https://github.com/spin-stack/spin/issues/1047))

## Verify

`spin admin runner list` shows the host revoked or gone. No new `unadmitted_heads_total` in 15
minutes, and the alert resolves. `spin admin volume rekeys` ends with nothing left marked.

## Escalate

Always. Report it privately at https://github.com/spin-stack/spin/security/advisories/new with
the host id, when it was revoked or removed if it was, the runners' log lines, the cloud's record
of the requests, and `spin version`. Do not open a public issue.
