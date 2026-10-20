# Changelog

All notable changes to **dmesg-watch** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0-beta.1] - 2026-10-20

The inaugural beta release of **dmesg-watch**, a lightweight Linux desktop background daemon that monitors kernel errors in real time.

### ✨ Highlights
* **Real-time Error Awareness**: Instantly flags kernel faults as they occur without manual terminal inspection.
* **Quiet by Default**: Background tray dashboard tracks error frequency across a rolling 60-second window, alerting only when issues surge.
* **Flood Protection**: Smart notification cooldown prevents banner spam when drivers or hardware emit error bursts.
* **Desktop Native**: Integrates into the system tray across KDE Plasma, GNOME, Sway, and Waybar.
* **At-a-Glance Dashboard**: Context menu presents live error counts, latest truncated message, and quick links to project information.

### 🛠️ Technical Notes
* Implemented in pure Go, maintaining a minimal idle memory footprint (~15MB RSS).
* Direct D-Bus Session Bus integration for StatusNotifierItem (SNI) and desktop notifications without external GUI dependencies.
* Packaged with systemd user unit (`graphical-session.target`), standalone tarballs, and automated `.deb` / `.rpm` distro packages.
