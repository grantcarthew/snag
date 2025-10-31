# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-11-XX

### Added

- Initial stable release of snag
- Intelligent web page fetching using Chrome DevTools Protocol
- Five output formats: Markdown (default), HTML, Text, PDF, PNG
- Smart browser management (auto-detect existing browser or launch headless)
- Tab management capabilities:
  - List all open browser tabs
  - Fetch content from specific tabs by index or URL pattern
  - Pattern matching with exact URL, substring, and regex support
  - Fetch all tabs at once with auto-generated filenames
- Authentication support via persistent browser sessions
- Multiple URL processing in single command
- Output options:
  - Write to stdout (default for text formats)
  - Save to file with `--output`
  - Auto-generate filenames to directory with `--output-dir`
  - Auto-generated timestamped filenames for binary formats (PDF, PNG)
- Page loading controls:
  - Configurable timeout with `--timeout`
  - Wait for specific CSS selectors with `--wait-for`
  - Custom user agent support with `--user-agent`
  - Custom Chrome profile support with `--user-data-dir`
- Browser controls:
  - Force headless mode with `--force-headless`
  - Open visible browser with `--open-browser`
  - Control tab closing behavior with `--close-tab`
  - Custom remote debugging port with `--port`
  - Kill orphaned browser processes with `--kill-browser`
- Logging levels: quiet, normal, verbose, debug
- Diagnostic information with `--doctor` flag
- Cross-platform support: macOS (arm64/amd64), Linux (amd64/arm64)
- Single binary distribution with no runtime dependencies
- Comprehensive documentation and examples

[unreleased]: https://github.com/grantcarthew/snag/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/grantcarthew/snag/releases/tag/v1.0.0
