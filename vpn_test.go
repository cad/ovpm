package ovpm

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cad/ovpm/supervisor"
)

var fs map[string]string

func setupTestCase() {
	// Initialize.
	fs = make(map[string]string)
	vpnProc.Stop()
}

func TestVPNInit(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	// Prepare:
	// Test:

	// Check database if the database has no server.
	var server dbServerModel
	db.First(&server)

	// Isn't server empty struct?
	if !db.NewRecord(&server) {
		t.Fatalf("server is expected to be empty struct(new record) but it isn't %+v", server)
	}

	// Wrongfully initialize server.

	if err := TheServer().Init("localhost", "asdf", UDPProto, "", ""); err == nil {
		t.Fatalf("error is expected to be not nil but it's nil instead")
	}

	// Initialize the server.
	TheServer().Init("localhost", "", UDPProto, "", "")

	// Check database if the database has no server.
	var server2 dbServerModel
	db.First(&server2)

	// Is server empty struct?
	if db.NewRecord(&server2) {
		t.Fatalf("server is expected to be not empty struct(new record) but it is %+v", server2)
	}
}

func TestVPNDeinit(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()

	// Prepare:
	// Initialize the server.
	TheServer().Init("localhost", "", UDPProto, "", "")
	u, err := CreateNewUser("user", "p", false, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	u.Delete()

	// Test:
	var server dbServerModel
	db.First(&server)

	// Isn't server empty struct?
	if db.NewRecord(&server) {
		t.Fatalf("server is expected to be not empty struct(new record) but it is %+v", server)
	}

	// Test if Revoked table contains the removed user's entries.
	var revoked dbRevokedModel
	db.First(&revoked)

	if db.NewRecord(&revoked) {
		t.Errorf("revoked shouldn't be empty")
	}

	// Deinitialize.
	TheServer().Deinit()

	// Get server from db.
	var server2 dbServerModel
	db.First(&server2)

	// Isn't server empty struct?
	if !db.NewRecord(&server2) {
		t.Fatalf("server is expected to be empty struct(new record) but it is not %+v", server2)
	}

	// Test if Revoked table contains the removed user's entries.
	var revoked2 dbRevokedModel
	db.First(&revoked2)

	// Is revoked empty?
	if !db.NewRecord(&revoked2) {
		t.Errorf("revoked should be empty")
	}
}
func TestVPNUpdate(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	// Prepare:
	TheServer().Init("localhost", "", UDPProto, "", "")
	// Test:

	var updatetests = []struct {
		vpnnet     string
		dns        string
		vpnChanged bool
		dnsChanged bool
	}{
		{"", "", false, false},
		{"192.168.9.0/24", "", true, false},
		{"", "2.2.2.2", false, true},
		{"9.9.9.0/24", "1.1.1.1", true, true},
	}
	for _, tt := range updatetests {
		svr := TheServer()

		oldIP := svr.Net
		oldDNS := svr.DNS
		svr.Update(tt.vpnnet, tt.dns)
		svr = nil
		svr = TheServer()
		if (svr.Net != oldIP) != tt.vpnChanged {
			t.Fatalf("expected vpn change: %t but opposite happened", tt.vpnChanged)
		}
		if (svr.DNS != oldDNS) != tt.dnsChanged {
			t.Fatalf("expected vpn change: %t but opposite happened", tt.dnsChanged)
		}
	}

}

func TestVPNIsInitialized(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()

	// Prepare:

	// Test:
	// Is initialized?
	if TheServer().IsInitialized() {
		t.Fatalf("IsInitialized() is expected to return false but it returned true")
	}

	// Initialize the server.
	TheServer().Init("localhost", "", UDPProto, "", "")

	// Isn't initialized?
	if !TheServer().IsInitialized() {
		t.Fatalf("IsInitialized() is expected to return true but it returned false")
	}
}

func TestVPNTheServer(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()

	// Prepare:

	// Test:
	svr := TheServer()

	// Isn't server nil?
	if svr.IsInitialized() {
		t.Fatal("server is expected to be not initialized it is")
	}

	// Initialize server.
	svr.Init("localhost", "", UDPProto, "", "")

	svr = TheServer()

	// Is server nil?
	if !svr.IsInitialized() {
		t.Fatal("server is expected to be initialized but it's not")
	}
}

func TestVPNDumpsClientConfig(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")

	// Prepare:
	user, _ := CreateNewUser("user", "password", false, 0, true)

	// Test:
	clientConfigBlob, err := svr.DumpsClientConfig(user.GetUsername())
	if err != nil {
		t.Fatalf("expected to dump client config but we got error instead: %v", err)
	}

	// Is empty?
	if len(clientConfigBlob) == 0 {
		t.Fatal("expected the dump not empty but it's empty instead")
	}
}

func TestVPNDumpClientConfig(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")

	// Prepare:
	noGW := false
	user, err := CreateNewUser("user", "password", noGW, 0, true)
	if err != nil {
		t.Fatalf("can not create user: %v", err)
	}

	// Test:
	if err = svr.DumpClientConfig(user.GetUsername(), "/tmp/user.ovpn"); err != nil {
		t.Fatalf("expected to dump client config but we got error instead: %v", err)
	}

	// Read file.
	clientConfigBlob := fs["/tmp/user.ovpn"]

	// Is empty?
	if len(clientConfigBlob) == 0 {
		t.Fatal("expected the dump not empty but it's empty instead")
	}

	// Is noGW honored?
	if strings.Contains(clientConfigBlob, "route-nopull") != noGW {
		logrus.Info(clientConfigBlob)
		t.Fatalf("client config generator doesn't honor NoGW")
	}

	user.Delete()

	noGW = true
	user, err = CreateNewUser("user", "password", noGW, 0, true)
	if err != nil {
		t.Fatalf("can not create user: %v", err)
	}

	if err = TheServer().DumpClientConfig(user.GetUsername(), "/tmp/user.ovpn"); err != nil {
		t.Fatalf("expected to dump client config but we got error instead: %v", err)
	}

	// Read file.
	clientConfigBlob = fs["/tmp/user.ovpn"]

}

func TestVPNGetSystemCA(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()

	// Prepare:
	svr := TheServer()

	// Test:
	ca, err := svr.GetSystemCA()
	if err == nil {
		t.Fatalf("GetSystemCA() is expected to give error but it didn't instead")
	}

	// Initialize system.
	svr.Init("localhost", "", UDPProto, "", "")

	ca, err = svr.GetSystemCA()
	if err != nil {
		t.Fatalf("GetSystemCA() is expected to get system ca, but it gave us an error instead: %v", err)
	}

	// Is it empty?
	if len(ca.Cert) == 0 {
		t.Fatalf("ca.Cert is expected to be not empty, but it's empty instead")
	}
	if len(ca.Key) == 0 {
		t.Fatalf("ca.Key is expected to be not empty, but it's empty instead")

	}
}

func TestVPNStartVPNProc(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()

	// Prepare:

	// Test:
	// Isn't it stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}

	// Call start without server initialization.
	svr.StartVPNProc()

	// Isn't it still stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}

	// Initialize OVPM server.
	svr.Init("localhost", "", UDPProto, "", "")

	// Call start again..
	svr.StartVPNProc()

	// Isn't it RUNNING?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}
}

func TestVPNStopVPNProc(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")

	// Prepare:
	vpnProc.Start()

	// Test:
	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}

	// Call stop.
	svr.StopVPNProc()

	// Isn't it stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}
}

