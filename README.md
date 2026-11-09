# spin-stack / runbooks

What an operator of a [Spinbox](https://github.com/spin-stack/spin) installation does when an
alert fires, a job keeps failing, a workspace's owner is told something went wrong, or a
procedure is asked for: one markdown file per runbook under `runbooks/`.

An installation's administrators read them inside the app, at **Admin → Runbooks**, without
leaving it: spin pins this repository at a commit and serves the runbooks of the release it
runs. Every Grafana alert rule links its runbook, and so do Admin → Jobs and the notices a
workspace's owner gets. **Open in repo**, on each runbook in the app, is the file here at that
commit, to propose a change.

```
go tool task check      # what CI runs: front matter, sections, links, nothing private
```

[CONTRIBUTING.md](CONTRIBUTING.md) is how a runbook is written and what the check refuses.

## License

Apache License 2.0, as spin.
