# DHCP AdGuard Sync for OPNsense

A service that synchronizes DHCP leases from OPNsense to AdGuard Home, ensuring DNS resolution works correctly for all DHCP clients. It supports both IPv4 and IPv6 addresses through DHCP leases and NDP table monitoring.

## Features

- Automatic synchronization of DHCP leases to AdGuard Home
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

1. Download the latest release from the releases page

Or you can try something like the following.

```bash
/bin/sh

# Set variables
OS="Freebsd"
ARCH=$(uname -m)
REPO="jeeftor/opnsense-lease-sync"
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

URL="https://github.com/${REPO}/releases/download/${VERSION}/dhcp-adguard-sync_${OS}_${ARCH}_${VERSION}.tar.gz"

curl -L -o /tmp/dhcp-adguard-sync.tar.gz $URL
tar /tmp/xzf dhcp-adguard-sync.tar.gz

# Clean up
rm /tmp/dhcp-adguard-sync.tar.gz
```


2. Copy to your OPNsense system:
```bash
scp dhcp-adguard-sync root@opnsense:/root/
```

3. SSH into your OPNsense system and install:
```bash
ssh root@opnsense
cd /root
chmod +x dhcp-adguard-sync
./dhcp-adguard-sync install --username "your-adguard-username" --password "your-adguard-password"
```

4. Start the service:
```bash
service dhcp-adguard-sync start
```

5. Enable at boot:
```bash
service dhcp-adguard-sync enable
```

## Configuration

The configuration file is located at `/usr/local/etc/dhcp-adguard-sync/config.yaml`.

Key configuration options:
```yaml
# AdGuard Home credentials
ADGUARD_USERNAME="admin"
ADGUARD_PASSWORD="password"

# AdGuard Home connection settings
ADGUARD_URL="127.0.0.1:3000"
ADGUARD_SCHEME="http"

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

To perform a one-time sync:
```bash
dhcp-adguard-sync sync --username "your-username" --password "your-password"
```

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
    - **Missing clients**: Check if DHCP lease file path is correct
    - **IPv6 not working**: Ensure NDP table is accessible

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details