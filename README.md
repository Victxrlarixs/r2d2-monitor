# R2-D2 Monitor

[![](https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org)
[![Windows](https://img.shields.io/badge/Windows-0078D4?style=flat&logo=windows11&logoColor=white)](https://www.microsoft.com/windows)
[![Star Wars](https://img.shields.io/badge/Star%20Wars-000000?style=flat&logo=starwars&logoColor=FFE81F)](https://www.starwars.com)

![Terminal](https://img.shields.io/badge/CLI-Lover-black?style=flat&logo=gnubash&logoColor=white)

```text
   ___________  
  /  ___ ___  \ 
 |  | (O) |  | |
 |--+-----+--|-|   R2 > *Bleep bloop!*
 | [=]   [=]   |        Systems online and scanning...
 | [ ]---[ ]   |        Signal acquired.
 | [_________] |
 |   |_____| | |
 |___|     |___|
 /____|___|____\
```

R2-D2 Monitor is a high-performance system telemetry console for Windows. It provides real-time monitoring of CPU, RAM, Disk, and Network metrics through a robust Terminal User Interface (TUI) inspired by retro astromech aesthetics.

## Key Features

- Real-time Telemetry: Live tracking of system resources and network traffic.
- Process Management: Integrated process list with the ability to search and forcefully terminate tasks.
- Deep Scan: Retrieve extended metadata for any process, including executable paths and developer information.
- Persistent Configuration: Remembers user preferences for themes and sorting orders.
- Responsive Design: Fluid layout that adapts to any terminal size automatically.
- Asynchronous Logging: Background auditing of system actions and errors.

## Quick Start

Using R2-D2 Monitor is as easy as:
1. Go to the [Releases](https://github.com/Victxrlarixs/r2d2-monitor/releases) page.
2. Download the latest `r2d2-monitor.exe`.
3. Run it and enjoy! No installation or Go environment required.

## Usage

Run the executable to start the monitor:
```bash
./r2d2-monitor.exe
```

### Basic Controls
- Navigation: Use Up/Down arrows to move the cursor.
- Deep Scan: Press Enter to inspect the selected process.
- Change Theme: Press F3 to cycle through visual styles.
- Search: Press / to filter the process list.
- Kill Process: Press F9 to terminate the selected task.
- Quit: Press Q or Esc to exit.

## Documentation

For more detailed information, please refer to the documentation in the /docs directory:

- [User Guide](docs/user_guide.md): Detailed usage instructions and commands.
- [Technical Overview](docs/technical_overview.md): Architecture and data flow.
- [Developer Reference](docs/developer_reference.md): Internal modules and extension guide.
