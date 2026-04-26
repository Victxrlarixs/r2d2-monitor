package r2d2

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds summarized telemetry data for a single system process.
type ProcessInfo struct {
	ID      string
	Name    string
	CPU     float64
	MEM     string
	Threads int64
}

// GPUInfo holds NVIDIA GPU telemetry collected via nvidia-smi.
type GPUInfo struct {
	Name        string
	Utilization float64 // %
	VRAMUsed    float64 // MB
	VRAMTotal   float64 // MB
	Temp        float64 // °C
	Power       float64 // W
	Available   bool
}

// BatteryInfo holds telemetry for the system's power source.
type BatteryInfo struct {
	Percent float64
	Status  string // Charging, Discharging, Full, etc.
}

// SysStats aggregates system-wide metrics and a list of active processes.
type SysStats struct {
	CPU          float64
	CPUCores     []float64
	CPUTemps     []float64 // °C per physical package (index 0 = package avg)
	RAM          float64
	RAMUsed      float64
	RAMTotal     float64
	RAMAvailable float64
	RAMCached    float64
	Swap         float64
	SwapUsed     float64
	SwapTotal    float64
	Disk         float64
	DiskUsed     float64
	DiskTotal    float64
	Uptime       string
	Processes    []ProcessInfo
	TotalProcs   int
	NetSent      float64
	NetRecv      float64
	TotalNetSent float64
	TotalNetRecv float64
	DiskRead     float64 // KB/s
	DiskWrite    float64 // KB/s
	NetPing      int     // ms
	OSName       string
	CPUModel     string
	Battery      BatteryInfo
	GPU          GPUInfo
	LocalIP      string
}

// StatsManager handles the collection and caching of system telemetry.
type StatsManager struct {
	procCache   map[int32]*process.Process
	nameCache   map[int32]string
	cpuCache    map[int32]float64
	memCache    map[int32]string
	cacheMutex  sync.Mutex
	lastRefresh time.Time
	tickCount   uint64

	lastDisk      float64
	lastDiskUsed  float64
	lastDiskTotal float64
	lastUptime    string

	lastNetRecv uint64
	lastNetSent uint64
	lastNetTime time.Time // timestamp for network KB/s delta

	lastDiskRead  uint64
	lastDiskWrite uint64
	lastDiskTime  time.Time // timestamp for disk KB/s delta — separate from net
	lastPing      int

	lastGPU      GPUInfo   // cached GPU stats (refreshed every 5 ticks)
	lastCPUTemps []float64 // cached CPU package temps (refreshed every 5 ticks)
}

// NewStatsManager initializes a new telemetry provider with empty caches.
func NewStatsManager() *StatsManager {
	return &StatsManager{
		procCache: make(map[int32]*process.Process),
		nameCache: make(map[int32]string),
		cpuCache:  make(map[int32]float64),
		memCache:  make(map[int32]string),
	}
}

