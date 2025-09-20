# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based service that synchronizes DHCP leases from OPNsense to AdGuard Home. The project has two main components:

1. **Main Application** (`opnsense-lease-sync`): A Go service that handles DHCP lease synchronization
2. **Embedded OPNsense Plugin**: A PHP-based web UI plugin that integrates with OPNsense, now embedded within the Go binary for unified installation

## Development Commands

### Building
```bash
# Build the main binary
make build

# Release build (requires goreleaser)
make release

# Local development build
go build -o build/opnsense-lease-sync .
```

### Code Quality
```bash
# Format Go code (used by pre-commit hooks)
goimports -w .
gofumpt -w -l .
golines -w .

# Run pre-commit hooks manually
pre-commit run --all-files
```

### Testing
```bash
# Run Go tests (if any exist)
go test ./...

# Test installation script
./test-install.sh
```

### Dependencies
```bash
# Update Go dependencies
go mod tidy
go mod download
```

### VM Development
```bash
# Deploy to local OPNsense VM for testing (via Makefile)
make dev-sync

# Force deploy (ignores errors, starts services)
make dev-sync-force

# Alternative: Use dev.sh script if available (not in repo)
./dev.sh deploy
./dev.sh binary
./dev.sh plugin
./dev.sh watch
./dev.sh ssh
./dev.sh test
```

## Architecture

### Core Components

- **main.go**: Entry point that delegates to the cobra CLI framework
- **cmd/**: Contains all CLI commands and application configuration
  - `root.go`: Base cobra command with flag definitions and validation
  - `sync_cmd.go`: One-time sync command
  - `serve_cmd.go`: Service/daemon mode
  - `install.go`/`uninstall.go`: Installation management
  - `install_plugin.go`/`uninstall_plugin.go`: Plugin installation management
  - `plugin.go`: Plugin-related commands
  - `embed.go`: Embedded file handling
  - `paths.go`: Path management utilities
  - `version.go`: Version information
- **pkg/**: Core business logic packages
  - `sync.go`: Main synchronization service orchestrator
  - `dhcp.go`/`dnsmasq.go`: DHCP lease file parsers (ISC DHCP vs DNSMasq formats)
  - `adguard.go`: AdGuard Home API client wrapper
  - `ndpWatcher.go`: IPv6 NDP table monitoring for IPv6 support
  - `types.go`: Common data structures and constants
  - `appConfig.go`: Configuration management
  - `logger.go`: Logging setup and configuration
  - `plugin/`: Embedded OPNsense plugin functionality
    - `plugin.go`: Plugin installation and management logic
    - `opnsense-plugin/`: PHP plugin files embedded in the binary

### OPNsense Plugin Structure

- **opnsense-plugin/src/opnsense/**: Standard OPNsense plugin layout
  - `mvc/app/models/OPNsense/DHCPAdGuardSync/`: Data models and XML configurations
    - `DHCPAdGuardSync.xml`: Main model configuration
    - `DHCPAdGuardSync.php`: Model class
    - `Menu/Menu.xml`: Navigation menu integration
    - `ACL/ACL.xml`: Access control definitions
  - `mvc/app/controllers/OPNsense/DHCPAdGuardSync/`: PHP controllers for web UI and API
    - `IndexController.php`: Main web interface controller
    - `Api/ServiceController.php`: Service management API
    - `Api/SettingsController.php`: Configuration API
    - `forms/`: XML form definitions for UI
  - `mvc/app/views/OPNsense/DHCPAdGuardSync/`: Volt templates for web interface
    - `index.volt`: Main plugin page
    - `settings.volt`: Settings configuration page
  - `service/conf/actions.d/actions_dhcpadguardsync.conf`: Service action definitions
  - `service/templates/OPNsense/DHCPAdGuardSync/dhcpadguardsync.conf`: Service template

### Key Design Patterns

- **Strategy Pattern**: Different lease readers (ISC DHCP vs DNSMasq) implement the `LeaseReader` interface
- **File Watching**: Uses fsnotify to monitor DHCP lease files for real-time updates
- **Configuration**: Supports both command-line flags and environment variables with precedence
- **Dual Mode**: Can run as either a one-time CLI tool or persistent service

### DHCP Lease Format Support

The application supports two DHCP server formats:
- **ISC DHCP**: Legacy format (`/var/dhcpd/var/db/dhcpd.leases`)
- **DNSMasq**: Default in current OPNsense (`/var/db/dnsmasq.leases`)

Use the `--lease-format` flag or `LEASE_FORMAT` environment variable to specify which parser to use.

## Configuration

### Environment Variables
Key environment variables that can be used instead of command-line flags:
- `ADGUARD_USERNAME`, `ADGUARD_PASSWORD`: AdGuard Home credentials
- `ADGUARD_URL`, `ADGUARD_SCHEME`: AdGuard Home connection settings
- `DHCP_LEASE_PATH`, `LEASE_FORMAT`: DHCP lease file configuration
- `LOG_LEVEL`, `LOG_FILE`: Logging configuration
- `DRY_RUN`, `DEBUG`: Development/testing flags

### Configuration File
Service mode uses `/usr/local/etc/dhcpsync/config.env` for persistent configuration.

## FreeBSD/OPNsense Specifics

- Designed to run as a FreeBSD service (`/usr/local/etc/rc.d/dhcpsync`)
- Plugin integrates with OPNsense's MVC framework
- Supports both IPv4 (DHCP leases) and IPv6 (NDP table monitoring)
- Uses standard OPNsense paths and conventions

## Development Workflow

### Local VM Testing
The project includes `dev.sh` (not checked into git) for deploying to a local OPNsense VM:

**Setup Requirements:**
- Local OPNsense VM at `192.168.1.158` (configurable)
- SSH access with username/password or key authentication
- `sshpass` for password authentication: `brew install hudochenkov/sshpass/sshpass`
- `rsync` for efficient file synchronization
- Optional: `fswatch` for automatic redeployment: `brew install fswatch`

**Deployment Process:**
1. **Binary**: Builds Go application and deploys to `/usr/local/bin/opnsense-lease-sync`
2. **Plugin**: Syncs PHP files to `/usr/local/opnsense/mvc/app/`
3. **Services**: Restarts `configd`, clears PHP cache, restarts `nginx`
4. **Testing**: Verifies deployment and file placement

**Development Modes:**
- `./dev.sh deploy`: Full deployment (default)
- `./dev.sh binary`: Deploy only Go binary (faster for Go changes)
- `./dev.sh plugin`: Deploy only PHP plugin (faster for UI changes)
- `./dev.sh watch`: Auto-deploy on file changes
- `./dev.sh ssh`: SSH into VM for debugging

This workflow enables rapid iteration on both the Go service and PHP plugin components without manual file copying or service management.

## Release Process

- Uses GoReleaser for cross-platform builds (FreeBSD, Linux on multiple architectures)
- Pre-commit hooks ensure code quality and conventional commits
- Builds are triggered by git tags and create GitHub releases automatically