func TestVPNRestartVPNProc(t *testing.T) {
	// Init:
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")

	// Prepare:

	// Test:

	// Call restart.
	svr.RestartVPNProc()

	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}

	// Call restart again.
	svr.RestartVPNProc()

	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}
}

func TestVPNEmit(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")

	// Prepare:

	// Test:
	svr.Emit()

	var emittests = []string{
		_DefaultVPNConfPath,
		_DefaultKeyPath,
		_DefaultCertPath,
		_DefaultCRLPath,
		_DefaultCACertPath,
		_DefaultCAKeyPath,
		_DefaultDHParamsPath,
	}

	for _, tt := range emittests {
		if len(fs[tt]) == 0 {
			t.Errorf("%s is expected to be not empty but it is", tt)
		}
	}

	// TODO(cad): Write test cases for ccd/ files as well.
}

func TestVPNemitToFile(t *testing.T) {
	// Initialize:

	// Prepare:
	path := "/test/file"
	content := "blah blah blah"

	// Test:
	// Is path exist?
	if _, ok := fs[path]; ok {
		t.Fatalf("key '%s' expected to be non-existent on fs, but it is instead", path)
	}

	// Emit the contents.

	if err := TheServer().emitToFile(path, content, 0); err != nil {
		t.Fatalf("expected  to be able to emit to the filesystem but we got this error instead: %v", err)
	}

	// Is the content on the filesystem correct?
	if fs[path] != content {
		t.Fatalf("content on the filesytem is expected to be same with '%s' but it's '%s' instead", content, fs[path])
	}
}

type fakeProcess struct {
	state supervisor.State
}

func (f *fakeProcess) Start() {
	f.state = supervisor.RUNNING
}

func (f *fakeProcess) Stop() {
	f.state = supervisor.STOPPED
}

func (f *fakeProcess) Restart() {
	f.state = supervisor.RUNNING
}

func (f *fakeProcess) Status() supervisor.State {
	return f.state
}

func TestGetConnectedUsers(t *testing.T) {
	// Init:
	setupTestCase()
	CreateDB("sqlite3", ":memory:")
	defer db.Cease()
	svr := TheServer()
	svr.Init("localhost", "", UDPProto, "", "")
	// Change openr from os.Open to our mock file injecter.
	svr.openFunc = func(path string) (io.Reader, error) {
		const testLog = `
OpenVPN CLIENT LIST
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
		return bytes.NewBufferString(testLog), nil
	}

	// Create the corresponding users for test.
	if _, err := CreateNewUser("usr1", "1234", true, 0, false); err != nil {
		t.Fatalf("user creation failed: %v", err)
	}
	if _, err := CreateNewUser("usr2", "1234", true, 0, false); err != nil {
		t.Fatalf("user creation failed: %v", err)
	}

	// Test:
	tests := []struct {
		name    string
		want    []User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TheServer().GetConnectedUsers()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConnectedUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConnectedUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func init() {
	// Init
	Testing = true
	fs = make(map[string]string)

	CreateDB("sqlite3", ":memory:")
	defer db.Cease()

	// Monkeypatch emitToFile()
	fmt.Println(TheServer())
	TheServer().emitToFileFunc = func(path, content string, mode uint) error {
		fs[path] = content
		return nil
	}
	vpnProc = &fakeProcess{state: supervisor.STOPPED}
}
