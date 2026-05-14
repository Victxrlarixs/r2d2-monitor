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
	ID       string
	Name     string
	CPU      float64
	MEM      string
	MemBytes uint64 // Numeric value for accurate sorting
	Threads  int64
}

// GPUInfo holds telemetry for NVIDIA and generic GPUs.
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

// NetStats holds delta state for network calculations
type netDelta struct {
	lastSent uint64
	lastRecv uint64
	lastTime time.Time
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
	procCache     map[int32]*process.Process
	nameCache     map[int32]string
	cpuCache      map[int32]float64
	memCache      map[int32]string
	memBytesCache map[int32]uint64
	cacheMutex    sync.Mutex
	lastRefresh   time.Time
	tickCount     uint64

	lastDiskRead  uint64
	lastDiskWrite uint64
	lastDiskTime  time.Time
	lastPing      int

	// Per-interface network history
	netDeltas  map[string]netDelta
	deltaMutex sync.Mutex

	lastGPU      GPUInfo
	lastCPUTemps []float64

	lastInventoryTick uint64
	cachedAllDisks    []string
	cachedAllNet      []string

	bgMutex       sync.RWMutex
	bgBattery     BatteryInfo
	bgCPUTemps    []float64
	bgGPU         GPUInfo
	bgDiskUsage   float64
	bgDiskUsed    float64
	bgDiskTotal   float64
	bgUptime      string
	bgLoopCounter uint64
}

func NewStatsManager() *StatsManager {
	sm := &StatsManager{
		procCache:     make(map[int32]*process.Process),
		nameCache:     make(map[int32]string),
		cpuCache:      make(map[int32]float64),
		memCache:      make(map[int32]string),
		memBytesCache: make(map[int32]uint64),
		netDeltas:     make(map[string]netDelta),
	}
	go sm.backgroundPollLoop()
	return sm
}

func (sm *StatsManager) backgroundPollLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		sm.bgLoopCounter++
		
		var bat BatteryInfo
		if sm.bgLoopCounter%2 == 0 {
			var dstBat []Win32_Battery
			if err := wmi.Query("SELECT EstimatedChargeRemaining, BatteryStatus FROM Win32_Battery", &dstBat); err == nil && len(dstBat) > 0 {
				bat.Percent = float64(dstBat[0].EstimatedChargeRemaining)
				statusMap := map[uint16]string{1: "Discharging", 2: "AC Power", 3: "Fully Charged", 6: "Charging"}
				bat.Status = statusMap[dstBat[0].BatteryStatus]
				if bat.Status == "" { bat.Status = "Unknown" }
			}
		}

		temps := collectCPUTemps()
		
		gpu := sm.bgGPU
		if sm.bgLoopCounter%5 == 0 || sm.bgLoopCounter == 1 {
			newGpu := collectGPU()
			if newGpu.Available || sm.bgLoopCounter == 1 { gpu = newGpu }
		}

		dUsage, dUsed, dTotal, uptime := sm.bgDiskUsage, sm.bgDiskUsed, sm.bgDiskTotal, sm.bgUptime
		if sm.bgLoopCounter%10 == 0 || sm.bgLoopCounter == 1 {
			d, _ := disk.Usage("C:")
			if d != nil {
				dUsage = d.UsedPercent
				dUsed = float64(d.Used) / 1024 / 1024 / 1024
				dTotal = float64(d.Total) / 1024 / 1024 / 1024
			}
			if u, err := host.Uptime(); err == nil {
				uptime = fmt.Sprintf("%dd %dh %dm", u/86400, (u%86400)/3600, (u%3600)/60)
			}
		}

		sm.bgMutex.Lock()
		if bat.Status != "" { sm.bgBattery = bat }
		sm.bgCPUTemps = temps
		sm.bgGPU = gpu
		sm.bgDiskUsage = dUsage
		sm.bgDiskUsed = dUsed
		sm.bgDiskTotal = dTotal
		sm.bgUptime = uptime
		sm.bgMutex.Unlock()
	}
}

