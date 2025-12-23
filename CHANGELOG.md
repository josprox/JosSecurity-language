# Changelog

## [3.0.7] - 2025-12-22
### Fixed
- **Auth**: Fixed `user_role` not being restored from JWT claims, preventing admins from seeing admin-only UI.
- **Request**: Fixed `Request::all()` and `Request::except()` to exclude internal `_cookies` map, preventing database errors.
- **Database**: Added safety check in `GranDB` (SQLite/MySQL) insert methods to ignore unsupported `map` types.
- **View**: Fixed `@foreach` rendering for `map` types (specifically dates) by using Regex replacement for `{{ $var }}` tags.
- **Handler**: Updated Session restoration logic to correctly populate `user_role` from JWT.
