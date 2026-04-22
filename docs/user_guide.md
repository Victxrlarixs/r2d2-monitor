# R2-D2 Monitor: User Guide

Welcome to the **R2-D2 Monitor**, a high-performance system telemetry console with a retro astromech aesthetic. This guide will help you master the controls and features of your new droid.

## Getting Started

Simply run `r2d2-monitor.exe`. R2-D2 will automatically launch in a dedicated terminal window and begin scanning your system.

### Controls & Navigation

| Key | Action | Description |
| :--- | :--- | :--- |
| `↑` / `↓` | **Navigate** | Move the cursor through the active process list. |
| `ENTER` | **Deep Scan** | Open a detailed inspection window for the selected process. |
| `ESC` | **Back/Close** | Close the scan view or exit search mode. |
| `F1` | **Sort by CPU** | Prioritize the list by processor usage. |
| `F2` | **Sort by RAM** | Prioritize the list by memory footprint. |
| `F3` | **Change Theme** | Cycle through available visual styles (Amber, Matrix, etc.). |
| `/` | **Search** | Filter the process list by name. |
| `F9` | **ZAP (Kill)** | Forcefully terminate the selected process and its children. |
| `Q` / `Ctrl+C` | **Quit** | Shut down R2-D2 and exit the console. |

---

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
R2-D2 remembers your preferences! Your chosen **Theme** and **Sort Order** are saved to a local config file and restored automatically every time you launch the app.

---

## Configuration & Logs

- **Config File**: `%USERPROFILE%\.r2d2-monitor\config.json`
- **Audit Logs**: `%USERPROFILE%\.r2d2-monitor\r2d2.log`

*If you encounter a process that won't terminate or a scanning error, check the `r2d2.log` file for technical details.*

---
*May the Force be with your CPU.*
