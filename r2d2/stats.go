package r2d2

import (
	"fmt"
	"math/rand"
	stdnet "net"
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
	"github.com/yusufpapurcu/wmi"
)

// Win32_Battery WMI struct
type Win32_Battery struct {
	EstimatedChargeRemaining uint16
	BatteryStatus            uint16
}

// MSAcpi_ThermalZoneTemperature WMI struct
type MSAcpi_ThermalZoneTemperature struct {
	CurrentTemperature uint32
}

// Win32_VideoController WMI struct for generic GPU info
type Win32_VideoController struct {
	Name       string
	AdapterRAM uint64
}

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
	SelectedDisk string
	AllDisks     []string
	SelectedNet  string
	AllNet       []string
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
// priorityPIDs are PIDs that should be polled every tick (e.g., currently visible in UI).
func (sm *StatsManager) GetStats(priorityPIDs []string, cfg Config) SysStats {
	sm.tickCount++
	stats := SysStats{
		Disk:         sm.lastDisk,
		DiskUsed:     sm.lastDiskUsed,
		DiskTotal:    sm.lastDiskTotal,
		Uptime:       sm.lastUptime,
		SelectedDisk: cfg.SelectedDisk,
		SelectedNet:  cfg.SelectedNetInt,
	}

	// Inventory Disks
	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			if p.Mountpoint != "" {
				stats.AllDisks = append(stats.AllDisks, p.Mountpoint)
			}
		}
	}
	if stats.SelectedDisk == "" && len(stats.AllDisks) > 0 {
		stats.SelectedDisk = stats.AllDisks[0]
	}

	// Inventory Net
	if interfaces, err := net.Interfaces(); err == nil {
		for _, i := range interfaces {
			// Only include interfaces that have an IP address and are not loopback
			if len(i.Addrs) > 0 && !strings.Contains(i.Name, "Loopback") {
				stats.AllNet = append(stats.AllNet, i.Name)
			}
		}
	}
	if stats.SelectedNet == "" && len(stats.AllNet) > 0 {
		stats.SelectedNet = stats.AllNet[0]
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

	// Battery via WMI
	var dstBat []Win32_Battery
	if err := wmi.Query("SELECT EstimatedChargeRemaining, BatteryStatus FROM Win32_Battery", &dstBat); err == nil && len(dstBat) > 0 {
		stats.Battery.Percent = float64(dstBat[0].EstimatedChargeRemaining)
		statusMap := map[uint16]string{1: "Discharging", 2: "AC Power", 3: "Fully Charged", 6: "Charging"}
		stats.Battery.Status = statusMap[dstBat[0].BatteryStatus]
		if stats.Battery.Status == "" {
			stats.Battery.Status = "Unknown"
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
		targetDisk := stats.SelectedDisk
		if targetDisk == "" { targetDisk = "C:" }
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
		// Quick TCP ping to Google DNS
		startPing := time.Now()
		conn, err := stdnet.DialTimeout("tcp", "8.8.8.8:53", 1*time.Second)
		if err == nil && conn != nil {
			stats.NetPing = int(time.Since(startPing).Milliseconds())
			sm.lastPing = stats.NetPing
			conn.Close()
		} else {
			stats.NetPing = sm.lastPing
		}
	} else {
		stats.NetPing = sm.lastPing
	}

	if io, err := net.IOCounters(true); err == nil && len(io) > 0 {
		var selectedIO *net.IOCountersStat
		for _, ni := range io {
			if ni.Name == stats.SelectedNet {
				selectedIO = &ni
				break
			}
		}
		// Fallback to first if selected not found
		if selectedIO == nil { selectedIO = &io[0] }

		if !sm.lastNetTime.IsZero() {
			dur := now.Sub(sm.lastNetTime).Seconds()
			if dur > 0 {
				stats.NetSent = float64(selectedIO.BytesSent-sm.lastNetSent) / 1024 / dur
				stats.NetRecv = float64(selectedIO.BytesRecv-sm.lastNetRecv) / 1024 / dur
			}
		}
		stats.TotalNetSent = float64(selectedIO.BytesSent) / 1024 / 1024 / 1024 // GB
		stats.TotalNetRecv = float64(selectedIO.BytesRecv) / 1024 / 1024 / 1024 // GB
		sm.lastNetSent = selectedIO.BytesSent
		sm.lastNetRecv = selectedIO.BytesRecv
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
	
	// Convert priorityPIDs to a map for O(1) lookup
	priorityMap := make(map[int32]bool)
	for _, pStr := range priorityPIDs {
		var pID int32
		if _, err := fmt.Sscanf(pStr, "%d", &pID); err == nil {
			priorityMap[pID] = true
		}
	}

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

			// OPTIMIZATION: Throttled Polling Logic
			// 1. Priority PIDs (visible) poll every tick.
			// 2. Active processes (CPU > 1%) poll every 2 ticks.
			// 3. New processes (no lastMEM) poll once immediately.
			// 4. Others poll every 10 ticks.
			isPriority := priorityMap[pID]
			isActive := lastCPU > 1.0
			isNew := lastMEM == ""
			
			shouldPollCPU := isPriority || isNew || (isActive && sm.tickCount%2 == 0) || sm.tickCount%10 == 0
			shouldPollMem := isPriority || isNew || sm.tickCount%20 == 0 || (isActive && sm.tickCount%5 == 0)
			
			cpuVal := lastCPU
			memVal := lastMEM

			if shouldPollCPU {
				c, _ := p.CPUPercent()
				cpuVal = c
				
				sm.cacheMutex.Lock()
				sm.cpuCache[pID] = cpuVal
				sm.cacheMutex.Unlock()
			}
			
			if shouldPollMem {
				m, _ := p.MemoryInfo()
				if m != nil {
					memVal = fmt.Sprintf("%.1fMB", float64(m.RSS)/1024/1024)
				}
				
				sm.cacheMutex.Lock()
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

// collectGPU queries NVIDIA via nvidia-smi with fallback to WMI for AMD/Intel.
func collectGPU() GPUInfo {
	// 1. Try NVIDIA-SMI
	out, err := ExecuteCommand(
		`nvidia-smi --query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits`,
	)
	if err == nil && strings.TrimSpace(out) != "" {
		fields := strings.Split(strings.TrimSpace(out), ",")
		if len(fields) >= 6 {
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
	}

	// 2. Fallback to WMI for generic GPU info (AMD/Intel)
	var dst []Win32_VideoController
	err = wmi.Query("SELECT Name, AdapterRAM FROM Win32_VideoController", &dst)
	if err == nil && len(dst) > 0 {
		return GPUInfo{
			Name:      dst[0].Name,
			VRAMTotal: float64(dst[0].AdapterRAM) / 1024 / 1024,
			Available: true,
		}
	}

	return GPUInfo{}
}

// collectCPUTemps queries WMI for CPU package temperatures directly.
// On systems where WMI thermal sensors are unavailable it returns nil.
func collectCPUTemps() []float64 {
	var dstTemp []MSAcpi_ThermalZoneTemperature
	err := wmi.QueryNamespace("SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature", &dstTemp, "root\\WMI")
	if err != nil || len(dstTemp) == 0 {
		return nil
	}
	var temps []float64
	for _, tz := range dstTemp {
		// Convert from 10ths of degrees Kelvin to Celsius
		t := (float64(tz.CurrentTemperature) - 2732.0) / 10.0
		if t > 0 && t < 120 {
			temps = append(temps, t)
		}
	}
	return temps
}
