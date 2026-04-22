# R2-D2 Monitor: Technical Overview

## Project Vision
R2-D2 Monitor is a high-performance system telemetry tool designed for Windows. It combines the low-latency capabilities of Go with the expressive power of the Bubble Tea (The Elm Architecture) framework to deliver a robust Terminal User Interface (TUI).

## Architecture
The project follows a **SOLID-compliant, domain-driven architecture** to ensure long-term maintainability.

### Core Layers:
1.  **Entry Point (`cmd/r2d2-monitor`)**: Handles application bootstrapping, terminal initialization, and the "Mousetrap" bypass for Windows double-click support.
2.  **Engine (`r2d2/`)**:
    *   `StatsManager`: Encapsulated telemetry provider using `gopsutil`. Features a thread-safe cache and smart polling (prioritizing high-CPU tasks).
    *   `Executor`: abstraction layer for OS-level commands (Taskkill, PowerShell/WMI).
    *   `Config`: Persistence layer for user settings using JSON.
    *   `Logger`: Asynchronous file-based logging for auditing system actions.
3.  **UI (`r2d2/ui/`)**: 
    *   **The Elm Architecture (TEA)**: Predictable state management via `Model`, `Update`, and `View`.
    *   **Modular Rendering**: Sub-components (Header, Dialogue, Table) are decoupled into helper functions for better readability and 100% responsive layout calculation.
    *   **High-Res Graphing**: Custom Braille-pattern rendering engine for 1D sparklines (Network/IO visualization).

## Tech Stack
- **Language**: Go 1.26+
- **TUI Framework**: `charmbracelet/bubbletea`
- **Styling**: `charmbracelet/lipgloss`
- **Metrics**: `shirou/gopsutil`
- **CLI Framework**: `spf13/cobra`
- **OS Integration**: PowerShell Core / WMI / Taskkill

## Data Flow
1.  **Init**: `main` loads config -> initializes `StatsManager` -> starts Bubble Tea loop.
2.  **Polling**: Every 2 seconds (or on user action), `fetchStats` triggers a background telemetry scan. Custom PowerShell execution supplements `gopsutil` for advanced Windows-native metrics (Disk IO, Ping, Total Threads/Handles).
3.  **Update**: Metrics are processed, sorted, and stored in the `Model`.
4.  **Render**: The `View` function calculates terminal dimensions and draws the interface using Lipgloss styles.

## CI/CD Pipeline
The project utilizes GitHub Actions for automated quality assurance and distribution:
- **Workflow**: `.github/workflows/build.yml`
- **Runner**: `windows-latest`
- **Automation**: Triggered on every push or pull request to the `main` branch.
- **Artifacts**: Each successful run produces a standalone Windows executable (`r2d2-monitor.exe`) available for download.

---
*Status: Production Ready | Architecture: Modular*
