package r2d2

import (
	"fmt"
	"sort"
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

// SysStats aggregates system-wide metrics and a list of active processes.
type SysStats struct {
	CPU          float64
	RAM          float64
	RAMUsed      float64
	RAMTotal     float64
	Disk         float64
	Uptime       string
	Processes    []ProcessInfo
	TotalProcs   int
	TotalThreads int64
	NetSent      float64
	NetRecv      float64
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

	lastDisk   float64
	lastUptime string

	lastNetRecv uint64
	lastNetSent uint64
	lastNetTime time.Time
}

// NewStatsManager initializes a new telemetry provider with empty caches.
func NewStatsManager() *StatsManager {
	return &StatsManager{
		procCache:  make(map[int32]*process.Process),
		nameCache:  make(map[int32]string),
		cpuCache:   make(map[int32]float64),
		memCache:   make(map[int32]string),
	}
}

// GetStats collects current system metrics and process telemetry.
func (sm *StatsManager) GetStats() SysStats {
	sm.tickCount++
	stats := SysStats{
		Disk:   sm.lastDisk,
		Uptime: sm.lastUptime,
	}
	if stats.Uptime == "" {
		stats.Uptime = "0d 0h 0m"
	}

	if v, err := mem.VirtualMemory(); err == nil && v != nil {
		stats.RAM = v.UsedPercent
		stats.RAMUsed = float64(v.Used) / 1024 / 1024 / 1024
		stats.RAMTotal = float64(v.Total) / 1024 / 1024 / 1024
	}
	if c, err := cpu.Percent(0, false); err == nil && len(c) > 0 {
		stats.CPU = c[0]
	}

	now := time.Now()
	if sm.tickCount%10 == 0 || sm.lastRefresh.IsZero() {
		if d, err := disk.Usage("C:"); err == nil && d != nil {
			sm.lastDisk = d.UsedPercent
			stats.Disk = sm.lastDisk
		}
		if u, err := host.Uptime(); err == nil {
			sm.lastUptime = fmt.Sprintf("%dd %dh %dm", u/86400, (u%86400)/3600, (u%3600)/60)
			stats.Uptime = sm.lastUptime
		}
		sm.lastRefresh = now
	}

	if io, err := net.IOCounters(false); err == nil && len(io) > 0 {
		if !sm.lastNetTime.IsZero() {
			dur := now.Sub(sm.lastNetTime).Seconds()
			if dur > 0 {
				stats.NetSent = float64(io[0].BytesSent-sm.lastNetSent) / 1024 / dur
				stats.NetRecv = float64(io[0].BytesRecv-sm.lastNetRecv) / 1024 / dur
			}
		}
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