func (sm *StatsManager) GetStats(priorityPIDs []string, cfg Config) SysStats {
	sm.tickCount++
	
	sm.bgMutex.RLock()
	stats := SysStats{
		Disk: sm.bgDiskUsage, DiskUsed: sm.bgDiskUsed, DiskTotal: sm.bgDiskTotal,
		Uptime: sm.bgUptime, SelectedDisk: cfg.SelectedDisk, SelectedNet: cfg.SelectedNetInt,
		Battery: sm.bgBattery, CPUTemps: sm.bgCPUTemps, GPU: sm.bgGPU,
	}
	sm.bgMutex.RUnlock()

	if sm.tickCount-sm.lastInventoryTick >= 60 || sm.lastInventoryTick == 0 {
		if parts, err := disk.Partitions(false); err == nil {
			sm.cachedAllDisks = nil
			for _, p := range parts { if p.Mountpoint != "" { sm.cachedAllDisks = append(sm.cachedAllDisks, p.Mountpoint) } }
		}
		if interfaces, err := net.Interfaces(); err == nil {
			sm.cachedAllNet = nil
			for _, i := range interfaces {
				if len(i.Addrs) > 0 && !strings.Contains(i.Name, "Loopback") {
					sm.cachedAllNet = append(sm.cachedAllNet, i.Name)
				}
			}
		}
		sm.lastInventoryTick = sm.tickCount
	}
	stats.AllDisks, stats.AllNet = sm.cachedAllDisks, sm.cachedAllNet

	if stats.SelectedDisk == "" && len(stats.AllDisks) > 0 { stats.SelectedDisk = stats.AllDisks[0] }
	if stats.SelectedNet == "" && len(stats.AllNet) > 0 { stats.SelectedNet = stats.AllNet[0] }

	if v, _ := mem.VirtualMemory(); v != nil {
		stats.RAM, stats.RAMUsed, stats.RAMTotal = v.UsedPercent, float64(v.Used)/1e9, float64(v.Total)/1e9
		stats.RAMAvailable, stats.RAMCached = float64(v.Available)/1e9, float64(v.Cached)/1e9
	}
	if s, _ := mem.SwapMemory(); s != nil {
		stats.Swap, stats.SwapUsed, stats.SwapTotal = s.UsedPercent, float64(s.Used)/1e9, float64(s.Total)/1e9
	}
	if c, err := cpu.Percent(0, false); err == nil && len(c) > 0 { stats.CPU = c[0] }
	if cores, err := cpu.Percent(0, true); err == nil { stats.CPUCores = cores }
	if info, err := cpu.Info(); err == nil && len(info) > 0 { stats.CPUModel = info[0].ModelName }
	if h, err := host.Info(); err == nil { stats.OSName = fmt.Sprintf("%s %s", h.Platform, h.PlatformVersion) }

	if n, err := net.Interfaces(); err == nil {
		for _, i := range n {
			for _, a := range i.Addrs {
				if strings.Contains(a.Addr, ".") && !strings.HasPrefix(a.Addr, "127.") {
					stats.LocalIP = strings.Split(a.Addr, "/")[0]
					break
				}
			}
			if stats.LocalIP != "" { break }
		}
	}

	now := time.Now()
	// Network Delta Calculation
	// If no interface selected, we use the SUM (IOCounters(false))
	useSum := cfg.SelectedNetInt == ""
	io, err := net.IOCounters(!useSum)
	if err == nil && len(io) > 0 {
		var activeIO *net.IOCountersStat
		if useSum {
			activeIO = &io[0]
		} else {
			for _, ni := range io {
				if ni.Name == cfg.SelectedNetInt { activeIO = &ni; break }
			}
			if activeIO == nil { activeIO = &io[0] }
		}

		sm.deltaMutex.Lock()
		history, ok := sm.netDeltas[activeIO.Name]
		if ok && !history.lastTime.IsZero() {
			dur := now.Sub(history.lastTime).Seconds()
			if dur > 0.1 {
				stats.NetSent = float64(activeIO.BytesSent-history.lastSent) / 1024 / dur
				stats.NetRecv = float64(activeIO.BytesRecv-history.lastRecv) / 1024 / dur
			}
		}
		sm.netDeltas[activeIO.Name] = netDelta{
			lastSent: activeIO.BytesSent,
			lastRecv: activeIO.BytesRecv,
			lastTime: now,
		}
		sm.deltaMutex.Unlock()
		stats.TotalNetSent, stats.TotalNetRecv = float64(activeIO.BytesSent)/1e9, float64(activeIO.BytesRecv)/1e9
	}

	if d, err := disk.IOCounters(); err == nil && len(d) > 0 {
		var tr, tw uint64
		for _, i := range d { tr += i.ReadBytes; tw += i.WriteBytes }
		if !sm.lastDiskTime.IsZero() {
			dur := now.Sub(sm.lastDiskTime).Seconds()
			if dur > 0.1 {
				stats.DiskRead = float64(tr-sm.lastDiskRead) / 1024 / dur
				stats.DiskWrite = float64(tw-sm.lastDiskWrite) / 1024 / dur
			}
		}
		sm.lastDiskRead, sm.lastDiskWrite, sm.lastDiskTime = tr, tw, now
	}

	if sm.tickCount%10 == 0 || sm.lastPing == 0 {
		start := time.Now()
		conn, err := stdnet.DialTimeout("tcp", "8.8.8.8:53", 1*time.Second)
		if err == nil { stats.NetPing = int(time.Since(start).Milliseconds()); sm.lastPing = stats.NetPing; conn.Close() } else { stats.NetPing = sm.lastPing }
	} else { stats.NetPing = sm.lastPing }

	procs, _ := process.Processes()
	stats.TotalProcs = len(procs)

	sm.cacheMutex.Lock()
	if sm.tickCount%50 == 0 {
		curr := make(map[int32]bool)
		for _, p := range procs { curr[p.Pid] = true }
		for pid := range sm.nameCache { if !curr[pid] { delete(sm.nameCache, pid); delete(sm.cpuCache, pid); delete(sm.memCache, pid); delete(sm.memBytesCache, pid) } }
	}
	sm.cacheMutex.Unlock()

	priorityMap := make(map[int32]bool)
	for _, pStr := range priorityPIDs {
		var pID int32
		if _, err := fmt.Sscanf(pStr, "%d", &pID); err == nil { priorityMap[pID] = true }
	}

	results := make(chan ProcessInfo, len(procs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)
	for _, p := range procs {
		wg.Add(1)
		go func(proc *process.Process) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			pID := proc.Pid
			sm.cacheMutex.Lock()
			name, nameOk := sm.nameCache[pID]
			lastCPU, lastMEM, lastMEMBytes := sm.cpuCache[pID], sm.memCache[pID], sm.memBytesCache[pID]
			sm.cacheMutex.Unlock()
			if !nameOk { name, _ = proc.Name(); if name == "" { return }; sm.cacheMutex.Lock(); sm.nameCache[pID] = name; sm.cacheMutex.Unlock() }
			isPriority := priorityMap[pID]
			shouldPollCPU := isPriority || lastMEM == "" || (lastCPU > 1.0 && sm.tickCount%2 == 0) || sm.tickCount%10 == 0
			shouldPollMem := isPriority || lastMEM == "" || sm.tickCount%20 == 0 || (lastCPU > 1.0 && sm.tickCount%5 == 0)
			cv, mv, mb := lastCPU, lastMEM, lastMEMBytes
			if shouldPollCPU { cv, _ = proc.CPUPercent(); sm.cacheMutex.Lock(); sm.cpuCache[pID] = cv; sm.cacheMutex.Unlock() }
			if shouldPollMem { m, _ := proc.MemoryInfo(); if m != nil { mb = m.RSS; mv = fmt.Sprintf("%.1fMB", float64(m.RSS)/1e6) }; sm.cacheMutex.Lock(); sm.memCache[pID] = mv; sm.memBytesCache[pID] = mb; sm.cacheMutex.Unlock() }
			results <- ProcessInfo{ID: fmt.Sprintf("%d", pID), Name: name, CPU: cv, MEM: mv, MemBytes: mb}
		}(p)
	}
	go func() { wg.Wait(); close(results) }()
	for r := range results { stats.Processes = append(stats.Processes, r) }
	sort.Slice(stats.Processes, func(i, j int) bool {
		if strings.EqualFold(cfg.SortBy, "mem") { return stats.Processes[i].MemBytes > stats.Processes[j].MemBytes }
		return stats.Processes[i].CPU > stats.Processes[j].CPU
	})
	return stats
}

