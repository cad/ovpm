package ovpm

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
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
func parseStatusLog(fPath string) ([]clEntry, []rtEntry) {
	// Parsing stages.
	const stageCL int = 0
	const stageRT int = 1
	const stageFin int = 2

	// Parsing variables.
	var currStage int
	var skipFor int
	var cl []clEntry
	var rt []rtEntry

	f, err := os.Open(fPath)
	if err != nil {
		panic(err)
	}

	// Scan and parse the file by dividing it into chunks.
	scanner, skipFor := bufio.NewScanner(f), 3
	for lc := 0; scanner.Scan(); lc++ {
		if skipFor > 0 {
			skipFor--
			continue
		}
		txt := scanner.Text()
		switch currStage {
		case stageCL:
			if txt == "ROUTING TABLE" {
				currStage = stageRT
				skipFor = 1
				continue
			}
			dat := strings.Split(txt, ",")
			cl = append(cl, clEntry{
				CommonName:     dat[0],
				RealAddress:    dat[1],
				BytesReceived:  stoui64(dat[2]),
				BytesSent:      stoui64(dat[3]),
				ConnectedSince: stodt(dat[4]),
			})
		case stageRT:
			if txt == "GLOBAL STATS" {
				currStage = stageFin
				break
			}
			dat := strings.Split(txt, ",")
			rt = append(rt, rtEntry{
				VirtualAddress: dat[0],
				CommonName:     dat[1],
				RealAddress:    dat[2],
				LastRef:        stodt(dat[3]),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return cl, rt
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
	t, err := time.ParseInLocation(time.ANSIC, s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}
