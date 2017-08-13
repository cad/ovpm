package ovpm

import (
	"testing"

	"github.com/bouk/monkey"
	"github.com/cad/ovpm/supervisor"
)

var fs map[string]string

func TestVPNInit(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	// Prepare:
	// Test:

	// Check database if the database has no server.
	var server DBServer
	db.First(&server)

	// Isn't server empty struct?
	if !db.NewRecord(&server) {
		t.Fatalf("server is expected to be empty struct(new record) but it isn't %+v", server)
	}

	// Initialize the server.
	Init("localhost", "")

	// Check database if the database has no server.
	var server2 DBServer
	db.First(&server2)

	// Is server empty struct?
	if db.NewRecord(&server2) {
		t.Fatalf("server is expected to be not empty struct(new record) but it is %+v", server2)
	}
}

func TestVPNDeinit(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()

	// Prepare:
	// Initialize the server.
	Init("localhost", "")
	u, _ := CreateNewUser("user", "p")
	u.Delete()

	// Test:
	var server DBServer
	db.First(&server)

	// Isn't server empty struct?
	if db.NewRecord(&server) {
		t.Fatalf("server is expected to be not empty struct(new record) but it is %+v", server)
	}

	// Test if Revoked table contains the removed user's entries.
	var revoked DBRevoked
	db.First(&revoked)

	if db.NewRecord(&revoked) {
		t.Errorf("revoked shouldn't be empty")
	}

	// Deinitialize.
	Deinit()

	// Get server from db.
	var server2 DBServer
	db.First(&server2)

	// Isn't server empty struct?
	if !db.NewRecord(&server2) {
		t.Fatalf("server is expected to be empty struct(new record) but it is not %+v", server2)
	}

	// Test if Revoked table contains the removed user's entries.
	var revoked2 DBRevoked
	db.First(&revoked2)

	// Is revoked empty?
	if !db.NewRecord(&revoked2) {
		t.Errorf("revoked should be empty")
	}
}

func TestVPNIsInitialized(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()

	// Prepare:

	// Test:
	// Is initialized?
	if IsInitialized() {
		t.Fatalf("IsInitialized() is expected to return false but it returned true")
	}

	// Initialize the server.
	Init("localhost", "")

	// Isn't initialized?
	if !IsInitialized() {
		t.Fatalf("IsInitialized() is expected to return true but it returned false")
	}
}

func TestVPNGetServerInstance(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()

	// Prepare:

	// Test:
	server, err := GetServerInstance()

	// Is it nil?
	if err == nil {
		t.Fatalf("GetServerInstance() is expected to give error since server is not initialized yet, but it gave no error instead")
	}

	// Isn't server nil?
	if server != nil {
		t.Fatal("server is expected to be nil but it's not")
	}

	// Initialize server.
	Init("localhost", "")

	server, err = GetServerInstance()

	// Isn't it nil?
	if err != nil {
		t.Fatalf("GetServerInstance() is expected to give no error since server is initialized yet, but it gave error instead")
	}

	// Is server nil?
	if server == nil {
		t.Fatal("server is expected to be not nil but it is")
	}
}

func TestVPNDumpsClientConfig(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	user, _ := CreateNewUser("user", "password")

	// Test:
	clientConfigBlob, err := DumpsClientConfig(user.GetUsername())
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
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	user, _ := CreateNewUser("user", "password")

	// Test:
	err := DumpClientConfig(user.GetUsername(), "/tmp/user.ovpn")
	if err != nil {
		t.Fatalf("expected to dump client config but we got error instead: %v", err)
	}

	// Read file.
	clientConfigBlob := fs["/tmp/user.ovpn"]

	// Is empty?
	if len(clientConfigBlob) == 0 {
		t.Fatal("expected the dump not empty but it's empty instead")
	}
}

func TestVPNGetSystemCA(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()

	// Prepare:

	// Test:
	ca, err := GetSystemCA()
	if err == nil {
		t.Fatalf("GetSystemCA() is expected to give error but it didn't instead")
	}

	// Initialize system.
	Init("localhost", "")

	ca, err = GetSystemCA()
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
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()

	// Prepare:

	// Test:
	// Isn't it stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}

	// Call start without server initialization.
	StartVPNProc()

	// Isn't it still stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}

	// Initialize OVPM server.
	Init("localhost", "")

	// Call start again..
	StartVPNProc()

	// Isn't it RUNNING?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}
}

func TestVPNStopVPNProc(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:
	vpnProc.Start()

	// Test:
	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}

	// Call stop.
	StopVPNProc()

	// Isn't it stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}
}

func TestVPNRestartVPNProc(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:

	// Test:

	// Call restart.
	// Isn't it stopped?
	if vpnProc.Status() != supervisor.STOPPED {
		t.Fatalf("expected state is STOPPED, got %s instead", vpnProc.Status())
	}

	RestartVPNProc()

	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}

	// Call restart again.
	RestartVPNProc()

	// Isn't it running?
	if vpnProc.Status() != supervisor.RUNNING {
		t.Fatalf("expected state is RUNNING, got %s instead", vpnProc.Status())
	}
}

func TestVPNEmit(t *testing.T) {
	// Init:
	Testing = true
	SetupDB("sqlite3", ":memory:")
	defer CeaseDB()
	Init("localhost", "")

	// Prepare:

	// Test:
	Emit()

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

func init() {
	// Initialize.
	fs = make(map[string]string)

	// Monkeypatch emitToFile()
	monkey.Patch(emitToFile, func(path, content string, mode uint) error {
		fs[path] = content
		return nil
	})

	vpnProc = &fakeProcess{state: supervisor.STOPPED}
}
