package plugin

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

func TestClient(t *testing.T) {
	process := helperProcess("mock")
	c := NewClient(&ClientConfig{Cmd: process, HandshakeConfig: testHandshake})
	defer c.Kill()

	// Test that it parses the proper address
	addr, err := c.Start()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	if addr.Network() != "tcp" {
		t.Fatalf("bad: %#v", addr)
	}

	if addr.String() != ":1234" {
		t.Fatalf("bad: %#v", addr)
	}

	// Test that it exits properly if killed
	c.Kill()

	if process.ProcessState == nil {
		t.Fatal("should have process state")
	}

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

// This tests a bug where Kill would start
func TestClient_killStart(t *testing.T) {
	// Create a temporary dir to store the result file
	td, err := ioutil.TempDir("", "plugin")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Start the client
	path := filepath.Join(td, "booted")
	process := helperProcess("bad-version", path)
	c := NewClient(&ClientConfig{Cmd: process, HandshakeConfig: testHandshake})
	defer c.Kill()

	// Verify our path doesn't exist
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		t.Fatalf("bad: %s", err)
	}

	// Test that it parses the proper address
	if _, err := c.Start(); err == nil {
		t.Fatal("expected error")
	}

	// Verify we started
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("bad: %s", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Test that Kill does nothing really
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}

	if process.ProcessState == nil {
		t.Fatal("should have no process state")
	}

	// Verify our path doesn't exist
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		t.Fatalf("bad: %s", err)
	}
}