func RandomInt(n int) int { if n <= 0 { return 0 }; return rand.Intn(n) }

func collectGPU() GPUInfo {
	out, err := ExecuteCommand(`nvidia-smi --query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw --format=csv,noheader,nounits`)
	if err == nil && strings.TrimSpace(out) != "" {
		f := strings.Split(strings.TrimSpace(out), ",")
		if len(f) >= 6 {
			var util, vramUsed, vramTotal, temp, power float64
			fmt.Sscanf(strings.TrimSpace(f[1]), "%f", &util); fmt.Sscanf(strings.TrimSpace(f[2]), "%f", &vramUsed)
			fmt.Sscanf(strings.TrimSpace(f[3]), "%f", &vramTotal); fmt.Sscanf(strings.TrimSpace(f[4]), "%f", &temp)
			fmt.Sscanf(strings.TrimSpace(f[5]), "%f", &power)
			return GPUInfo{Name: strings.TrimSpace(f[0]), Utilization: util, VRAMUsed: vramUsed, VRAMTotal: vramTotal, Temp: temp, Power: power, Available: true}
		}
	}
	var dst []Win32_VideoController
	if err = wmi.Query("SELECT Name, AdapterRAM FROM Win32_VideoController", &dst); err == nil && len(dst) > 0 {
		return GPUInfo{Name: dst[0].Name, VRAMTotal: float64(dst[0].AdapterRAM) / 1e6, Available: true}
	}
	return GPUInfo{}
}

func collectCPUTemps() []float64 {
	var dst []MSAcpi_ThermalZoneTemperature
	if err := wmi.QueryNamespace("SELECT CurrentTemperature FROM MSAcpi_ThermalZoneTemperature", &dst, "root\\WMI"); err != nil || len(dst) == 0 { return nil }
	var temps []float64
	for _, tz := range dst {
		t := (float64(tz.CurrentTemperature) - 2732.0) / 10.0
		if t > 0 && t < 120 { temps = append(temps, t) }
	}
	return temps
}