// GetStats collects current system metrics and process telemetry.
func (sm *StatsManager) GetStats() SysStats {
	sm.tickCount++
	stats := SysStats{
		Disk:      sm.lastDisk,
		DiskUsed:  sm.lastDiskUsed,
		DiskTotal: sm.lastDiskTotal,
		Uptime:    sm.lastUptime,
	}
	if stats.Uptime == "" {
		stats.Uptime = "0d 0h 0m"
	}

	if v, err := mem.VirtualMemory(); err == nil && v != nil {
		stats.RAM = v.UsedPercent
		stats.RAMUsed = float64(v.Used) / 1024 / 1024 / 1024
		stats.RAMTotal = float64(v.Total) / 1024 / 1024 / 1024
		stats.RAMAvailable = float64(v.Available) / 1024 / 1024 / 1024
		stats.RAMCached = float64(v.Cached) / 1024 / 1024 / 1024
	}
	if s, err := mem.SwapMemory(); err == nil && s != nil {
		stats.Swap = s.UsedPercent
		stats.SwapUsed = float64(s.Used) / 1024 / 1024 / 1024
		stats.SwapTotal = float64(s.Total) / 1024 / 1024 / 1024
	}
	if c, err := cpu.Percent(0, false); err == nil && len(c) > 0 {
		stats.CPU = c[0]
	}
	if cores, err := cpu.Percent(0, true); err == nil {
		stats.CPUCores = cores
	}
	if info, err := cpu.Info(); err == nil && len(info) > 0 {
		stats.CPUModel = info[0].ModelName
	}
	if h, err := host.Info(); err == nil {
		stats.OSName = fmt.Sprintf("%s %s", h.Platform, h.PlatformVersion)
	}

	// Local IP retrieval (Simple)
	if n, err := net.Interfaces(); err == nil {
		for _, i := range n {
			for _, a := range i.Addrs {
				if strings.Contains(a.Addr, ".") && !strings.HasPrefix(a.Addr, "127.") {
					stats.LocalIP = strings.Split(a.Addr, "/")[0]
					break
				}
			}
			if stats.LocalIP != "" {
				break
			}
		}
	}

	// Battery via PowerShell (Robust for Windows)
	batOut, _ := ExecuteCommand("Get-CimInstance -ClassName Win32_Battery | Select-Object EstimatedChargeRemaining, BatteryStatus | ConvertTo-Json")
	if strings.Contains(batOut, "EstimatedChargeRemaining") {
		var b struct {
			EstimatedChargeRemaining int
			BatteryStatus            int
		}
		if json.Unmarshal([]byte(batOut), &b) == nil {
			stats.Battery.Percent = float64(b.EstimatedChargeRemaining)
			statusMap := map[int]string{1: "Discharging", 2: "AC Power", 3: "Fully Charged", 6: "Charging"}
			stats.Battery.Status = statusMap[b.BatteryStatus]
			if stats.Battery.Status == "" {
				stats.Battery.Status = "Unknown"
			}
		}
	}

	// GPU + CPU temps: refresh every 5 ticks to avoid blocking the UI goroutine.
	if sm.tickCount%5 == 0 || sm.tickCount == 1 {
		sm.lastGPU = collectGPU()
		sm.lastCPUTemps = collectCPUTemps()
	}
	stats.GPU = sm.lastGPU
	stats.CPUTemps = sm.lastCPUTemps

	now := time.Now()
	if sm.tickCount%10 == 0 || sm.lastRefresh.IsZero() {
		targetDisk := "C:"
		d, err := disk.Usage(targetDisk)
		if err != nil {
			targetDisk = "/"
			d, _ = disk.Usage(targetDisk)
		}
		if d != nil {
			sm.lastDisk = d.UsedPercent
			sm.lastDiskUsed = float64(d.Used) / 1024 / 1024 / 1024
			sm.lastDiskTotal = float64(d.Total) / 1024 / 1024 / 1024
		}

		if u, err := host.Uptime(); err == nil {
			sm.lastUptime = fmt.Sprintf("%dd %dh %dm", u/86400, (u%86400)/3600, (u%3600)/60)
		}
		sm.lastRefresh = now

		// Update the current stats object with these fresh values immediately
		stats.Disk = sm.lastDisk
		stats.DiskUsed = sm.lastDiskUsed
		stats.DiskTotal = sm.lastDiskTotal
		stats.Uptime = sm.lastUptime
	}

	if d, err := disk.IOCounters(); err == nil && len(d) > 0 {
		var totalRead, totalWrite uint64
		for _, io := range d {
			totalRead += io.ReadBytes
			totalWrite += io.WriteBytes
		}
		if !sm.lastDiskTime.IsZero() {
			dur := now.Sub(sm.lastDiskTime).Seconds()
			if dur > 0 {
				stats.DiskRead = float64(totalRead-sm.lastDiskRead) / 1024 / dur
				stats.DiskWrite = float64(totalWrite-sm.lastDiskWrite) / 1024 / dur
			}
		}
		sm.lastDiskRead = totalRead
		sm.lastDiskWrite = totalWrite
		sm.lastDiskTime = now
	}

	if sm.tickCount%10 == 0 || sm.lastPing == 0 {
		// Quick ping to Google DNS
		pOut, _ := ExecuteCommand("Test-Connection 8.8.8.8 -Count 1 -Quiet; (Test-Connection 8.8.8.8 -Count 1).ResponseTime")
		lines := strings.Split(strings.TrimSpace(pOut), "\n")
		if len(lines) >= 2 && strings.Contains(lines[0], "True") {
			stats.NetPing = ParseInt(lines[1])
			sm.lastPing = stats.NetPing
		} else {
			stats.NetPing = sm.lastPing
		}
	} else {
		stats.NetPing = sm.lastPing
	}

	if io, err := net.IOCounters(false); err == nil && len(io) > 0 {
		if !sm.lastNetTime.IsZero() {
			dur := now.Sub(sm.lastNetTime).Seconds()
			if dur > 0 {
				stats.NetSent = float64(io[0].BytesSent-sm.lastNetSent) / 1024 / dur
				stats.NetRecv = float64(io[0].BytesRecv-sm.lastNetRecv) / 1024 / dur
			}
		}
		stats.TotalNetSent = float64(io[0].BytesSent) / 1024 / 1024 / 1024 // GB
		stats.TotalNetRecv = float64(io[0].BytesRecv) / 1024 / 1024 / 1024 // GB
		sm.lastNetSent = io[0].BytesSent
		sm.lastNetRecv = io[0].BytesRecv
		sm.lastNetTime = now
	}

	pids, err := process.Pids()
	if err != nil {
		return stats
	}
	stats.TotalProcs = len(pids)

	sm.cacheMutex.Lock()
	if sm.tickCount%50 == 0 {
		currentPids := make(map[int32]bool)
		for _, pid := range pids {
			currentPids[pid] = true
		}
		for pid := range sm.procCache {
			if !currentPids[pid] {
				delete(sm.procCache, pid)
				delete(sm.nameCache, pid)
				delete(sm.cpuCache, pid)
				delete(sm.memCache, pid)
			}
		}
	}
	sm.cacheMutex.Unlock()

	results := make(chan ProcessInfo, len(pids))
	var wg sync.WaitGroup
	const maxWorkers = 12
	sem := make(chan struct{}, maxWorkers)

	for _, pid := range pids {
		wg.Add(1)
		go func(pID int32) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sm.cacheMutex.Lock()
			p, ok := sm.procCache[pID]
			if !ok {
				p, _ = process.NewProcess(pID)
				sm.procCache[pID] = p
			}
			name, nameOk := sm.nameCache[pID]
			lastCPU := sm.cpuCache[pID]
			lastMEM := sm.memCache[pID]
			sm.cacheMutex.Unlock()

			if p == nil {
				return
			}

			if !nameOk {
				name, _ = p.Name()
				if name == "" {
					return
				}
				sm.cacheMutex.Lock()
				sm.nameCache[pID] = name
				sm.cacheMutex.Unlock()
			}

			shouldPoll := lastCPU > 0.5 || sm.tickCount%5 == 0 || lastMEM == ""
			cpuVal := lastCPU
			memVal := lastMEM

			if shouldPoll {
				c, _ := p.CPUPercent()
				cpuVal = c

				if sm.tickCount%10 == 0 || lastMEM == "" || cpuVal > 1.0 {
					m, _ := p.MemoryInfo()
					if m != nil {
						memVal = fmt.Sprintf("%.1fMB", float64(m.RSS)/1024/1024)
					}
				}

				sm.cacheMutex.Lock()
				sm.cpuCache[pID] = cpuVal
				sm.memCache[pID] = memVal
				sm.cacheMutex.Unlock()
			}

			results <- ProcessInfo{
				ID:      fmt.Sprintf("%d", pID),
				Name:    name,
				CPU:     cpuVal,
				MEM:     memVal,
				Threads: 0,
			}
		}(pid)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var procInfos []ProcessInfo
	for r := range results {
		procInfos = append(procInfos, r)
	}

	sort.Slice(procInfos, func(i, j int) bool {
		return procInfos[i].CPU > procInfos[j].CPU
	})

	stats.Processes = procInfos
	return stats
}