func TestClient_testCleanup(t *testing.T) {
	// Create a temporary dir to store the result file
	td, err := ioutil.TempDir("", "plugin")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	// Create a path that the helper process will write on cleanup
	path := filepath.Join(td, "output")

	// Test the cleanup
	process := helperProcess("cleanup", path)
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})

	// Grab the client so the process starts
	if _, err := c.Client(); err != nil {
		c.Kill()
		t.Fatalf("err: %s", err)
	}

	// Kill it gracefully
	c.Kill()

	// Test for the file
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestClient_testInterface(t *testing.T) {
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_grpc(t *testing.T) {
	process := helperProcess("test-grpc")
	c := NewClient(&ClientConfig{
		Cmd:              process,
		HandshakeConfig:  testHandshake,
		Plugins:          testPluginMap,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})
	defer c.Kill()

	if _, err := c.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	if v := c.Protocol(); v != ProtocolGRPC {
		t.Fatalf("bad: %s", v)
	}

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_grpcNotAllowed(t *testing.T) {
	process := helperProcess("test-grpc")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	if _, err := c.Start(); err == nil {
		t.Fatal("should error")
	}
}

func TestClient_cmdAndReattach(t *testing.T) {
	config := &ClientConfig{
		Cmd:      helperProcess("start-timeout"),
		Reattach: &ReattachConfig{},
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_reattach(t *testing.T) {
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	// Grab the RPC client
	_, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Get the reattach configuration
	reattach := c.ReattachConfig()

	// Create a new client
	c = NewClient(&ClientConfig{
		Reattach:        reattach,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_reattachNoProtocol(t *testing.T) {
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	// Grab the RPC client
	_, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Get the reattach configuration
	reattach := c.ReattachConfig()
	reattach.Protocol = ""

	// Create a new client
	c = NewClient(&ClientConfig{
		Reattach:        reattach,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_reattachGRPC(t *testing.T) {
	process := helperProcess("test-grpc")
	c := NewClient(&ClientConfig{
		Cmd:              process,
		HandshakeConfig:  testHandshake,
		Plugins:          testPluginMap,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})
	defer c.Kill()

	// Grab the RPC client
	_, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Get the reattach configuration
	reattach := c.ReattachConfig()

	// Create a new client
	c = NewClient(&ClientConfig{
		Reattach:         reattach,
		HandshakeConfig:  testHandshake,
		Plugins:          testPluginMap,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_reattachNotFound(t *testing.T) {
	// Find a bad pid
	var pid int = 5000
	for i := pid; i < 32000; i++ {
		if _, err := os.FindProcess(i); err != nil {
			pid = i
			break
		}
	}

	// Addr that won't work
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	addr := l.Addr()
	l.Close()

	// Reattach
	c := NewClient(&ClientConfig{
		Reattach: &ReattachConfig{
			Addr: addr,
			Pid:  pid,
		},
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})

	// Start shouldn't error
	if _, err := c.Start(); err == nil {
		t.Fatal("should error")
	} else if err != ErrProcessNotFound {
		t.Fatalf("err: %s", err)
	}
}

func TestClientStart_badVersion(t *testing.T) {
	config := &ClientConfig{
		Cmd:             helperProcess("bad-version"),
		StartTimeout:    50 * time.Millisecond,
		HandshakeConfig: testHandshake,
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_Start_Timeout(t *testing.T) {
	config := &ClientConfig{
		Cmd:             helperProcess("start-timeout"),
		StartTimeout:    50 * time.Millisecond,
		HandshakeConfig: testHandshake,
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestClient_Stderr(t *testing.T) {
	stderr := new(bytes.Buffer)
	process := helperProcess("stderr")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		Stderr:          stderr,
		HandshakeConfig: testHandshake,
	})
	defer c.Kill()

	if _, err := c.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	for !c.Exited() {
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(stderr.String(), "HELLO\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}

	if !strings.Contains(stderr.String(), "WORLD\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}
}

func TestClient_StderrJSON(t *testing.T) {
	stderr := new(bytes.Buffer)
	process := helperProcess("stderr-json")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		Stderr:          stderr,
		HandshakeConfig: testHandshake,
	})
	defer c.Kill()

	if _, err := c.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	for !c.Exited() {
		time.Sleep(10 * time.Millisecond)
	}

	if !strings.Contains(stderr.String(), "[\"HELLO\"]\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}

	if !strings.Contains(stderr.String(), "12345\n") {
		t.Fatalf("bad log data: '%s'", stderr.String())
	}
}

func TestClient_Stdin(t *testing.T) {
	// Overwrite stdin for this test with a temporary file
	tf, err := ioutil.TempFile("", "terraform")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	if _, err = tf.WriteString("hello"); err != nil {
		t.Fatalf("error: %s", err)
	}

	if err = tf.Sync(); err != nil {
		t.Fatalf("error: %s", err)
	}

	if _, err = tf.Seek(0, 0); err != nil {
		t.Fatalf("error: %s", err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tf

	process := helperProcess("stdin")
	c := NewClient(&ClientConfig{Cmd: process, HandshakeConfig: testHandshake})
	defer c.Kill()

	_, err = c.Start()
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	for {
		if c.Exited() {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	if !process.ProcessState.Success() {
		t.Fatal("process didn't exit cleanly")
	}
}

func TestClient_SecureConfig(t *testing.T) {
	// Test failure case
	secureConfig := &SecureConfig{
		Checksum: []byte{'1'},
		Hash:     sha256.New(),
	}
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
		SecureConfig:    secureConfig,
	})

	// Grab the RPC client, should error
	_, err := c.Client()
	c.Kill()
	if err != ErrChecksumsDoNotMatch {
		t.Fatalf("err should be %s, got %s", ErrChecksumsDoNotMatch, err)
	}

	// Get the checksum of the executable
	file, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	hash := sha256.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		t.Fatal(err)
	}

	sum := hash.Sum(nil)

	secureConfig = &SecureConfig{
		Checksum: sum,
		Hash:     sha256.New(),
	}

	c = NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
		SecureConfig:    secureConfig,
	})
	defer c.Kill()

	// Grab the RPC client
	_, err = c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}
}

func TestClient_TLS(t *testing.T) {
	// Test failure case
	process := helperProcess("test-interface-tls")
	cBad := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer cBad.Kill()

	// Grab the RPC client
	clientBad, err := cBad.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := clientBad.Dispense("test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	cBad.Kill()

	// Add TLS config to client
	tlsConfig, err := helperTLSProvider()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	process = helperProcess("test-interface-tls")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
		TLSConfig:       tlsConfig,
	})
	defer c.Kill()

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err = client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_TLS_grpc(t *testing.T) {
	// Add TLS config to client
	tlsConfig, err := helperTLSProvider()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	process := helperProcess("test-grpc-tls")
	c := NewClient(&ClientConfig{
		Cmd:              process,
		HandshakeConfig:  testHandshake,
		Plugins:          testPluginMap,
		TLSConfig:        tlsConfig,
		AllowedProtocols: []Protocol{ProtocolGRPC},
	})
	defer c.Kill()

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}

func TestClient_secureConfigAndReattach(t *testing.T) {
	config := &ClientConfig{
		SecureConfig: &SecureConfig{},
		Reattach:     &ReattachConfig{},
	}

	c := NewClient(config)
	defer c.Kill()

	_, err := c.Start()
	if err != ErrSecureConfigAndReattach {
		t.Fatalf("err should not be %s, got %s", ErrSecureConfigAndReattach, err)
	}
}

func TestClient_ping(t *testing.T) {
	process := helperProcess("test-interface")
	c := NewClient(&ClientConfig{
		Cmd:             process,
		HandshakeConfig: testHandshake,
		Plugins:         testPluginMap,
	})
	defer c.Kill()

	// Get the client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Ping, should work
	if err := client.Ping(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Kill it
	c.Kill()
	if err := client.Ping(); err == nil {
		t.Fatal("should error")
	}
}

func TestClient_logger(t *testing.T) {
	t.Run("net/rpc", func(t *testing.T) { testClient_logger(t, "netrpc") })
	t.Run("grpc", func(t *testing.T) { testClient_logger(t, "grpc") })
}

func testClient_logger(t *testing.T, proto string) {
	var buffer bytes.Buffer
	stderr := io.MultiWriter(os.Stderr, &buffer)
	// Custom hclog.Logger
	clientLogger := hclog.New(&hclog.LoggerOptions{
		Name:   "test-logger",
		Level:  hclog.Trace,
		Output: stderr,
	})

	process := helperProcess("test-interface-logger-" + proto)
	c := NewClient(&ClientConfig{
		Cmd:              process,
		HandshakeConfig:  testHandshake,
		Plugins:          testPluginMap,
		Logger:           clientLogger,
		AllowedProtocols: []Protocol{ProtocolNetRPC, ProtocolGRPC},
	})
	defer c.Kill()

	// Grab the RPC client
	client, err := c.Client()
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	// Grab the impl
	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err should be nil, got %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	{
		// Discard everything else, and capture the output we care about
		buffer.Reset()
		impl.PrintKV("foo", "bar")
		time.Sleep(100 * time.Millisecond)
		line, err := buffer.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(line, "foo=bar") {
			t.Fatalf("bad: %q", line)
		}
	}

	{
		// Try an integer type
		buffer.Reset()
		impl.PrintKV("foo", 12)
		time.Sleep(100 * time.Millisecond)
		line, err := buffer.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(line, "foo=12") {
			t.Fatalf("bad: %q", line)
		}
	}

	// Kill it
	c.Kill()

	// Test that it knows it is exited
	if !c.Exited() {
		t.Fatal("should say client has exited")
	}
}
