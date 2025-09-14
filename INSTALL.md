# Installing DHCP AdGuard Sync Plugin

This document explains how to install the DHCP AdGuard Sync plugin on your OPNsense firewall using our custom repository.

## Method 1: Using the Custom Repository (Recommended)

### Step 1: Add the Repository

SSH into your OPNsense firewall and run:

```bash
fetch -o /usr/local/etc/pkg/repos/dhcpadguardsync.conf https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/repo-packages/repo/dhcpadguardsync.conf
```

### Step 2: Update Package Cache

```bash
pkg update
```

### Step 3: Install the Plugin

You can install the plugin in two ways:

#### Option A: Using the Command Line

```bash
pkg install os-dhcpadguardsync
```

#### Option B: Using the OPNsense Web Interface

1. Navigate to **System > Firmware > Plugins**
2. Click **Check for updates**
3. Find **os-dhcpadguardsync** in the list
4. Click **Install**

### Step 4: Configure the Plugin

1. Navigate to **Services > DHCP AdGuard Sync**
2. Configure the plugin settings:
   - Enter your AdGuard Home credentials
   - Select the lease format (ISC DHCP or DNSMasq)
   - Set the lease file path
   - Configure other options as needed
3. Click **Save**
4. Click **Enable** to start the service

## Method 2: Manual Installation

If you prefer to install the plugin manually:

### Step 1: Download the Plugin Package

Download the latest package from our releases page:
https://github.com/jeeftor/opnsense-lease-sync/releases

### Step 2: Copy to OPNsense

Copy the package to your OPNsense system:

```bash
scp os-dhcpadguardsync-*.txz root@your-opnsense-ip:/tmp/
```

### Step 3: Install the Package

SSH into your OPNsense system and install the package:

```bash
pkg add /tmp/os-dhcpadguardsync-*.txz
```

### Step 4: Configure as in Method 1, Step 4

## Updating the Plugin

When a new version is released, you can update using:

### Option A: Using the Command Line

```bash
pkg update
pkg upgrade os-dhcpadguardsync
```

### Option B: Using the OPNsense Web Interface

1. Navigate to **System > Firmware > Plugins**
2. Click **Check for updates**
3. If an update is available, click **Update** next to the plugin

## Troubleshooting

If you encounter issues:

1. Check the OPNsense system logs at **System > Log Files > General**
2. Verify your AdGuard Home is accessible from OPNsense
3. Ensure the lease file path is correct and readable
4. Verify the lease format setting matches your DHCP server type

For additional help, please open an issue on our GitHub repository:
https://github.com/jeeftor/opnsense-lease-sync/issues
