# Making DHCP AdGuard Sync Available to Other OPNsense Users

This document outlines how to distribute your DHCP AdGuard Sync plugin to other OPNsense users.

## Method 1: Direct Installation (Simplest)

The simplest way to distribute your plugin is through direct installation. Users can run:

```bash
fetch -o /tmp/install.sh https://raw.githubusercontent.com/jeeftor/opnsense-lease-sync/master/install.sh
chmod +x /tmp/install.sh
/tmp/install.sh
```

This method is straightforward and doesn't require maintaining a package repository.

## Method 2: GitHub Releases (Recommended)

GitHub Releases provide a reliable way to distribute your plugin:

1. Create detailed GitHub releases with clear version numbers
2. Include the FreeBSD binary in the release assets
3. Provide clear installation instructions in the release notes

Users can then download and install the binary directly:

```bash
fetch -o /tmp/opnsense-lease-sync https://github.com/jeeftor/opnsense-lease-sync/releases/latest/download/dhcp-adguard-sync_freebsd_amd64_v0.0.15
chmod +x /tmp/opnsense-lease-sync
/tmp/opnsense-lease-sync install --username "your-adguard-username" --password "your-adguard-password"
```

## Method 3: Custom Repository (Advanced)

For a more professional approach like mimugmail/opn-repo, you'll need:

1. A FreeBSD system to build proper packages
2. A web server to host your repository
3. Knowledge of FreeBSD package management

### Setting Up the Repository

1. Create a repository configuration file:

```
# jeeftor.conf
jeeftor: {
  url: "https://your-domain.com/repo/FreeBSD:13:amd64",
  mirror_type: "http",
  signature_type: "none",
  enabled: yes,
  priority: 50
}
```

2. Build proper FreeBSD packages using `pkg create`
3. Create repository metadata using `pkg repo`
4. Host everything on a web server or GitHub Pages

### User Installation

Users would install your repository with:

```bash
fetch -o /usr/local/etc/pkg/repos/jeeftor.conf https://your-domain.com/jeeftor.conf
pkg update
pkg install os-dhcpadguardsync
```

## Recommended Approach

For your current setup, I recommend:

1. Focus on Method 1 (Direct Installation) and Method 2 (GitHub Releases)
2. Create a simple landing page with installation instructions
3. Keep your GitHub releases well-documented and up-to-date

This approach balances ease of distribution with user-friendliness without requiring complex infrastructure.
