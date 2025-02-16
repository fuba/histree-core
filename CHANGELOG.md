## v0.2.0

### Breaking Changes
- Remove session_label concept, replace with hostname and process_id
- Update database schema (migration script provided)

### Features
- Add version information (`-version` flag)
- Add explicit hostname and process_id tracking
- Improve error handling in database operations
- Add database migration support

### Migration
If upgrading from v0.1.x, run the migration script:
```bash
./scripts/migrate_v0.1_v0.2.sh
```

### Installation
1. Download the `histree.tar.gz` file
2. Extract it: `tar xzf histree.tar.gz`
3. Run the installation: `make install`