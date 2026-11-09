---
id: admin-lockout
title: No administrator can sign in
applies_to:
  alerts: []
  jobs: []
  events: []
  dashboards: []
severity: info
owner: security
last_reviewed: 2026-11-09
---

## Symptoms

Nobody can reach the Admin pages: every administrator is disabled, demoted, deleted, or signs in
through an identity provider that is disabled or gone. Or the one administrator left lost every
passkey and no other can clear them.

The control plane refuses any change that would cause this: the database keeps an administrator
who can sign in (an instance admin, not disabled, with a password or whose last sign-in was
through an enabled installation identity provider), and the dashboard greys out such actions on
[Admin → Users](/admin/users) and [Admin → Security](/admin/security) with the reason. A lockout
still happens from outside the API: a database restored from before, or an edit nobody reviewed.

## Impact

Workspaces keep running and people keep working, but nobody can administer the installation:
no runners, settings, users, jobs or approvals. There is no break-glass in the API by design: the
way back is on the control plane's machine.

## Diagnose

1. If somebody is still an administrator but lost their passkeys, another administrator clears
   them on [Admin → Users](/admin/users) (*Clear passkeys*); they register a new one on their next
   session. Stop here if that works.
2. If the installation's first administrator never chose their own password, that account still
   signs in. On the control plane's machine:

   ```
   sudo spin-controlplane bootstrap-password --user
   sudo spin-controlplane bootstrap-password
   ```

3. Otherwise open the database. `controlplane database exec` is what the database's unit runs,
   not a shell; the shell is the machine's psql over the database's own socket, as root:

   ```
   sudo /usr/libexec/postgresql18/psql -h /run/spin-postgres -U spin spin
   ```

   Who is an administrator, and how they sign in:

   ```
   SELECT id, email, disabled, identity_provider_id, last_login_at
     FROM users WHERE is_instance_admin;
   SELECT id, name, provider_type, enabled
     FROM identity_providers WHERE team_id IS NULL;
   ```

## Fix

Change as little as gets one person back in, in one of these ways, then do the rest from the
dashboard:

- **The installation's identity provider was disabled**: enable it again.

  ```
  UPDATE identity_providers SET enabled = true WHERE id = '<provider-id>';
  ```

- **Make an administrator of somebody who signs in through an enabled provider**:

  ```
  UPDATE users SET is_instance_admin = true WHERE email = '<email>';
  ```

- **An administrator was disabled by mistake**:

  ```
  UPDATE users SET disabled = false, disabled_at = NULL, disabled_reason = NULL
   WHERE email = '<email>';
  ```

- **The last administrator lost every passkey**: what *Clear passkeys* does.

  ```
  DELETE FROM user_passkeys WHERE user_id = '<user-id>';
  ```

These writes are not in the audit log. Write down what you changed, when and why, and give it to
whoever owns security for the installation.

## Verify

The person signs in and opens [Admin → Users](/admin/users). The database's own rule holds again:

```
SELECT installation_has_an_admin_way_in(NULL, NULL);
```

answers `t`. Then make a second administrator, so one person's slip cannot do this again.

## Escalate

If none of these gets anyone in, or the database refuses the update, stop. Do not delete users
or providers. Bring the two `SELECT` results (with emails removed if you post them), the
installation's release (`spin version` from any CLI that still reaches it), and open an issue at
https://github.com/spin-stack/spin/issues. If the lockout came from somebody acting on purpose,
treat it as a security incident and follow https://github.com/spin-stack/spin/security/policy.
