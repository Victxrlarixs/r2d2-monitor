# R2-D2 Monitor: User Guide

Welcome to the **R2-D2 Monitor**, a high-performance system telemetry console with a retro astromech aesthetic. This guide will help you master the controls and features of your new droid.

## Getting Started

Simply run `r2d2-monitor.exe`. R2-D2 will automatically launch in a dedicated terminal window and begin scanning your system.

### Controls & Navigation

| Key | Action | Description |
| :--- | :--- | :--- |
| `TAB` | **Switch Focus** | Toggle focus between Processes, Disks, and Network panels. |
| `↑` / `↓` | **Navigate/Cycle** | Move the cursor (Processes) or cycle hardware (Disks/Net). |
| `ENTER` | **Deep Scan** | Open a detailed inspection window for the selected process. |
| `ESC` | **Back/Close** | Close the scan view or exit search mode. |
| `F1` | **Sort by CPU** | Prioritize the list by processor usage. |
| `F2` | **Sort by RAM** | Prioritize the list by memory footprint. |
| `F3` | **Change Theme** | Cycle through available visual styles (Amber, Matrix, etc.). |
| `P` | **Change Layout** | Cycle through layout presets. |
| `/` | **Search** | Filter the process list by name. |
| `F9` | **ZAP (Kill)** | Request process termination. Requires `Y` to confirm. |
| `Q` / `Ctrl+C` | **Quit** | Shut down R2-D2 and exit the console. |

---

## New & Optimized Features

- **Hardware Browser**: Press `TAB` to focus the Disk or Network panels, then use `↑/↓` to cycle through available drives and adapters.
- **Safe Termination**: Killing a process with `F9` now requires a `Y` confirmation to prevent accidental shutdowns.
- **Native Telemetry**: Replaced heavy sub-processes with native WMI/Go calls for Battery, Temperatures, and Network Ping.
- **Multi-Vendor GPU**: Detailed support for NVIDIA (utilization, temp, power) with generic fallback for AMD/Intel.
- **Throttled Polling**: Intelligent polling logic only performs "deep" queries on visible processes, drastically reducing monitor overhead.
- **Braille Graphing**: High-resolution network activity graphs using Unicode Braille patterns.
- **Portability**: All configuration and logs are stored next to the executable.

## Key Features

### 1. Deep Scan (Process Metadata)
By pressing `ENTER`, R2-D2 performs a real-time WMI/PowerShell scan to retrieve:
- Full executable path.
- Company/Developer information.
- Process description and start time.

### 2. Smart Telemetry
The top header provides a high-density overview of your system's health:
- **CPU/RAM/Disk Bars**: Real-time load indicators.
- **Network Traffic**: Live upload/download speeds.
- **Uptime**: How long your droid (system) has been operational.

### 3. Persistent Memory
R2-D2 remembers your preferences! Your chosen **Theme**, **Sort Order**, and **Layout Preset** are saved to a local config file and restored automatically every time you launch the app.

---

## Configuration & Logs

R2-D2 is now fully portable. All files are kept in the same directory as the executable:
- **Config File**: `config.json`
- **Audit Logs**: `logs/r2d2.log`

*If you encounter a process that won't terminate or a scanning error, check the `r2d2.log` file for technical details.*

---
*May the Force be with your CPU.*
