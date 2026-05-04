# R2-D2 Monitor: Developer Reference

This document provides internal details for developers looking to extend or contribute to the R2-D2 Monitor.

## Project Structure

```text
/
├── cmd/r2d2-monitor/      # Main entry point (bootstrapping)
├── r2d2/                  # Core logic package
│   ├── config.go          # Persistence (JSON)
│   ├── executor.go        # Command execution (PS/WMI)
│   ├── logger.go          # Background auditing
│   ├── stats.go           # Telemetry & Polling logic
│   └── ui/                # UI Package (Bubble Tea)
│       ├── components.go  # Reusable widgets (Bars, etc.)
│       ├── monitor.go     # Main Model & TEA loop
│       ├── view_helpers.go# Header & Dialogue rendering
│       ├── view_details.go# Process Scan view
│       ├── themes.go      # Color palettes
│       ├── reactions.go   # ASCII art & dialogues
│       └── easter_eggs.go # Hidden features
└── docs/                  # Documentation
```

## Key Modules

### `StatsManager`
- **Optimization**: Implements a prioritized polling strategy. Visible/active processes are updated frequently, while idle processes are throttled to minimize overhead. Expensive metrics (like Disk or Uptime) use a `tickCount` system for infrequent refreshes.
- **Concurrency**: Parallelizes process polling using a worker pool (goroutines + semaphores) to ensure the UI remains smooth even with 200+ processes.
- **Robust Parsing**: Implements native WMI struct mapping for reliable data extraction, reducing reliance on external subprocesses.

### `Focus & Hardware Cycling`
- Implements a `FocusMode` system (`Procs`, `Disk`, `Net`) toggled by the `TAB` key.
- Hardware selection (Disk/Net) is persisted in `r2d2.Config`.
- `cycleDisk` and `cycleNet` helpers handle the index rotation and config saving.

### `Responsive Layout`
- The `View()` function in `monitor.go` calculates `topH` (header height) dynamically.
- `listH = H - topH - 1` ensures the table fills every available row until the footer.
- Uses `strings.TrimSuffix` and manual padding for pixel-perfect anchoring.

## Adding New Themes
To add a theme, update `r2d2/ui/themes.go`:
```go
{
    Name: "Hoth",
    CPU:  lipgloss.Color("#76E1FF"), // Primary color
    RAM:  lipgloss.Color("#FFFFFF"), // Secondary
    DSK:  lipgloss.Color("#B3B3B3"), // Accent
    CharAccent: lipgloss.Color("#76E1FF"),
    CharMain:   lipgloss.Color("#FFFFFF"),
},
```

## Extending R2-D2 Reactions
Reactions are defined in `r2d2/ui/reactions.go`. You can add new "Faces" or expand the dialogue pools. Ensure ASCII art is kept at **10 lines** height for layout consistency. The `monitor.go` Update loop now guarantees a continuous feed of `Idle` messages when no specific action is taking place.

## Building
Use the following command to build the project from the root:
```powershell
go build -o r2d2-monitor.exe ./cmd/r2d2-monitor
```

---
*Maintained by the R2-D2 Dev Team*
