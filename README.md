# DHCP AdGuard Sync for OPNsense

[![Go Version](https://img.shields.io/github/go-mod/go-version/jeeftor/dhcpsync)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/jeeftor/dhcpsync)](https://github.com/jeeftor/dhcpsync/releases)
[![License](https://img.shields.io/github/license/jeeftor/dhcpsync)](LICENSE)

> **Automatically sync DHCP clients from OPNsense to AdGuard Home for seamless DNS filtering**

Ever notice that devices on your network don't show up by name in AdGuard Home? This service solves that by automatically synchronizing DHCP lease information from OPNsense to AdGuard Home, ensuring all your devices are properly identified for DNS filtering and monitoring.

## ✨ What This Solves

- **📱 Device Recognition**: See device names instead of IP addresses in AdGuard Home
- **🔄 Automatic Sync**: No manual client configuration in AdGuard Home
- **📊 Better Analytics**: Proper device identification for detailed statistics
- **🛡️ Enhanced Filtering**: Apply DNS rules based on device names

## 🚀 Quick Start

**One-line installation:**
```bash
curl -sSL https://raw.githubusercontent.com/jeeftor/dhcpsync/master/install.sh | sh
```

**Then configure via OPNsense Web UI:**
1. Navigate to **Services > DHCP AdGuard Sync**
2. Enter your AdGuard Home credentials
3. Select your DHCP server type (DNSMasq is default)
4. Click **Save** - service auto-restarts!

**That's it!** Your DHCP clients will now appear in AdGuard Home.

## Architecture

This project consists of two main components:

1. **Main Application** (`dhcpsync`): A Go-based service that handles the actual synchronization
2. **OPNsense Plugin**: A web UI plugin that integrates with OPNsense for easy configuration and management

The plugin provides a user-friendly interface within OPNsense while the main application handles the core functionality.

## 🎯 Features

| Feature | Description |
|---------|-------------|
| 🔄 **Auto-Sync** | Real-time DHCP lease monitoring and synchronization |
| 🖥️ **Web UI** | Native OPNsense plugin for easy configuration |
| 🏷️ **Device Names** | See friendly hostnames instead of IP addresses |
| 📡 **IPv6 Support** | Handles both IPv4 and IPv6 via NDP table monitoring |
| ⚙️ **Multi-Format** | Supports both ISC DHCP and DNSMasq lease formats |
| 🧪 **Test Mode** | Dry-run capability for safe testing |
| 📝 **Logging** | Configurable log levels with rotation |
| 🚀 **Service** | Runs as native FreeBSD service |

## 📋 Prerequisites

- ✅ OPNsense firewall
- ✅ AdGuard Home installed and running
- ✅ AdGuard Home admin credentials
- ✅ Root access to OPNsense (for installation)

## 💾 Installation

> **💡 Pro Tip**: The automatic installation is recommended for most users

### 🚀 Option 1: Automatic Installation (Recommended)

Install both components with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/jeeftor/dhcpsync/master/install.sh | sh
```

This script will:
- Download the latest release for your platform
- Install the main application binary and service
- Install the OPNsense plugin (if running on OPNsense)
- Set up initial configuration

### Option 2: Manual Installation

#### Step 1: Install Main Application

1. **Download the binary** from the [releases page](https://github.com/jeeftor/dhcpsync/releases):
   ```bash
   # Example for FreeBSD amd64
   curl -L -o dhcpsync.tar.gz \
     "https://github.com/jeeftor/dhcpsync/releases/download/v0.0.26/dhcpsync_freebsd_amd64_v0.0.26.tar.gz"
   tar -xzf dhcpsync.tar.gz
   ```

2. **Install on your OPNsense system**:
   ```bash
   # Copy binary to OPNsense
   scp dhcpsync root@opnsense:/root/

   # SSH and install
   ssh root@opnsense
   cd /root
   chmod +x dhcpsync
   ./dhcpsync install --username "your-adguard-username" --password "your-adguard-password"
   ```

3. **Start and enable the service**:
   ```bash
   service dhcpsync start
   service dhcpsync enable
   ```

#### Step 2: Install OPNsense Plugin (Optional)

The plugin provides a web UI for easy configuration and management.

1. **Download the plugin package**:
   ```bash
   curl -L -o os-dhcpsync-plugin.tar.gz \
     "https://github.com/jeeftor/dhcpsync/releases/download/v0.0.26/os-dhcpsync-plugin.tar.gz"
   ```

2. **Extract and install the plugin**:
   ```bash
   tar -xzf os-dhcpsync-plugin.tar.gz
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
- **Via command line**: Edit `/usr/local/etc/dhcpsync/config.env`
- **Via CLI commands**: Use `dhcpsync --help` for options

## Configuration

The configuration file is located at `/usr/local/etc/dhcpsync/config.env`.

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
LOG_FILE="/var/log/dhcpsync.log"
```

## Usage

### Service Management

Start the service:
```bash
service dhcpsync start
```

Stop the service:
```bash
service dhcpsync stop
```

Check status:
```bash
service dhcpsync status
```

### View Logs

Via OPNsense UI:
1. Navigate to System > Log Files
2. Select "General" tab
3. Look for entries from "dhcpsync"

Via command line:
```bash
# View service log file
tail -f /var/log/dhcpsync.log

# View system log entries
grep dhcpsync /var/log/messages
```

### Manual Sync

To perform a one-time sync with ISC DHCP lease format (default):
```bash
dhcpsync sync --username "your-username" --password "your-password" --lease-path "/var/dhcpd/var/db/dhcpd.leases"
```

To perform a one-time sync with DNSMasq lease format:
```bash
dhcpsync sync --username "your-username" --password "your-password" --lease-path "/var/db/dnsmasq.leases" --lease-format "dnsmasq"
```

The application will read the lease file using the specified format and synchronize all clients to AdGuard Home.

### Command-Line Help

For a complete list of available options:
```bash
dhcpsync --help
dhcpsync sync --help
```

This will show all available options, including `--lease-path` for the lease file path and `--lease-format` to select between "isc" and "dnsmasq" formats.

## Uninstallation

1. Stop and disable the service:
```bash
service dhcpsync stop
service dhcpsync disable
```

2. Run the uninstall command:
```bash
# Keep configuration files
dhcpsync uninstall

# Remove configuration files as well
dhcpsync uninstall --remove-config

# Force uninstallation if experiencing issues
dhcpsync uninstall --force
```

## 🔧 Troubleshooting

### Quick Diagnostics

```bash
# Check if service is running
service dhcpsync status

# Test configuration
dhcpsync sync --dry-run

# View recent logs
tail -50 /var/log/dhcpsync.log
```

### Common Issues & Solutions

<details>
<summary><strong>🚫 Service won't start</strong></summary>

**Symptoms**: Service fails to start or immediately stops
**Solutions**:
1. Check configuration file exists: `ls -la /usr/local/etc/dhcpsync/config.env`
2. Verify binary permissions: `ls -la /usr/local/bin/dhcpsync`
3. Check logs: `grep dhcpsync /var/log/messages`
</details>

<details>
<summary><strong>🔌 No clients appearing in AdGuard Home</strong></summary>

**Symptoms**: Service runs but no devices show up in AdGuard Home
**Solutions**:
1. Verify AdGuard credentials work: Test login at AdGuard Home web interface
2. Check DHCP lease file: `ls -la /var/db/dnsmasq.leases` (or `/var/dhcpd/var/db/dhcpd.leases` for ISC)
3. Confirm lease format matches your DHCP server
4. Run test sync: `dhcpsync sync --dry-run`
</details>

<details>
<summary><strong>🔄 Sync happens but clients disappear</strong></summary>

**Symptoms**: Devices appear briefly then vanish from AdGuard Home
**Solutions**:
1. Enable "Preserve Deleted Hosts" in plugin settings
2. Check for conflicting AdGuard Home settings
3. Verify DHCP lease renewal times aren't too short
</details>

<details>
<summary><strong>📡 IPv6 devices not syncing</strong></summary>

**Symptoms**: Only IPv4 devices appear in AdGuard Home
**Solutions**:
1. Ensure IPv6 is enabled in OPNsense DHCP settings
2. Check NDP table: `ndp -a`
3. Verify IPv6 DHCP leases exist
</details>

### Enable Debug Mode

For detailed troubleshooting, enable debug logging:

**Via Web UI**: Navigate to Services > DHCP AdGuard Sync > Enable Debug Mode
**Via CLI**: Edit config file and set `LOG_LEVEL="debug"`, then restart service

### Getting Help

If issues persist:
1. Enable debug logging
2. Reproduce the issue
3. Collect logs: `tail -100 /var/log/dhcpsync.log`
4. [Open an issue](https://github.com/jeeftor/dhcpsync/issues) with logs and configuration details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details
