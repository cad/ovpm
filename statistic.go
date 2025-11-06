package ovpm

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jinzhu/gorm"
)

var (
	globalFileWatcher *FileWatcher
	fileWatcherOnce   sync.Once
	fileWatcherMu     sync.RWMutex
)

type FileWatcher struct {
	filename    string
	watcher     *fsnotify.Watcher
	lastProcess time.Time
	debounce    time.Duration
	mu          sync.RWMutex
	data        []clEntry
	statistic   []SpeedStat
	lastUpdate  time.Time
}

type SpeedStat struct {
	commonName string
	tx         float32
	rx         float32
}

type dbStatisticModel struct {
	gorm.Model
	UserID         uint
	ConnectedSince time.Time
	ConnectedUntil time.Time
	BytesReceived  uint64
	BytesSent      uint64
	CommonName     string
	RealAddress    string
}

// GetFileWatcher возвращает глобальный экземпляр FileWatcher
func GetFileWatcher() *FileWatcher {
	fileWatcherMu.RLock()
	defer fileWatcherMu.RUnlock()
	return globalFileWatcher
}

// InitializeFileWatcher инициализирует глобальный FileWatcher
func InitializeFileWatcher() error {
	var initErr error
	fileWatcherOnce.Do(func() {
		fw, err := NewStatisticFileWatcher()
		if err != nil {
			initErr = err
			return
		}
		globalFileWatcher = fw
		go fw.Watch()
	})
	return initErr
}

func NewFileWatcher(filename string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(filename)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		filename: filename,
		watcher:  watcher,
		debounce: 100 * time.Millisecond,
	}, nil
}

func NewStatisticFileWatcher() (*FileWatcher, error) {
	return NewFileWatcher(_DefaultStatusLogPath)
}

func (fw *FileWatcher) Watch() {
	var debounceTimer *time.Timer

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}

				debounceTimer = time.AfterFunc(fw.debounce, func() {
					fw.processFile()
				})
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func (fw *FileWatcher) processFile() {
	if time.Since(fw.lastProcess) < fw.debounce {
		return
	}
	fw.getNewValue()
	fw.lastProcess = time.Now()
}

func (fw *FileWatcher) Close() {
	err := fw.watcher.Close()
	if err != nil {
		return
	}
}

func (fw *FileWatcher) getNewValue() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	data, lastUpdate := ConnectionList()

	if fw.lastUpdate.IsZero() {
		fw.data = data
		fw.lastUpdate = lastUpdate
		return
	}

	dt := lastUpdate.Sub(fw.lastUpdate)
	if dt <= 0 {
		dt = time.Second
	}
	seconds := dt.Seconds()
	if seconds == 0 {
		seconds = 1
	}
	fw.lastUpdate = lastUpdate

	clMap := make(map[string]clEntry)
	for _, el := range fw.data {
		clMap[el.CommonName] = el
	}

	statIndexMap := make(map[string]int)
	for i, stat := range fw.statistic {
		statIndexMap[stat.commonName] = i
	}

	activeUsers := make(map[string]bool)
	for _, entry := range data {
		activeUsers[entry.CommonName] = true

		var tx, rx float32
		if oldEntry, exists := clMap[entry.CommonName]; exists {
			tx = (float32(entry.BytesSent) - float32(oldEntry.BytesSent)) / float32(seconds)
			rx = (float32(entry.BytesReceived) - float32(oldEntry.BytesReceived)) / float32(seconds)
		} else {
			tx = float32(entry.BytesSent) / float32(seconds)
			rx = float32(entry.BytesReceived) / float32(seconds)
		}

		if index, exists := statIndexMap[entry.CommonName]; exists {
			fw.statistic[index].tx = tx
			fw.statistic[index].rx = rx
		} else {
			fw.statistic = append(fw.statistic, SpeedStat{
				commonName: entry.CommonName,
				tx:         tx,
				rx:         rx,
			})
		}
	}

	i := 0
	for _, stat := range fw.statistic {
		if activeUsers[stat.commonName] {
			fw.statistic[i] = stat
			i++
		}
	}
	fw.statistic = fw.statistic[:i]

	fw.data = data

	for commonName, clEntry := range clMap {
		if !activeUsers[commonName] {
			onDisconnect(clEntry)
		}
	}

}

// ConnectionList returns information about user's connections to the VPN server.
func ConnectionList() (list []clEntry, lastUpdate time.Time) {

	svr := TheServer()

	// Open the status log file.
	f, err := svr.openFunc(_DefaultStatusLogPath)
	if err != nil {
		panic(err)
	}

	cl, _, lU := parseStatusLogWUpdate(f)

	return cl, lU
}

func (fw *FileWatcher) GetStatistics() []SpeedStat {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	stats := make([]SpeedStat, len(fw.statistic))
	copy(stats, fw.statistic)
	return stats
}

func onDisconnect(clE clEntry) {
	var user dbUserModel
	result := db.Where("username = ?", clE.CommonName).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			fmt.Println("User " + clE.CommonName + " not found")
		} else {
			fmt.Println("Error:", result.Error)
		}
	} else {
		statistic := dbStatisticModel{
			UserID:         user.ID,
			ConnectedSince: clE.ConnectedSince,
			ConnectedUntil: time.Now(),
			BytesReceived:  clE.BytesReceived,
			BytesSent:      clE.BytesSent,
			CommonName:     clE.CommonName,
			RealAddress:    clE.RealAddress,
		}
		db.Save(&statistic)
	}
}

