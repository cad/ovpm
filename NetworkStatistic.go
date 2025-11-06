package ovpm

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NetworkInterface представляет информацию о сетевом интерфейсе
type NetworkInterface struct {
	Name string
	IP   string
	MAC  string
	IsUp bool
}

// NetworkInterfacesResponse представляет ответ со списком интерфейсов
type NetworkInterfacesResponse struct {
	Interfaces []NetworkInterface
}

// GetNetworkInterfaces возвращает список сетевых интерфейсов
func GetNetworkInterfaces() (*NetworkInterfacesResponse, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	response := &NetworkInterfacesResponse{
		Interfaces: make([]NetworkInterface, 0),
	}

	for _, iface := range interfaces {
		// Пропускаем loopback интерфейсы
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Получаем IP адреса интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ip string
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip = ipnet.IP.String()
					break
				}
			}
		}

		// Если у интерфейса нет IPv4 адреса, пропускаем
		if ip == "" {
			continue
		}

		networkInterface := NetworkInterface{
			Name: iface.Name,
			IP:   ip,
			MAC:  iface.HardwareAddr.String(),
			IsUp: iface.Flags&net.FlagUp != 0,
		}

		response.Interfaces = append(response.Interfaces, networkInterface)
	}

	return response, nil
}

// InterfaceStats представляет статистику сетевого интерфейса
type InterfaceStats struct {
	InterfaceName string    `json:"interface_name"`
	TXBytes       uint64    `json:"tx_bytes"`
	RXBytes       uint64    `json:"rx_bytes"`
	TXBytesPerSec uint64    `json:"tx_bytes_per_sec"`
	RXBytesPerSec uint64    `json:"rx_bytes_per_sec"`
	TXPackets     uint64    `json:"tx_packets"`
	RXPackets     uint64    `json:"rx_packets"`
	Timestamp     time.Time `json:"timestamp"`
}

// InterfaceStatsResponse представляет ответ со статистикой интерфейса
type InterfaceStatsResponse struct {
	Stats InterfaceStats `json:"stats"`
}

// Глобальные переменные для хранения предыдущих значений
var (
	prevStats     = make(map[string]InterfaceStats)
	statsMutex    sync.RWMutex
	lastCheckTime = make(map[string]time.Time)
)

// GetInterfaceStats возвращает статистику для указанного интерфейса
func GetInterfaceStats(interfaceName string) (*InterfaceStatsResponse, error) {
	currentStats, err := getCurrentInterfaceStats(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for interface %s: %v", interfaceName, err)
	}

	// Вычисляем скорость передачи данных
	currentStats = calculateDataRates(interfaceName, currentStats)

	response := &InterfaceStatsResponse{
		Stats: currentStats,
	}

	return response, nil
}

// getCurrentInterfaceStats читает текущую статистику из /proc/net/dev
func getCurrentInterfaceStats(interfaceName string) (InterfaceStats, error) {
	stats := InterfaceStats{
		InterfaceName: interfaceName,
		Timestamp:     time.Now(),
	}

	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Ищем строку с нашим интерфейсом
		if strings.Contains(line, interfaceName+":") {
			fields := strings.Fields(line)
			if len(fields) < 17 {
				return stats, fmt.Errorf("invalid interface stats format")
			}

			// Парсим полученные байты и пакеты (поля 1 и 2)
			rxBytes, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return stats, err
			}
			rxPackets, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				return stats, err
			}

			// Парсим отправленные байты и пакеты (поля 9 и 10)
			txBytes, err := strconv.ParseUint(fields[9], 10, 64)
			if err != nil {
				return stats, err
			}
			txPackets, err := strconv.ParseUint(fields[10], 10, 64)
			if err != nil {
				return stats, err
			}

			stats.RXBytes = rxBytes
			stats.RXPackets = rxPackets
			stats.TXBytes = txBytes
			stats.TXPackets = txPackets

			return stats, nil
		}
	}

	return stats, fmt.Errorf("interface %s not found", interfaceName)
}

// calculateDataRates вычисляет скорость передачи данных в байтах/сек
func calculateDataRates(interfaceName string, currentStats InterfaceStats) InterfaceStats {
	statsMutex.Lock()
	defer statsMutex.Unlock()

	prev, exists := prevStats[interfaceName]
	lastTime, timeExists := lastCheckTime[interfaceName]

	// Сохраняем текущие значения для следующего вызова
	prevStats[interfaceName] = currentStats
	lastCheckTime[interfaceName] = currentStats.Timestamp

	// Если предыдущих значений нет, возвращаем нулевые скорости
	if !exists || !timeExists {
		return currentStats
	}

	// Вычисляем разницу во времени
	timeDiff := currentStats.Timestamp.Sub(lastTime).Seconds()
	if timeDiff <= 0 {
		return currentStats
	}

	// Вычисляем скорости (байт/сек)
	if currentStats.TXBytes > prev.TXBytes {
		currentStats.TXBytesPerSec = uint64(float64(currentStats.TXBytes-prev.TXBytes) / timeDiff)
	}

	if currentStats.RXBytes > prev.RXBytes {
		currentStats.RXBytesPerSec = uint64(float64(currentStats.RXBytes-prev.RXBytes) / timeDiff)
	}

	return currentStats
}

// GetAvailableInterfaces возвращает список доступных интерфейсов из /proc/net/dev
func GetAvailableInterfaces() ([]string, error) {
	content, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}

	var interfaces []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				// Убираем двоеточие из имени интерфейса
				ifaceName := strings.TrimSuffix(fields[0], ":")
				// Пропускаем заголовки и loopback интерфейсы
				if ifaceName != "face" && ifaceName != "lo" {
					interfaces = append(interfaces, ifaceName)
				}
			}
		}
	}

	return interfaces, nil
}
