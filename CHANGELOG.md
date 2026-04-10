# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.5] - 2026-04-10

### Fixed
- **Nil pointer panics**: Fixed potential crashes in `filter`, `paginate`, `wizard`, `stats`, and `doctor` commands when database queries failed
- **Error handling**: Improved error handling for flag parsing in root command
- **Download reliability**: Added retry logic (3 attempts with exponential backoff) for binary downloads
- **Checksum verification**: Added SHA256 verification for downloaded binaries

### Added
- **Windows ARM64 support**: Added Windows ARM64 binary build to release pipeline
- **Better error messages**: Improved error messages for unsupported platforms and download failures

### Changed
- **NPM installer**: Enhanced install.js with better error handling and file size validation

## [1.0.4] - Previous Release

See git history for previous changes.
