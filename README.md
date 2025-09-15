# DHCP AdGuard Sync for OPNsense

A service that synchronizes DHCP leases from OPNsense to AdGuard Home, ensuring DNS resolution works correctly for all DHCP clients. It supports both IPv4 and IPv6 addresses through DHCP leases and NDP table monitoring.

The application supports both ISC DHCP and DNSMasq lease formats, allowing you to choose which DHCP server's leases to synchronize with AdGuard Home. This ensures that all clients will be properly synchronized to AdGuard Home regardless of which DHCP server you're using.

## Architecture

This project consists of two main components:

1. **Main Application** (`dhcp-adguard-sync`): A Go-based service that handles the actual synchronization
2. **OPNsense Plugin**: A web UI plugin that integrates with OPNsense for easy configuration and management

The plugin provides a user-friendly interface within OPNsense while the main application handles the core functionality.

## Features

- Automatic synchronization of DHCP leases to AdGuard Home
- Support for both ISC DHCP and DNSMasq lease formats (configurable)
- **OPNsense plugin with web UI** for easy configuration and management
- IPv6 support through NDP table monitoring
- Real-time lease file monitoring
- Support for hostname customization
- Dry-run mode for testing
- Configurable logging
- Runs as a FreeBSD service
- **Service management through OPNsense web interface**

## Prerequisites

- OPNsense
- AdGuard Home installed and running
- AdGuard Home API credentials

## Installation

This project has two components that work together:
1. **Main Application**: The Go binary that performs the actual DHCP-AdGuard synchronization
2. **OPNsense Plugin**: Web UI for configuration and management (optional but recommended)

### Option 1: Automatic Installation (Recommended)