// ParseInt robustly converts string to int.
func ParseInt(s string) int {
	var n int
	fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n
}

// RandomInt returns a random integer in [0, n) using math/rand.
func RandomInt(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}

// collectGPU queries nvidia-smi for NVIDIA GPU telemetry.
// Returns an empty GPUInfo{Available:false} when nvidia-smi is absent or fails.
func collectGPU() GPUInfo {
	out, err := ExecuteCommand(
		`nvidia-smi --query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits`,
	)
	if err != nil || strings.TrimSpace(out) == "" {
		return GPUInfo{}
	}
	fields := strings.Split(strings.TrimSpace(out), ",")
	if len(fields) < 6 {
		return GPUInfo{}
	}
	var util, vramUsed, vramTotal, temp, power float64
	fmt.Sscanf(strings.TrimSpace(fields[1]), "%f", &util)
	fmt.Sscanf(strings.TrimSpace(fields[2]), "%f", &vramUsed)
	fmt.Sscanf(strings.TrimSpace(fields[3]), "%f", &vramTotal)
	fmt.Sscanf(strings.TrimSpace(fields[4]), "%f", &temp)
	fmt.Sscanf(strings.TrimSpace(fields[5]), "%f", &power)
	return GPUInfo{
		Name:        strings.TrimSpace(fields[0]),
		Utilization: util,
		VRAMUsed:    vramUsed,
		VRAMTotal:   vramTotal,
		Temp:        temp,
		Power:       power,
		Available:   true,
	}
}

// collectCPUTemps queries WMI for CPU package temperatures via PowerShell.
// On systems where WMI thermal sensors are unavailable it returns nil.
func collectCPUTemps() []float64 {
	out, err := ExecuteCommand(
		`(Get-CimInstance -Namespace root/WMI -ClassName MSAcpi_ThermalZoneTemperature).CurrentTemperature | ForEach-Object { [math]::Round(($_ - 2732) / 10.0, 1) }`,
	)
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	var temps []float64
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var t float64
		if _, err := fmt.Sscanf(line, "%f", &t); err == nil && t > 0 && t < 120 {
			temps = append(temps, t)
		}
	}
	return temps
}
