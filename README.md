## Overview
This fork of CockroachDB adds a few small quality of life improvements.

See the [diff for version 19.2.6](https://github.com/cockroachdb/cockroach/compare/v19.2.6...karlseguin:x19.2.6).

All changes to code are also marked with the special string `+vtfb`.

## Notice
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, TITLE AND NON-INFRINGEMENT

### Insecure Password Authentication
Support has been added for password authentication in --insecure mode. The most common reason to want this is if your network is already secured and you want to avoid the performance and management overhead.

Specifically:
1. `create/alter user X with password 'Y'` is allowed
2. `cockroach sql --insecure` allows a password in the url (or will prompt for one)
3. The server will enforce the password set for a user.
4. `root` can login without a password, but only from 127.0.0.1

### Telemetry Disabled By Default
Telemetry has been turned off by default.

Set the environemnt variable `COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING` to `false` to enable it.