func GetStatisticList() ([]StatisticSummary, error) {
	return GetStatisticsByDateRange(nil, nil, "")
}

// GetStatisticsByDateRange возвращает статистику за указанный период с группировкой по CommonName
func GetStatisticsByDateRange(startDate, endDate *time.Time, commonNameFilter string) ([]StatisticSummary, error) {
	var results []StatisticSummary

	query := db.Table("db_statistic_models").
		Select(`
            common_name,
            COUNT(*) as connection_count,
            SUM(bytes_received) as total_bytes_received,
            SUM(bytes_sent) as total_bytes_sent,
            SUM(bytes_received + bytes_sent) as total_bytes,
			AVG(CAST((strftime('%s', connected_until) - strftime('%s', connected_since)) AS REAL)) 
				as avg_connection_duration_seconds`)

	if startDate != nil && !startDate.IsZero() {
		query = query.Where("connected_since > ?", startDate)
	}

	if endDate != nil && !endDate.IsZero() {
		query = query.Where("connected_since < ?", endDate)
	}

	if commonNameFilter != "" {
		query = query.Where("common_name LIKE ?", "%"+commonNameFilter+"%")
	}

	query = query.Group("common_name")

	result := query.Find(&results)
	if result.Error != nil {
		return nil, result.Error
	}

	return results, nil
}

// StatisticSummary структура для результатов группировки
type StatisticSummary struct {
	CommonName                string  `json:"common_name"`
	ConnectionCount           int64   `json:"connection_count"`
	TotalBytesReceived        int64   `json:"total_bytes_received"`
	TotalBytesSent            int64   `json:"total_bytes_sent"`
	TotalBytes                int64   `json:"total_bytes"`
	AvgConnectionDurationSecs float64 `json:"avg_connection_duration_seconds"`
}

// GetDetailedStatistics возвращает детальные записи с различными фильтрами
func GetDetailedStatistics(filters StatisticFilters) ([]dbStatisticModel, error) {
	var statistics []dbStatisticModel

	query := db.Model(&dbStatisticModel{})

	if !filters.StartDate.IsZero() {
		query = query.Where("connected_since >= ?", filters.StartDate)
	}

	if !filters.EndDate.IsZero() {
		query = query.Where("connected_until <= ?", filters.EndDate)
	}

	if filters.CommonName != "" {
		query = query.Where("common_name = ?", filters.CommonName)
	}

	if filters.RealAddress != "" {
		query = query.Where("real_address LIKE ?", "%"+filters.RealAddress+"%")
	}

	if filters.UserID != 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}

	// Сортировка по умолчанию - от новых к старым
	if filters.SortBy == "" {
		filters.SortBy = "connected_since"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "DESC"
	}

	query = query.Order(filters.SortBy + " " + filters.SortOrder)

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	result := query.Find(&statistics)
	if result.Error != nil {
		return nil, result.Error
	}

	return statistics, nil
}

// StatisticFilters структура для фильтров
type StatisticFilters struct {
	StartDate   time.Time `form:"start_date"`
	EndDate     time.Time `form:"end_date"`
	CommonName  string    `form:"common_name"`
	RealAddress string    `form:"real_address"`
	UserID      uint      `form:"user_id"`
	SortBy      string    `form:"sort_by"`
	SortOrder   string    `form:"sort_order"`
	Limit       int       `form:"limit"`
	Offset      int       `form:"offset"`
}

// GetUserStatistics возвращает статистику по конкретному пользователю
func GetUserStatistics(commonName string, startDate, endDate time.Time) (*UserStatistics, error) {
	var userStats UserStatistics

	// Получаем суммарную статистику
	result := db.Table("db_statistic_models").
		Select(`
            COUNT(*) as total_connections,
            SUM(bytes_received) as total_bytes_received,
            SUM(bytes_sent) as total_bytes_sent,
            SUM(bytes_received + bytes_sent) as total_bytes,
			AVG(CAST((strftime('%s', connected_until) - strftime('%s', connected_since)) AS REAL)) 
				as avg_connection_duration_seconds,
			MAX(connected_since) as last_connection
        `).
		Where("common_name = ? AND connected_since >= ? AND connected_until <= ?",
			commonName, startDate, endDate).
		Scan(&userStats)

	if result.Error != nil {
		return nil, result.Error
	}

	return &userStats, nil
}

// UserStatistics структура для статистики пользователя
type UserStatistics struct {
	Username                  string    `json:"username"`
	UserID                    uint      `json:"user_id"`
	TotalConnections          int64     `json:"total_connections"`
	TotalBytesReceived        int64     `json:"total_bytes_received"`
	TotalBytesSent            int64     `json:"total_bytes_sent"`
	TotalBytes                int64     `json:"total_bytes"`
	AvgConnectionDurationSecs float64   `json:"avg_connection_duration_seconds"`
	LastConnection            time.Time `json:"last_connection"`
}
