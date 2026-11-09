---
id: third-party-terms
title: VS Code servers or agent CLIs are not published because their terms are not accepted
applies_to:
  alerts: []
  jobs: [vscode-servers, agent-clis]
  events: []
  dashboards: []
severity: info
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

- The **vscode-servers** or **agent-clis** job on [Admin → Jobs](/admin/jobs) succeeds every day
  but counts only `off`, and publishes nothing.
- A publish is refused with "its third-party terms are not accepted", naming Admin → Extensions.
- `spin code` does not wait for the installation to prepare a VS Code server; VS Code downloads
  its own into each workspace instead.
- Workspaces have no Claude Code, Codex or T3 Code on their `PATH`.
- The job fails (an error, not `off`) on Admin → Jobs.

## Impact

No Spinbox release carries the VS Code server or the agent CLIs: only their makers may
distribute them, so the installation downloads them itself, and only once an administrator
accepts their terms. Each item (the VS Code server; the agent CLIs) is accepted on its own.
Until then:

- VS Code downloads its own server into every workspace it connects to: slower, and it needs
  the workspace's network to reach Microsoft.
- Workspaces carry no agent CLIs from the installation.

An acceptance of an older text counts as none: when a release changes the terms, they must be
accepted again. While a job fails with the terms accepted, workspaces keep the server and agents
last published, and a newer VS Code may not connect.

## Diagnose

```
spin admin jobs list
```

The job's counts say why a day published nothing: `off` (terms not accepted), `released` (the
newest is published already), `busy` (another publish held the job), or `published`.

[Admin → Extensions](/admin/extensions), *Third-party software*, shows each item, who accepted it
and when, and a warning when the accepted text is older than the current one. Its *VS Code
servers* and *Agent CLIs* tabs show what is published.

## Fix

- **Terms not accepted**: an administrator reads them on
  [Admin → Extensions](/admin/extensions) (*Third-party software*) with each part's license, and
  accepts the items the organization agrees to. The pinned releases start publishing at once, and
  the daily jobs keep them current. Accepting is the organization's decision as licensee, not the
  operator's alone.
- **Accepted as an older text**: accept again on the same page.
- **Accepted, and the job fails**: read its error with `spin admin jobs list`. It downloads from
  the makers and uploads to the object store, so check that the control plane reaches both
  (`spin admin health`, [object-store-requests-runaway](object-store-requests-runaway.md)). Then
  run it again:

  ```
  spin admin jobs run vscode-servers
  spin admin jobs run agent-clis
  ```

- **To stop downloading**: *Revoke* on the same page. New downloads stop; what was published stays
  until an administrator removes it.

## Verify

The next run counts `published` or `released`, not `off`. The *VS Code servers* and *Agent CLIs*
tabs list the new releases, and `spin code <workspace>` says it is preparing the installation's
server instead of letting VS Code download one. [Admin → Audit](/admin/audit) has
`third_party_terms_accepted`.

## Escalate

If the terms are accepted and the job still fails with the network and store healthy, stop and
open an issue at https://github.com/spin-stack/spin/issues with the job's error, the run's trace
id from [Admin → Traces](/admin/traces) and `spin admin versions`. Questions about the licenses
themselves go to the organization's own legal counsel, not to Spinbox.