Install both components with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/master/install.sh | sh
```

This script will:
- Download the latest release for your platform
- Install the main application binary and service
- Install the OPNsense plugin (if running on OPNsense)
- Set up initial configuration

### Option 2: Manual Installation

#### Step 1: Install Main Application

1. **Download the binary** from the [releases page](https://github.com/jeeftor/opnsense-lease-sync/releases):
   ```bash
   # Example for FreeBSD amd64
   curl -L -o dhcp-adguard-sync.tar.gz \
     "https://github.com/jeeftor/opnsense-lease-sync/releases/download/v0.0.26/dhcp-adguard-sync_freebsd_amd64_v0.0.26.tar.gz"
   tar -xzf dhcp-adguard-sync.tar.gz
   ```

2. **Install on your OPNsense system**:
   ```bash
   # Copy binary to OPNsense
   scp dhcp-adguard-sync root@opnsense:/root/

   # SSH and install
   ssh root@opnsense
   cd /root
   chmod +x dhcp-adguard-sync
   ./dhcp-adguard-sync install --username "your-adguard-username" --password "your-adguard-password"
   ```

3. **Start and enable the service**:
   ```bash
   service dhcp-adguard-sync start
   service dhcp-adguard-sync enable
   ```

#### Step 2: Install OPNsense Plugin (Optional)

The plugin provides a web UI for easy configuration and management.

1. **Download the plugin package**:
   ```bash
   curl -L -o os-dhcpadguardsync-plugin.tar.gz \
     "https://github.com/jeeftor/opnsense-lease-sync/releases/download/v0.0.26/os-dhcpadguardsync-plugin.tar.gz"
   ```

2. **Extract and install the plugin**:
   ```bash
   tar -xzf os-dhcpadguardsync-plugin.tar.gz
   cd opnsense-plugin/src
   cp -r opnsense/* /usr/local/opnsense/
   ```

3. **Clear OPNsense caches**:
   ```bash
   rm -f /tmp/opnsense_menu_cache.xml
   rm -f /tmp/opnsense_acl_cache.json
   service configd restart
   service php-fpm restart
   ```

### Configuration Options

After installation, you can configure the service:

- **Via OPNsense Web UI** (if plugin installed): Navigate to Services > DHCP AdGuard Sync
- **Via command line**: Edit `/usr/local/etc/dhcp-adguard-sync/config.yaml`
- **Via CLI commands**: Use `dhcp-adguard-sync --help` for options

## Configuration

The configuration file is located at `/usr/local/etc/dhcp-adguard-sync/config.yaml`.

### Lease Format Selection

This application supports both ISC DHCP and DNSMasq lease formats:

- **Important**: DNSMasq is now the default DHCP server in OPNsense
- Choose the lease format that matches your DHCP server configuration
- The selected lease file is monitored for changes in real-time
- When the file changes, a synchronization is triggered automatically
- The application will parse the lease file according to the selected format

#### DNSMasq Configuration (Default in OPNsense)

If you're using DNSMasq (the default in current OPNsense versions), make sure to set:
```yaml
LEASE_FORMAT="dnsmasq"
DHCP_LEASE_PATH="/var/db/dnsmasq.leases"
```

#### ISC DHCP Configuration (Legacy)

If you're using the older ISC DHCP server:
```yaml
LEASE_FORMAT="isc"
DHCP_LEASE_PATH="/var/dhcpd/var/db/dhcpd.leases"
```

Key configuration options:
```yaml
# AdGuard Home credentials
ADGUARD_USERNAME="admin"
ADGUARD_PASSWORD="password"

# AdGuard Home connection settings
ADGUARD_URL="127.0.0.1:3000"
ADGUARD_SCHEME="http"

# DHCP lease file configuration
DHCP_LEASE_PATH="/var/db/dnsmasq.leases"            # Path to the DNSMasq lease file (default in OPNsense)
LEASE_FORMAT="dnsmasq"                             # Lease format: "dnsmasq" or "isc"

# Legacy ISC DHCP configuration (commented out for reference)
#DHCP_LEASE_PATH="/var/dhcpd/var/db/dhcpd.leases"   # Path to the ISC DHCP lease file
#LEASE_FORMAT="isc"                                # For ISC DHCP server

# Optional settings
#PRESERVE_DELETED_HOSTS="false"
#DEBUG="false"
#DRY_RUN="false"
ADGUARD_TIMEOUT="10"

# Logging configuration
LOG_LEVEL="info"
LOG_FILE="/var/log/dhcp-adguard-sync.log"
```

## Usage

### Service Management

Start the service:
```bash
service dhcp-adguard-sync start
```

Stop the service:
```bash
service dhcp-adguard-sync stop
```

Check status:
```bash
service dhcp-adguard-sync status
```

### View Logs

Via OPNsense UI:
1. Navigate to System > Log Files
2. Select "General" tab
3. Look for entries from "dhcp-adguard-sync"

Via command line:
```bash
# View service log file
tail -f /var/log/dhcp-adguard-sync.log

# View system log entries
grep dhcp-adguard-sync /var/log/messages
```

### Manual Sync

To perform a one-time sync with ISC DHCP lease format (default):
```bash
dhcp-adguard-sync sync --username "your-username" --password "your-password" --lease-path "/var/dhcpd/var/db/dhcpd.leases"
```

To perform a one-time sync with DNSMasq lease format:
```bash
dhcp-adguard-sync sync --username "your-username" --password "your-password" --lease-path "/var/db/dnsmasq.leases" --lease-format "dnsmasq"
```

The application will read the lease file using the specified format and synchronize all clients to AdGuard Home.

### Command-Line Help

For a complete list of available options:
```bash
dhcp-adguard-sync --help
dhcp-adguard-sync sync --help
```

This will show all available options, including `--lease-path` for the lease file path and `--lease-format` to select between "isc" and "dnsmasq" formats.

## Uninstallation

1. Stop and disable the service:
```bash
service dhcp-adguard-sync stop
service dhcp-adguard-sync disable
```

2. Run the uninstall command:
```bash
# Keep configuration files
dhcp-adguard-sync uninstall

# Remove configuration files as well
dhcp-adguard-sync uninstall --remove-config

# Force uninstallation if experiencing issues
dhcp-adguard-sync uninstall --force
```

## Troubleshooting

1. Check service status:
```bash
service dhcp-adguard-sync status
```

2. Enable debug logging:
    - Edit `/usr/local/etc/dhcp-adguard-sync/config.yaml`
    - Set `LOG_LEVEL="debug"`
    - Set `DEBUG="true"`
    - Restart the service

3. Common issues:
    - **Service won't start**: Check logs for permissions issues
    - **No synchronization**: Verify AdGuard credentials and connection settings
    - **Missing clients**: Check if the lease file path is correct and the lease format matches your DHCP server:
      - ISC DHCP lease file: `/var/dhcpd/var/db/dhcpd.leases` (default)
      - DNSMasq lease file: `/var/db/dnsmasq.leases` (typical location)
    - **Incorrect client information**: Ensure the lease format setting matches your DHCP server type
    - **IPv6 not working**: Ensure NDP table is accessible

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details
