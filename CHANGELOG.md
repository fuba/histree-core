## v0.3.3
### Features
- **Library Support**: Refactored core functionality into a reusable Go library
- **Go Get Support**: Project can now be installed via `go install` or imported as a library via `go get`
- **API Documentation**: Added comprehensive documentation for library usage
- **Performance Improvements**: Enhanced database query performance with optimized indexes
- **Timezone Handling**: Improved handling of timestamps across different timezones

### API Changes
- Added public library package `github.com/ec/histree-core/pkg/histree`
- Created stable API for database operations and formatting
- Exported key types and constants for third-party integration

### Bug Fixes
- Fixed issue with timestamp display in different timezones
- Improved error handling in database connections
- Enhanced buffer management for large output sets

### Installation
#### Command-line tool
```bash
go install github.com/ec/histree-core/cmd/histree@latest
```

#### Library
```bash
go get github.com/ec/histree-core
```

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