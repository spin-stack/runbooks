---
id: volume-not-served
title: A host will not serve a workspace's disk
applies_to:
  alerts: [volume-refused, volume-restores-failing]
  jobs: []
  events: []
  dashboards: [fleet, storage]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

- **A Volume Is Not Served** fires when the database records any volume whose host reported it
  will not serve it (`volumes_refused`, by `refusal`), for 5 minutes. The `refusal` label says why.
- **A Volume Keeps Failing To Restore** fires when more than 3 restores of a volume's published
  chain onto a host failed in 30 minutes, across the fleet
  (`volume_restores_total{outcome="failed"}`), for 10 minutes. Restores back off from five seconds
  to five minutes, so this is usually one volume stuck, not a host in trouble.

The workspace's owner sees a workspace that will not start, or one that stopped writing.

## Impact

The workspace on that volume cannot start, or runs with a disk that is not being published: what
it writes from now on is lost with its host. Each failed restore reads the object store again.

## Diagnose

1. Find the volume and its refusal in the runners' log:

   ```
   spin admin logs 'service_name:spin-runner AND body.message:"will not serve a volume"' --since 24h
   ```

   Each line names the `volume_id` and the `refusal` (as `VOLUME_REFUSAL_ISOLATED` for the
   alert's `isolated`). For restores, search
   `body.message:"restore finished without a chain"` the same way.
2. Find whose it is: `spin admin storage list` lists every volume with what it is.
3. The workspace's owner sees every party's view, the volume's included, with
   `spin workspace status <workspace>`.
4. Read the refusal:

   | `refusal` | What it means |
   |---|---|
   | `image_missing` | The volume's layers are not where the host must find them: a stray delete, a lifecycle expiry in the bucket. |
   | `durability_lost` | QEMU flagged the disk image corrupt. The last commit is the newest point that can be restored. |
   | `no_read_view` | The chain cannot be opened end to end: an unreadable ancestry, or a store that would not answer. |
   | `no_key` | The volume is encrypted and the host holds no key, or the wrong one. |
   | `lease_lost` | The host's lease lapsed and it gave the device up. |
   | `isolated` | The host could reach neither the control plane nor the store, and paused the guest. |
   | `publish_fenced` | Another host published to this volume: this one is not its writer any more. |
   | `history_unanchored` | The bucket's history is not the one the database anchors: something other than the writer changed the bucket. |
   | `attach_failed` | Everything else: QMP not answering, a `qemu-img` run that failed. The log line's error says which. |

5. Look at the host: [Admin → Runners](/admin/runners) and *Volumes: refused, unplaced, at the depth
   ceiling* on [Admin → Metrics](/admin/metrics?dashboard=fleet).

## Fix

- **`isolated`, `lease_lost`**: the host's network. When it reaches the control plane and the
  store again it reports and serves. If it does not come back, follow
  [host-not-reporting](host-not-reporting.md).
- **`publish_fenced`**: two hosts wrote one volume. Follow
  [host-wrote-foreign-volume](host-wrote-foreign-volume.md) before anything else.
- **`history_unanchored`, `image_missing`**: the bucket was changed from outside Spinbox. Find
  who in the cloud's audit log of the bucket; Object Lock kept the version the database anchors.
  Do not delete anything in the bucket by hand. Escalate.
- **`no_key`**: the host's key does not open the volume. Check the host's access to the key
  service (on AWS, its role's permission on the KMS key); the log line's detail says which.
- **`durability_lost`, `no_read_view`, `attach_failed`**: move the workspace off the host. The
  owner runs, and reads the cost before accepting it:

  ```
  spin workspace recover <workspace>
  spin workspace recover <workspace> --accept-data-loss
  ```

  A refused volume counts as stopped, so recover does not wait for the host. The workspace
  restarts from the last commit the store has; see [workspace-rehomed](workspace-rehomed.md).
- **The host is the cause for many volumes**: stop it taking new work with
  `spin admin runner drain <runner-id>`, and look at it on a [host-shell](host-shell.md).

TODO: an administrator cannot list which workspaces have a refused volume, or recover a workspace they do not own ([spin#1045](https://github.com/spin-stack/spin/issues/1045))

## Verify

`volumes_refused` is back to 0 on [Admin → Metrics](/admin/metrics?dashboard=fleet) and the
workspace starts. The restore alert resolves once 30 minutes pass without failed restores.

## Escalate

`history_unanchored` and `image_missing` are always escalated: open an issue at
https://github.com/spin-stack/spin/issues with the refusal, the runner's log line and
`spin version`. If you think someone changed the bucket on purpose, report it privately instead:
https://github.com/spin-stack/spin/security/advisories/new.
