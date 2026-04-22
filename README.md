# R2-D2 Monitor

R2-D2 Monitor is a high-performance system telemetry console for Windows. It provides real-time monitoring of CPU, RAM, Disk, and Network metrics through a robust Terminal User Interface (TUI) inspired by retro astromech aesthetics.

## Key Features

- Real-time Telemetry: Live tracking of system resources and network traffic.
- Process Management: Integrated process list with the ability to search and forcefully terminate tasks.
- Deep Scan: Retrieve extended metadata for any process, including executable paths and developer information.
- Persistent Configuration: Remembers user preferences for themes and sorting orders.
- Responsive Design: Fluid layout that adapts to any terminal size automatically.
- Asynchronous Logging: Background auditing of system actions and errors.

## Tech Stack

- Language: Go
- TUI Framework: Bubble Tea
- Styling: Lipgloss
- Metrics: gopsutil
- CLI: Cobra

## Automated Builds

This project uses GitHub Actions for Continuous Integration. On every push or merge to the main branch, a Windows executable is automatically generated. You can find the latest builds in the **Actions** tab of the repository under the "Build and Package R2-D2 Monitor" workflow.

## Installation

### Prerequisites
- Windows OS
- Go 1.21 or higher (for building from source)

### Building from Source
1. Clone the repository:
   ```bash
   git clone https://github.com/victx/r2d2-monitor.git
   ```
2. Navigate to the project directory:
   ```bash
   cd r2d2-monitor
   ```
3. Build the executable:
   ```bash
   go build -o r2d2-monitor.exe ./cmd/r2d2-monitor
   ```

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

## License

This project is licensed under the MIT License.
