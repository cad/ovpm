package ovpm

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// clEntry reprsents a parsed entry that is present on OpenVPN
// log section CLIENT LIST.
type clEntry struct {
	CommonName     string    `json:"common_name"`
	RealAddress    string    `json:"real_address"`
	BytesReceived  uint64    `json:"bytes_received"`
	BytesSent      uint64    `json:"bytes_sent"`
	ConnectedSince time.Time `json:"connected_since"`
}

// rtEntry reprsents a parsed entry that is present on OpenVPN
// log section ROUTING TABLE.
type rtEntry struct {
	VirtualAddress string    `json:"virtual_address"`
	CommonName     string    `json:"common_name"`
	RealAddress    string    `json:"real_address"`
	LastRef        time.Time `json:"last_ref"`
}

// parseStatusLog parses the received OpenVPN status log file.
// And then returns the parsed client information.

func parseStatusLog(f io.Reader) ([]clEntry, []rtEntry) {
	clE, rtE, _ := parseStatusLogWUpdate(f)
	return clE, rtE
}
func parseStatusLogWUpdate(f io.Reader) ([]clEntry, []rtEntry, time.Time) {
	// Recover any panics
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("OpenVPN log file is corrupt")
		}
	}()

	// Parsing stages.
	const stageUpd int = 0
	const stageCL int = 1
	const stageRT int = 2
	const stageFin int = 3

	// Parsing variables.
	var currStage int
	var skipFor int
	var cl []clEntry
	var rt []rtEntry
	var lastUpdate = time.Now()
	// Scan and parse the file by dividing it into chunks.
	scanner, skipFor := bufio.NewScanner(f), 1
	for lc := 0; scanner.Scan(); lc++ {
		if skipFor > 0 {
			skipFor--
			continue
		}
		txt := scanner.Text()
		switch currStage {
		case stageUpd:
			currStage = stageCL
			skipFor = 1
			if strings.Contains(txt, "Common Name") {
				continue
			}
			dat := strings.Split(txt, ",")
			lastUpdate = stodt(dat[1])
		case stageCL:
			if strings.Contains(txt, "ROUTING TABLE") {
				currStage = stageRT
				skipFor = 1
				continue
			}
			dat := strings.Split(txt, ",")
			cl = append(cl, clEntry{
				CommonName:     trim(dat[0]),
				RealAddress:    trim(dat[1]),
				BytesReceived:  stoui64(trim(dat[2])),
				BytesSent:      stoui64(trim(dat[3])),
				ConnectedSince: stodt(trim(dat[4])),
			})
		case stageRT:
			if strings.Contains(txt, "GLOBAL STATS") {
				currStage = stageFin
				break
			}
			dat := strings.Split(txt, ",")
			rt = append(rt, rtEntry{
				VirtualAddress: trim(dat[0]),
				CommonName:     trim(dat[1]),
				RealAddress:    trim(dat[2]),
				LastRef:        stodt(trim(dat[3])),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return cl, rt, lastUpdate
}

// stoi64 converts string to uint64.
func stoui64(s string) uint64 {
	i, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		panic(err)
	}
	return uint64(i)
}

// stodt converts string to date time.
func stodt(s string) time.Time {
	if s == "" {
		return time.Time{}
	}

	s = strings.TrimSpace(s)

	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"02.01.2006 15:04:05",
		"2006/01/02 15:04:05",
		time.ANSIC,
		"Jan _2 15:04:05 2006",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t
		}
	}

	if sec, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(sec, 0)
	}

	if millis, err := strconv.ParseInt(s, 10, 64); err == nil && millis > 1000000000000 {
		return time.Unix(millis/1000, (millis%1000)*1000000)
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	logrus.WithField("time_string", s).Warn("Failed to parse time string")
	return time.Time{}
}

// trim will trim all leading and trailing whitespace from the s.
func trim(s string) string {
	return strings.TrimSpace(s)
}
