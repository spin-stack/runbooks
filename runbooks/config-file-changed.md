---
id: config-file-changed
title: The config file proposes a change nobody has answered
applies_to:
  alerts: []
  jobs: []
  events: []
  dashboards: []
severity: info
owner: operator
last_reviewed: 2026-11-22
---

## Symptoms

The overview lists **Config file** as needing you: the installation's config file says something
the installation does not, and the file has said it for more than a day. The config file seeds an
installation once; after that the database is what runs, and a later change in the file is a
proposal that waits for an administrator.

## Impact

Nothing changes until it is answered, so the installation keeps running what it had, which is not
what the file's author meant. A release's new default, an address added to a trusted proxy range or
a size resized stays unapplied: workspaces keep the old floor and ceiling, a proxy keeps refusing the new
range.

## Diagnose

1. Read what each change says and what the installation has: the overview lists each with the file's
   words, the installation's and since when.
2. Decide per item whether the file or the installation is right. A file that is older than the
   installation's own edits is usually the one that is wrong.

## Fix

Answer each change, from the overview or the page that shows the item:

- **Apply** writes what the file says into the installation.
- **Keep current** keeps what the installation has. The answer is remembered until the file says
  something else again, so the change stops waiting.

When the file is the one that is wrong, correct it as well, or the next release will propose the
same change again.

## Verify

The overview no longer lists **Config file**, and the item shows what you chose.

## Escalate

If a change cannot be answered, or an item reappears after it was answered with the file unchanged,
open an issue at https://github.com/spin-stack/spin/issues with the item's name and what the file
and the installation say.
