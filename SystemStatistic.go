package ovpm

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SystemStatus представляет данные о состоянии системы
type SystemStatus struct {
	CPUUsage    float64
	MemoryTotal uint64
	MemoryUsed  uint64
	SwapTotal   uint64
	SwapUsed    uint64
	LoadAverage []float64
	DiskUsage   []DiskUsage
	Timestamp   time.Time
}

// DiskUsage представляет использование диска
type DiskUsage struct {
	Mount          string
	Total          uint64
	Used           uint64
	UsedPercentage float64
}

// GetSystemStatus возвращает полную статистику системы
func GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		Timestamp: time.Now(),
	}

	// Получаем использование CPU
	cpuUsage, err := getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	status.CPUUsage = cpuUsage

	// Получаем использование памяти
	memTotal, memUsed, swapTotal, swapUsed, err := getMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %v", err)
	}
	status.MemoryTotal = memTotal
	status.MemoryUsed = memUsed
	status.SwapTotal = swapTotal
	status.SwapUsed = swapUsed

	// Получаем load average
	loadAvg, err := getLoadAverage()
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %v", err)
	}
	status.LoadAverage = loadAvg

	// Получаем использование дисков
	diskUsage, err := getDiskUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}
	status.DiskUsage = diskUsage

	return status, nil
}

// getCPUUsage возвращает использование CPU в процентах (0.0 - 1.0)
func getCPUUsage() (float64, error) {
	// Читаем статистику CPU из /proc/stat
	content, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 8 {
				return 0, fmt.Errorf("invalid cpu stat line")
			}

			var total, idle uint64
			for i := 1; i < len(fields); i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					return 0, err
				}
				total += val
				if i == 4 {
					idle = val
				}
			}

			// Вычисляем использование
			if total > 0 {
				return 1.0 - float64(idle)/float64(total), nil
			}
		}
	}

	return 0, fmt.Errorf("cpu stat not found")
}

// getMemoryUsage возвращает информацию об использовании памяти
func getMemoryUsage() (uint64, uint64, uint64, uint64, error) {
	// Читаем информацию о памяти из /proc/meminfo
	content, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, 0, err
	}

	var memTotal, memAvailable, swapTotal, swapFree uint64
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal = parseMemInfoValue(fields[1])
		case "MemAvailable:":
			memAvailable = parseMemInfoValue(fields[1])
		case "SwapTotal:":
			swapTotal = parseMemInfoValue(fields[1])
		case "SwapFree:":
			swapFree = parseMemInfoValue(fields[1])
		}
	}

	memUsed := memTotal - memAvailable
	swapUsed := swapTotal - swapFree

	return memTotal * 1024, memUsed * 1024, swapTotal * 1024, swapUsed * 1024, nil
}

// parseMemInfoValue парсит значения из /proc/meminfo
func parseMemInfoValue(value string) uint64 {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// getLoadAverage возвращает load average
func getLoadAverage() ([]float64, error) {
	content, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid loadavg format")
	}

	loadAvg := make([]float64, 3)
	for i := 0; i < 3; i++ {
		load, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, err
		}
		loadAvg[i] = load
	}

	return loadAvg, nil
}

// getDiskUsage возвращает использование дисков
func getDiskUsage() ([]DiskUsage, error) {
	//content, err := ioutil.ReadFile("/proc/mounts")
	//if err != nil {
	//	return nil, err
	//}

	var diskUsage []DiskUsage
	//lines := strings.Split(string(content), "\n")

	mounts := []string{"/"} //, "/home", "/var"}

	for _, mount := range mounts {
		var stat syscall.Statfs_t
		err := syscall.Statfs(mount, &stat)
		if err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		used := total - free
		usedPercentage := float64(used) / float64(total)

		diskUsage = append(diskUsage, DiskUsage{
			Mount:          mount,
			Total:          total,
			Used:           used,
			UsedPercentage: usedPercentage,
		})
	}

	return diskUsage, nil
}
