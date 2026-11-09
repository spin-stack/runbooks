package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

const good = `---
id: %s
title: A host stopped reporting
applies_to:
  alerts: [host-heartbeat-stale]
severity: critical
owner: operator
last_reviewed: 2026-11-09
---

## Symptoms

The alert fired. See [the other](other.md) and [jobs](/admin/jobs).

## Impact

Workspaces on it.

## Diagnose

` + "`spin admin hosts`" + `

## Fix

Drain it.

## Verify

It reports.

## Escalate

The release's maintainers: https://github.com/spin-stack/spin/issues
`

func runbook(id string, edit func(string) string) string {
	s := strings.Replace(good, "%s", id, 1)
	if edit != nil {
		s = edit(s)
	}
	return s
}

func replace(old, new string) func(string) string {
	return func(s string) string { return strings.Replace(s, old, new, 1) }
}

func TestCheckRefusesWhatTheAppCannotRead(t *testing.T) {
	cases := []struct {
		name string
		file string
		edit func(string) string
		want string // empty: no problem
	}{
		{name: "a good runbook", file: "host-down", want: ""},
		{name: "id is not the file's name", file: "host-gone", want: `is not the file's name`},
		{name: "a section missing", file: "host-down", edit: replace("## Verify\n", ""), want: "sections are"},
		{name: "sections out of order", file: "host-down", edit: replace("## Impact", "## Fix2"), want: "sections are"},
		{name: "raw HTML", file: "host-down", edit: replace("Drain it.", "Drain <b>it</b>."), want: "raw HTML"},
		{name: "an HTML block", file: "host-down", edit: replace("Drain it.", "<script>alert(1)</script>"), want: "raw HTML"},
		{name: "a runbook that does not exist", file: "host-down", edit: replace("other.md", "nowhere.md"), want: `links runbook "nowhere"`},
		{name: "a javascript link", file: "host-down", edit: replace("(/admin/jobs)", "(javascript:alert(1))"), want: "neither another runbook"},
		{name: "a severity of its own", file: "host-down", edit: replace("severity: critical", "severity: high"), want: "severity"},
		{name: "a date that is not one", file: "host-down", edit: replace("2026-11-09", "last week"), want: "last_reviewed"},
		{name: "a field nobody reads", file: "host-down", edit: replace("owner: operator", "owner: operator\nteam: x"), want: "front matter"},
		{name: "an event spelled as the proto does", file: "host-down", edit: replace("alerts: [host-heartbeat-stale]", "events: [WORKSPACE_EVENT_TYPE_HELD]"), want: "event"},
		{name: "applies to nothing and pages", file: "host-down", edit: replace("  alerts: [host-heartbeat-stale]\n", ""), want: "applies to nothing"},
		{name: "an account id", file: "host-down", edit: replace("Drain it.", "Account 123456789012."), want: "AWS account id"},
		{name: "an installation's hostname", file: "host-down", edit: replace("Drain it.", "Open app.spinbox.dev."), want: "hostname"},
		{name: "a level-one heading", file: "host-down", edit: replace("## Symptoms", "# Title\n\n## Symptoms"), want: "level-one"},
		{name: "no front matter", file: "host-down", edit: replace("---\nid:", "id:"), want: "front matter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"runbooks/" + c.file + ".md": {Data: []byte(runbook("host-down", c.edit))},
				"runbooks/other.md":         {Data: []byte(runbook("other", replace("(other.md)", "(host-down.md)")))},
			}
			problems, err := check(fsys, "runbooks")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.Join(problems, "\n")
			if c.want == "" {
				if got != "" {
					t.Fatalf("want no problem, got:\n%s", got)
				}
				return
			}
			if !strings.Contains(got, c.want) {
				t.Fatalf("want a problem saying %q, got:\n%s", c.want, got)
			}
		})
	}
}

func TestCheckRefusesAnIDUsedTwice(t *testing.T) {
	fsys := fstest.MapFS{
		"runbooks/other.md": {Data: []byte(runbook("other", nil))},
		"runbooks/copy.md":  {Data: []byte(runbook("other", nil))},
	}
	problems, err := check(fsys, "runbooks")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(problems, "\n"), "used by") {
		t.Fatalf("want an id used twice refused, got %q", problems)
	}
}
