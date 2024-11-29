# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Before any major/minor/patch bump all unit tests will be run to verify they pass.

## [Unreleased]

-   [x]

# [0.1.0] - 2024-11-29

### Changed

-   `-P` flag changed to `-r` for setting rcon password. This is to disambiguate it from the port (-p) flag.

# [0.0.3] - 2024-11-24

### Changed

-   {Rcon}.login is no longer exported since it's called internally by the constructor.
-   When checking the timeouts map the cmd is split from its arguments. This allows setting a timeout value for all `map mp_` for example.

### Added

-   Timeout values for commands in the timeouts map are now logged at Debug level.

# [0.0.1] - 2024-11-04

### Added

-   Initial release, package implements Rcon using the Q3 protocol.
-   A basic CLI implementation accepting configuration flags.
