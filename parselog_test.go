package ovpm

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func Test_parseStatusLog(t *testing.T) {
	const exampleLogFile = `OpenVPN CLIENT LIST
	Updated,Mon Mar 26 13:26:10 2018
	Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
	google.DNS,8.8.8.8:53246,527914279,3204562859,Sat Mar 17 16:26:38 2018
	google1.DNS,8.8.4.4:33974,42727443,291595456,Mon Mar 26 08:24:08 2018
	ROUTING TABLE
	Virtual Address,Common Name,Real Address,Last Ref
	10.20.30.6,google.DNS,8.8.8.8:33974,Mon Mar 26 13:26:04 2018
	10.20.30.5,google1.DNS,8.8.4.4:53246,Mon Mar 26 13:25:57 2018
	GLOBAL STATS
	Max bcast/mcast queue length,4
	END
	`

	// Mock the status log file.
	f := bytes.NewBufferString(exampleLogFile)

	type args struct {
		f io.Reader
	}
	tests := []struct {
		name  string
		args  args
		want  []clEntry
		want1 []rtEntry
	}{
		{
			"google", args{f},
			[]clEntry{
				clEntry{
					CommonName:     "google.DNS",
					RealAddress:    "8.8.8.8:53246",
					BytesReceived:  527914279,
					BytesSent:      3204562859,
					ConnectedSince: stodt("Sat Mar 17 16:26:38 2018"),
				},
				clEntry{
					CommonName:     "google1.DNS",
					RealAddress:    "8.8.4.4:33974",
					BytesReceived:  42727443,
					BytesSent:      291595456,
					ConnectedSince: stodt("Mon Mar 26 08:24:08 2018"),
				},
			},
			[]rtEntry{
				rtEntry{
					VirtualAddress: "10.20.30.6",
					CommonName:     "google.DNS",
					RealAddress:    "8.8.8.8:33974",
					LastRef:        stodt("Mon Mar 26 13:26:04 2018"),
				},
				rtEntry{
					VirtualAddress: "10.20.30.5",
					CommonName:     "google1.DNS",
					RealAddress:    "8.8.4.4:53246",
					LastRef:        stodt("Mon Mar 26 13:25:57 2018"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseStatusLog(tt.args.f)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStatusLog() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseStatusLog() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
