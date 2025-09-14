# DHCP AdGuard Sync for OPNsense

A service that synchronizes DHCP leases from OPNsense to AdGuard Home, ensuring DNS resolution works correctly for all DHCP clients. It supports both IPv4 and IPv6 addresses through DHCP leases and NDP table monitoring.

The application supports both ISC DHCP and DNSMasq lease formats, allowing you to choose which DHCP server's leases to synchronize with AdGuard Home. This ensures that all clients will be properly synchronized to AdGuard Home regardless of which DHCP server you're using.

## Features

- Automatic synchronization of DHCP leases to AdGuard Home
- Support for both ISC DHCP and DNSMasq lease formats (configurable)
- OPNsense plugin with web UI for easy configuration
- IPv6 support through NDP table monitoring
- Real-time lease file monitoring
- Support for hostname customization
- Dry-run mode for testing
- Configurable logging
- Runs as a FreeBSD service

## Prerequisites

- OPNsense
- AdGuard Home installed and running
- AdGuard Home API credentials

## Installation

### One-Line Installation (Recommended)

SSH into your OPNsense firewall and run this single command:

```bash
fetch -o - https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/master/install-oneliner.sh | sh -s -- --username "your-adguard-username" --password "your-adguard-password"
```

This command will:
1. Download the latest version of the plugin
2. Install both the service and the OPNsense GUI components
3. Configure it with your AdGuard Home credentials

After installation:

```bash
# Start the service
service dhcp-adguard-sync start

# Enable at boot
service dhcp-adguard-sync enable
```

Access the plugin in the OPNsense web interface under **Services > DHCP AdGuard Sync**

### Manual Installation (Alternative)

If you prefer to install manually:

```bash
# Download the latest release
fetch -o /tmp/opnsense-lease-sync https://github.com/jeeftor/opnsense-lease-sync/releases/latest/download/dhcp-adguard-sync_freebsd_amd64_v$(curl -s https://api.github.com/repos/jeeftor/opnsense-lease-sync/releases/latest | grep tag_name | cut -d '"' -f 4)

# Make executable
chmod +x /tmp/opnsense-lease-sync

# Install (includes both service and GUI components)
/tmp/opnsense-lease-sync install --username "your-adguard-username" --password "your-adguard-password"
```

For more detailed instructions, see [INSTALL.md](INSTALL.md).

## Configuration

The configuration file is located at `/usr/local/etc/dhcp-adguard-sync/config.yaml`.

### Lease Format Selection

This application supports both ISC DHCP and DNSMasq lease formats:

- Choose the lease format that matches your DHCP server configuration
- The selected lease file is monitored for changes in real-time
- When the file changes, a synchronization is triggered automatically
- The application will parse the lease file according to the selected format

This allows you to use the application with either ISC DHCP (the default in OPNsense) or DNSMasq, depending on your network configuration.

Key configuration options:
```yaml
# AdGuard Home credentials
ADGUARD_USERNAME="admin"
ADGUARD_PASSWORD="password"

# AdGuard Home connection settings
ADGUARD_URL="127.0.0.1:3000"
ADGUARD_SCHEME="http"

# DHCP lease file configuration
DHCP_LEASE_PATH="/var/dhcpd/var/db/dhcpd.leases"    # Path to the DHCP lease file
LEASE_FORMAT="isc"                                  # Lease format: "isc" or "dnsmasq"

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
