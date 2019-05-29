package ovpm

import (
	"io"
	"reflect"
	"testing"
	"time"
)

func TestUser_ConnectionStatus(t *testing.T) {
	// Init:
	db := CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "", "", "", false)

	origOpenFunc := svr.openFunc
	defer func() { svr.openFunc = origOpenFunc }()
	svr.openFunc = func(path string) (io.Reader, error) {
		return nil, nil
	}
	usr1, err := CreateNewUser("usr1", "1234", true, 0, false)
	if err != nil {
		t.Fatalf("user creation failed: %v", err)
	}
	now := time.Now()
	svr.parseStatusLogFunc = func(f io.Reader) ([]clEntry, []rtEntry) {
		clt := []clEntry{
			clEntry{
				CommonName:     usr1.GetUsername(),
				RealAddress:    "1.1.1.1",
				ConnectedSince: now,
				BytesReceived:  1,
				BytesSent:      5,
			},
		}
		rtt := []rtEntry{
			rtEntry{
				CommonName:     usr1.GetUsername(),
				RealAddress:    "1.1.1.1",
				LastRef:        now,
				VirtualAddress: "10.10.10.1",
			},
		}
		return clt, rtt
	}

	// Test:
	type fields struct {
		dbUserModel    dbUserModel
		isConnected    bool
		connectedSince time.Time
		bytesReceived  uint64
		bytesSent      uint64
	}
	tests := []struct {
		name               string
		fields             fields
		wantIsConnected    bool
		wantConnectedSince time.Time
		wantBytesSent      uint64
		wantBytesReceived  uint64
	}{
		{"default", fields{dbUserModel: dbUserModel{Username: "usr1"}}, true, now, 5, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				dbUserModel:    tt.fields.dbUserModel,
				isConnected:    tt.fields.isConnected,
				connectedSince: tt.fields.connectedSince,
				bytesReceived:  tt.fields.bytesReceived,
				bytesSent:      tt.fields.bytesSent,
			}
			gotIsConnected, gotConnectedSince, gotBytesSent, gotBytesReceived := u.ConnectionStatus()
			if gotIsConnected != tt.wantIsConnected {
				t.Errorf("User.ConnectionStatus() gotIsConnected = %v, want %v", gotIsConnected, tt.wantIsConnected)
			}
			if !reflect.DeepEqual(gotConnectedSince, tt.wantConnectedSince) {
				t.Errorf("User.ConnectionStatus() gotConnectedSince = %v, want %v", gotConnectedSince, tt.wantConnectedSince)
			}
			if gotBytesSent != tt.wantBytesSent {
				t.Errorf("User.ConnectionStatus() gotBytesSent = %v, want %v", gotBytesSent, tt.wantBytesSent)
			}
			if gotBytesReceived != tt.wantBytesReceived {
				t.Errorf("User.ConnectionStatus() gotBytesReceived = %v, want %v", gotBytesReceived, tt.wantBytesReceived)
			}
		})
	}
}

func init() {
	Testing = true
}
