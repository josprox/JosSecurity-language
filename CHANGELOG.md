# Changelog

## [3.6.3] - 2026-07-28

### Fixed

- WebSocket `onMessage` callbacks now retain `$ws`, route parameters, and local variables after the handler returns.
- Captured callbacks created in the same scope share state across messages and `onClose`, and their invocations are serialized per connection.
- WebSockets now support `onClose`, local channel subscriptions/publication, automatic subscription cleanup, write-error reporting, message limits, ping/pong keepalive, and safe callback panic isolation.

## [3.6.2] - 2026-07-28

### Fixed

- Verification resends reuse an unexpired token instead of invalidating links already delivered by email.
- Login failures distinguish an unverified account from an incorrect password without relying on token generation.
- Password resets are atomic, reject weak or expired credentials, consume each token once, and verify the account after proving access to its email.
- Authentication emails and expiry timestamps are normalized consistently across SQLite, MySQL, and PostgreSQL.

### Security

- Verified accounts no longer produce a truthy fake verification token.
- Password reset table initialization is only cached after successful creation.
- API and web 2FA challenges now expire after five minutes, cannot authorize protected routes, and can only be exchanged once for an access token.

## [3.6.0] - 2026-07-14

### Added

- JP v2 bytecode packages with public symbol metadata and native `joss-rpc-v1` payloads.
- Plugin SDKs for C/C++, Python, PHP, Java, Kotlin, Dart/Flutter and Rust.
- Manual multiplatform distribution workflow and platform-aware remote installers.

### Changed

- `func` is the only function keyword; the former `function` spelling is rejected with a migration diagnostic.
- Dependencies declared in `joss.yaml` autoload; source-level `import`, `use` and namespace statements have been removed.
- Documentation now describes the actual parser, runtime, CLI, server, views, database, plugins and known limits.

### Fixed

- Registered native method surfaces now match their runtime handlers.
- `Response::error()` returns structured JSON, redirects honor an optional status, and request accessors honor defaults.
- Generated projects and editor snippets use syntax accepted by the current parser.

## [3.0.7] - 2025-12-22
### Fixed
- **Auth**: Fixed `user_role` not being restored from JWT claims, preventing admins from seeing admin-only UI.
- **Request**: Fixed `Request::all()` and `Request::except()` to exclude internal `_cookies` map, preventing database errors.
- **Database**: Added safety check in `GranDB` (SQLite/MySQL) insert methods to ignore unsupported `map` types.
- **View**: Fixed `@foreach` rendering for `map` types (specifically dates) by using Regex replacement for `{{ $var }}` tags. Added support for both dot (`$item.key`) and bracket (`$item['key']`) notation.
- **Handler**: Updated Session restoration logic to correctly populate `user_role` from JWT.
