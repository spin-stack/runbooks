---
id: kept-volumes-growing
title: A host keeps disk for more and more volumes
applies_to:
  alerts: [kept-volumes-growing]
  jobs: []
  events: []
  dashboards: [storage]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

**A Host Keeps Disk For More And More Volumes** fires when a host (`instance`) keeps local disk
for more than 20 volumes more than it did 24 hours ago (`kept_volumes` now minus 24 hours ago),
for an hour.

A kept volume is one the host no longer serves but still holds on its disk: a released volume
often comes back, so the host keeps it until the control plane says it was deleted, or that
another host holds it now.

## Impact

Each kept volume is a workspace's disk on that host. Growing without end, they fill the host's
disk and leave less room to place new workspaces. Nothing is lost.

## Diagnose

1. Open *Kept, not served: released volumes a host still holds, and the ones it freed* on
   [Admin → Metrics](/admin/metrics?dashboard=storage). It shows the kept count per host and the
   ones freed, by why (`deleted`, `moved`).
   - **Nothing freed for `deleted`** while workspaces are being deleted: deletions are not
     reaching the host.
   - **Freed as usual, kept still climbing**: released volumes nobody takes back - stopped
     workspaces whose disk stays on this host.
2. Check that deletions finish: the deleted workspaces wait out the installation's trash days,
   then the `workspace-trash` job erases them. A host frees a deleted volume only once it is gone
   both from the database and from the object store, and only when the store answers. On
   [Admin → Jobs](/admin/jobs?job=workspace-trash), is the job succeeding?
3. Check that the host asks: the runner logs when it cannot free a volume.

   ```
   spin admin logs 'service_name:spin-runner AND body.message:"could not free the disk"' --since 24h
   spin admin logs 'service_name:spin-runner AND body.message:"volumes this host keeps"' --since 24h
   ```

4. Check how full the host's disk is: *Disk promised against disk taken, per host* on the same
   dashboard.

## Fix

- **`workspace-trash` failing**: follow [job-failing](job-failing.md); the kept disk is freed at
  the host's next report after the trash is erased.
- **The runner cannot free the disk** (log lines from step 3): the error says why; usually the
  host's own disk or file system. Look at it on a [host-shell](host-shell.md).
- **Stopped workspaces piling up on this host**: that is what keeping is for, and they free
  themselves when deleted or moved. If the host's disk is running out, stop placing on it with
  `spin admin runner drain <runner-id>`, and ask the owners of workspaces they no longer use to
  delete them.
- **Workspaces in the trash hold it**: the trash keeps their disks for the installation's
  trash days ([Admin → Settings](/admin/settings)); fewer days frees them sooner.

## Verify

The kept count on the storage dashboard stops climbing, and freed counts move again. The alert
resolves when the host keeps no more than 20 volumes over the day before.

## Escalate

If deleted workspaces are erased (the job succeeds) and the host still frees none for `deleted`,
it is a defect: open an issue at https://github.com/spin-stack/spin/issues with the host id, the
panel over two days, the runner's log lines and `spin version`.
