---
id: control-plane-resources
title: The control plane holds too many goroutines or too much memory
applies_to:
  alerts: [high-goroutines, high-memory]
  jobs: []
  events: []
  dashboards: [overview]
severity: warning
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

- **High Goroutine Count** fires when the control plane's process
  (`go_goroutine_count{job="spin-controlplane"}`) runs more than 10,000 goroutines for 5 minutes.
- **High Memory Usage** fires when its Go runtime uses more than 1 GiB
  (`go_memory_used_bytes{job="spin-controlplane"}`) for 10 minutes.

Usually nobody sees anything yet. Later the API slows ([api-latency](api-latency.md)) and the
process can be killed for memory.

## Impact

Both are early signs. A count that keeps climbing is something started and never ended: streams
from runners or browsers that are not closed, or work that waits forever. If the machine runs out
of memory, the control plane is killed and every API call fails until it is back.

## Diagnose

1. Look at the shape on [Admin → Metrics](/admin/metrics?dashboard=overview), *Memory each
   service's Go runtime uses*, over a day or more:
   - A step that stays flat after a burst (many workspaces, many people) is load: the process
     holds what the load needs.
   - A line that climbs steadily without load to match is a leak.
2. Compare with what it serves: hosts connected (`spin admin runner list`), workspaces running and
   people signed in. Each connected runner and each open terminal or editor holds goroutines.
3. Read what the control plane logged around when the climb started:

   ```
   spin admin logs 'service_name:spin-controlplane AND (severity_text:error OR severity_text:warn)' --since 6h
   ```

4. For a leak, take the goroutines' stacks before anything else, on the control plane's machine.
   `SIGQUIT` makes the Go runtime print every goroutine's stack to the journal and exit; systemd
   starts the unit again (`Restart=on-failure`), so this is also a restart:

   ```
   systemctl kill -s QUIT spin-controlplane.service
   journalctl -u spin-controlplane.service --since "5 minutes ago" > goroutines.txt
   ```

## Fix

- **Load**: nothing is wrong; if the machine has room, the thresholds are low for this
  installation. If it does not, give the control plane's machine more memory (a larger instance
  type); read [database-restore](database-restore.md) before replacing the machine.
- **A leak**: take the stacks (step 4), which restarts the process and gives the memory back.
  If you cannot take them, `systemctl restart spin-controlplane.service` alone does the same.
  The API is down for the seconds the process takes to start; running workspaces are not
  touched.

## Verify

The goroutine count and memory fall on [Admin → Metrics](/admin/metrics?dashboard=overview) and the
alerts resolve. Watch for a day: a leak climbs again at the same rate.

## Escalate

A leak is a defect in Spinbox. Open an issue at https://github.com/spin-stack/spin/issues with the
memory panel over the climb, the goroutine stacks from step 4 (read them before you post them: the
tracker is public), and `spin version`.
