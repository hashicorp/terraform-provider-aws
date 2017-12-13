package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin/test/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Test that NetRPCUnsupportedPlugin implements the correct interfaces.
var _ Plugin = new(NetRPCUnsupportedPlugin)

// testAPIVersion is the ProtocolVersion we use for testing.
var testHandshake = HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_MAGIC_COOKIE",
	MagicCookieValue: "test",
}

// testInterface is the test interface we use for plugins.
type testInterface interface {
	Double(int) int
	PrintKV(string, interface{})
}

// testInterfacePlugin is the implementation of Plugin to create
// RPC client/server implementations for testInterface.
type testInterfacePlugin struct {
	Impl testInterface
}

func (p *testInterfacePlugin) Server(b *MuxBroker) (interface{}, error) {
	return &testInterfaceServer{Impl: p.impl()}, nil
}

func (p *testInterfacePlugin) Client(b *MuxBroker, c *rpc.Client) (interface{}, error) {
	return &testInterfaceClient{Client: c}, nil
}

func (p *testInterfacePlugin) GRPCServer(s *grpc.Server) error {
	grpctest.RegisterTestServer(s, &testGRPCServer{Impl: p.impl()})
	return nil
}

func (p *testInterfacePlugin) GRPCClient(c *grpc.ClientConn) (interface{}, error) {
	return &testGRPCClient{Client: grpctest.NewTestClient(c)}, nil
}

func (p *testInterfacePlugin) impl() testInterface {
	if p.Impl != nil {
		return p.Impl
	}

	return &testInterfaceImpl{
		logger: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Trace,
			Output:     os.Stderr,
			JSONFormat: true,
		}),
	}
}

// testInterfaceImpl implements testInterface concretely
type testInterfaceImpl struct {
	logger hclog.Logger
}

func (i *testInterfaceImpl) Double(v int) int { return v * 2 }

func (i *testInterfaceImpl) PrintKV(key string, value interface{}) {
	i.logger.Info("PrintKV called", key, value)
}

// testInterfaceClient implements testInterface to communicate over RPC
type testInterfaceClient struct {
	Client *rpc.Client
}

func (impl *testInterfaceClient) Double(v int) int {
	var resp int
	err := impl.Client.Call("Plugin.Double", v, &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

func (impl *testInterfaceClient) PrintKV(key string, value interface{}) {
	err := impl.Client.Call("Plugin.PrintKV", map[string]interface{}{
		"key":   key,
		"value": value,
	}, &struct{}{})
	if err != nil {
		panic(err)
	}
}

// testInterfaceServer is the RPC server for testInterfaceClient
type testInterfaceServer struct {
	Broker *MuxBroker
	Impl   testInterface
}

func (s *testInterfaceServer) Double(arg int, resp *int) error {
	*resp = s.Impl.Double(arg)
	return nil
}

func (s *testInterfaceServer) PrintKV(args map[string]interface{}, _ *struct{}) error {
	s.Impl.PrintKV(args["key"].(string), args["value"])
	return nil
}

// testPluginMap can be used for tests as a plugin map
var testPluginMap = map[string]Plugin{
	"test": new(testInterfacePlugin),
}

// testGRPCServer is the implementation of our GRPC service.
type testGRPCServer struct {
	Impl testInterface
}

func (s *testGRPCServer) Double(
	ctx context.Context,
	req *grpctest.TestRequest) (*grpctest.TestResponse, error) {
	return &grpctest.TestResponse{
		Output: int32(s.Impl.Double(int(req.Input))),
	}, nil
}

func (s *testGRPCServer) PrintKV(
	ctx context.Context,
	req *grpctest.PrintKVRequest) (*grpctest.PrintKVResponse, error) {
	var v interface{}
	switch rv := req.Value.(type) {
	case *grpctest.PrintKVRequest_ValueString:
		v = rv.ValueString

	case *grpctest.PrintKVRequest_ValueInt:
		v = rv.ValueInt

	default:
		panic(fmt.Sprintf("unknown value: %#v", req.Value))
	}

	s.Impl.PrintKV(req.Key, v)
	return &grpctest.PrintKVResponse{}, nil
}

// testGRPCClient is an implementation of TestInterface that communicates
// over gRPC.
type testGRPCClient struct {
	Client grpctest.TestClient
}

func (c *testGRPCClient) Double(v int) int {
	resp, err := c.Client.Double(context.Background(), &grpctest.TestRequest{
		Input: int32(v),
	})
	if err != nil {
		panic(err)
	}

	return int(resp.Output)
}

func (c *testGRPCClient) PrintKV(key string, value interface{}) {
	req := &grpctest.PrintKVRequest{Key: key}
	switch v := value.(type) {
	case string:
		req.Value = &grpctest.PrintKVRequest_ValueString{
			ValueString: v,
		}

	case int:
		req.Value = &grpctest.PrintKVRequest_ValueInt{
			ValueInt: int32(v),
		}

	default:
		panic(fmt.Sprintf("unknown type: %T", value))
	}

	_, err := c.Client.PrintKV(context.Background(), req)
	if err != nil {
		panic(err)
	}
}

func helperProcess(s ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"PLUGIN_MIN_PORT=10000",
		"PLUGIN_MAX_PORT=25000",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	return cmd
}

// This is not a real test. This is just a helper process kicked off by
// tests.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}

		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	// override testPluginMap with one that uses
	// hclog logger on its implementation
	pluginLogger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	testPlugin := &testInterfaceImpl{
		logger: pluginLogger,
	}

	testPluginMap := map[string]Plugin{
		"test": &testInterfacePlugin{Impl: testPlugin},
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "bad-version":
		// If we have an arg, we write there on start
		if len(args) > 0 {
			path := args[0]
			err := ioutil.WriteFile(path, []byte("foo"), 0644)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("%d|%d1|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		<-make(chan int)
	case "invalid-rpc-address":
		fmt.Println("lolinvalid")
	case "mock":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		<-make(chan int)
	case "start-timeout":
		time.Sleep(1 * time.Minute)
		os.Exit(1)
	case "stderr":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		os.Stderr.WriteString("HELLO\n")
		os.Stderr.WriteString("WORLD\n")
	case "stderr-json":
		// write values that might be JSON, but aren't KVs
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		os.Stderr.WriteString("[\"HELLO\"]\n")
		os.Stderr.WriteString("12345\n")
	case "stdin":
		fmt.Printf("%d|%d|tcp|:1234\n", CoreProtocolVersion, testHandshake.ProtocolVersion)
		data := make([]byte, 5)
		if _, err := os.Stdin.Read(data); err != nil {
			log.Printf("stdin read error: %s", err)
			os.Exit(100)
		}

		if string(data) == "hello" {
			os.Exit(0)
		}

		os.Exit(1)
	case "cleanup":
		// Create a defer to write the file. This tests that we get cleaned
		// up properly versus just calling os.Exit
		path := args[0]
		defer func() {
			err := ioutil.WriteFile(path, []byte("foo"), 0644)
			if err != nil {
				panic(err)
			}
		}()

		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})

		// Exit
		return
	case "test-grpc":
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
			GRPCServer:      DefaultGRPCServer,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-grpc-tls":
		// Serve!
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
			GRPCServer:      DefaultGRPCServer,
			TLSProvider:     helperTLSProvider,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface":
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface-logger-netrpc":
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})
		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface-logger-grpc":
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
			GRPCServer:      DefaultGRPCServer,
		})
		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface-daemon":
		// Serve!
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	case "test-interface-tls":
		// Serve!
		Serve(&ServeConfig{
			HandshakeConfig: testHandshake,
			Plugins:         testPluginMap,
			TLSProvider:     helperTLSProvider,
		})

		// Shouldn't reach here but make sure we exit anyways
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		os.Exit(2)
	}
}

func helperTLSProvider() (*tls.Config, error) {
	serverCert, err := tls.X509KeyPair([]byte(TestClusterServerCert), []byte(TestClusterServerKey))
	if err != nil {
		return nil, err
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM([]byte(TestClusterCACert))
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      rootCAs,
		ClientCAs:    rootCAs,
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ServerName:   "127.0.0.1",
	}
	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}

const (
	TestClusterCACert = `-----BEGIN CERTIFICATE-----
MIIDPjCCAiagAwIBAgIUfIKsF2VPT7sdFcKOHJH2Ii6K4MwwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLbXl2YXVsdC5jb20wIBcNMTYwNTAyMTYwNTQyWhgPMjA2
NjA0MjAxNjA2MTJaMBYxFDASBgNVBAMTC215dmF1bHQuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuOimEXawD2qBoLCFP3Skq5zi1XzzcMAJlfdS
xz9hfymuJb+cN8rB91HOdU9wQCwVKnkUtGWxUnMp0tT0uAZj5NzhNfyinf0JGAbP
67HDzVZhGBHlHTjPX0638yaiUx90cTnucX0N20SgCYct29dMSgcPl+W78D3Jw3xE
JsHQPYS9ASe2eONxG09F/qNw7w/RO5/6WYoV2EmdarMMxq52pPe2chtNMQdSyOUb
cCcIZyk4QVFZ1ZLl6jTnUPb+JoCx1uMxXvMek4NF/5IL0Wr9dw2gKXKVKoHDr6SY
WrCONRw61A5Zwx1V+kn73YX3USRlkufQv/ih6/xThYDAXDC9cwIDAQABo4GBMH8w
DgYDVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOuKvPiU
G06iHkRXAOeMiUdBfHFyMB8GA1UdIwQYMBaAFOuKvPiUG06iHkRXAOeMiUdBfHFy
MBwGA1UdEQQVMBOCC215dmF1bHQuY29thwR/AAABMA0GCSqGSIb3DQEBCwUAA4IB
AQBcN/UdAMzc7UjRdnIpZvO+5keBGhL/vjltnGM1dMWYHa60Y5oh7UIXF+P1RdNW
n7g80lOyvkSR15/r1rDkqOK8/4oruXU31EcwGhDOC4hU6yMUy4ltV/nBoodHBXNh
MfKiXeOstH1vdI6G0P6W93Bcww6RyV1KH6sT2dbETCw+iq2VN9CrruGIWzd67UT/
spe/kYttr3UYVV3O9kqgffVVgVXg/JoRZ3J7Hy2UEXfh9UtWNanDlRuXaZgE9s/d
CpA30CHpNXvKeyNeW2ktv+2nAbSpvNW+e6MecBCTBIoDSkgU8ShbrzmDKVwNN66Q
5gn6KxUPBKHEtNzs5DgGM7nq
-----END CERTIFICATE-----`

	TestClusterServerCert = `-----BEGIN CERTIFICATE-----
MIIDtzCCAp+gAwIBAgIUBLqh6ctGWVDUxFhxJX7m6S/bnrcwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLbXl2YXVsdC5jb20wIBcNMTYwNTAyMTYwOTI2WhgPMjA2
NjA0MjAxNTA5NTZaMBsxGTAXBgNVBAMTEGNlcnQubXl2YXVsdC5jb20wggEiMA0G
CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDY3gPB29kkdbu0mPO6J0efagQhSiXB
9OyDuLf5sMk6CVDWVWal5hISkyBmw/lXgF7qC2XFKivpJOrcGQd5Ep9otBqyJLzI
b0IWdXuPIrVnXDwcdWr86ybX2iC42zKWfbXgjzGijeAVpl0UJLKBj+fk5q6NvkRL
5FUL6TRV7Krn9mrmnrV9J5IqV15pTd9W2aVJ6IqWvIPCACtZKulqWn4707uy2X2W
1Stq/5qnp1pDshiGk1VPyxCwQ6yw3iEcgecbYo3vQfhWcv7Q8LpSIM9ZYpXu6OmF
+czqRZS9gERl+wipmmrN1MdYVrTuQem21C/PNZ4jo4XUk1SFx6JrcA+lAgMBAAGj
gfUwgfIwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMB0GA1UdDgQWBBSe
Cl9WV3BjGCwmS/KrDSLRjfwyqjAfBgNVHSMEGDAWgBTrirz4lBtOoh5EVwDnjIlH
QXxxcjA7BggrBgEFBQcBAQQvMC0wKwYIKwYBBQUHMAKGH2h0dHA6Ly8xMjcuMC4w
LjE6ODIwMC92MS9wa2kvY2EwIQYDVR0RBBowGIIQY2VydC5teXZhdWx0LmNvbYcE
fwAAATAxBgNVHR8EKjAoMCagJKAihiBodHRwOi8vMTI3LjAuMC4xOjgyMDAvdjEv
cGtpL2NybDANBgkqhkiG9w0BAQsFAAOCAQEAWGholPN8buDYwKbUiDavbzjsxUIX
lU4MxEqOHw7CD3qIYIauPboLvB9EldBQwhgOOy607Yvdg3rtyYwyBFwPhHo/hK3Z
6mn4hc6TF2V+AUdHBvGzp2dbYLeo8noVoWbQ/lBulggwlIHNNF6+a3kALqsqk1Ch
f/hzsjFnDhAlNcYFgG8TgfE2lE/FckvejPqBffo7Q3I+wVAw0buqiz5QL81NOT+D
Y2S9LLKLRaCsWo9wRU1Az4Rhd7vK5SEMh16jJ82GyEODWPvuxOTI1MnzfnbWyLYe
TTp6YBjGMVf1I6NEcWNur7U17uIOiQjMZ9krNvoMJ1A/cxCoZ98QHgcIPg==
-----END CERTIFICATE-----`

	TestClusterServerKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2N4DwdvZJHW7tJjzuidHn2oEIUolwfTsg7i3+bDJOglQ1lVm
peYSEpMgZsP5V4Be6gtlxSor6STq3BkHeRKfaLQasiS8yG9CFnV7jyK1Z1w8HHVq
/Osm19oguNsyln214I8xoo3gFaZdFCSygY/n5Oaujb5ES+RVC+k0Veyq5/Zq5p61
fSeSKldeaU3fVtmlSeiKlryDwgArWSrpalp+O9O7stl9ltUrav+ap6daQ7IYhpNV
T8sQsEOssN4hHIHnG2KN70H4VnL+0PC6UiDPWWKV7ujphfnM6kWUvYBEZfsIqZpq
zdTHWFa07kHpttQvzzWeI6OF1JNUhceia3APpQIDAQABAoIBAQCH3vEzr+3nreug
RoPNCXcSJXXY9X+aeT0FeeGqClzIg7Wl03OwVOjVwl/2gqnhbIgK0oE8eiNwurR6
mSPZcxV0oAJpwiKU4T/imlCDaReGXn86xUX2l82KRxthNdQH/VLKEmzij0jpx4Vh
bWx5SBPdkbmjDKX1dmTiRYWIn/KjyNPvNvmtwdi8Qluhf4eJcNEUr2BtblnGOmfL
FdSu+brPJozpoQ1QdDnbAQRgqnh7Shl0tT85whQi0uquqIj1gEOGVjmBvDDnL3GV
WOENTKqsmIIoEzdZrql1pfmYTk7WNaD92bfpN128j8BF7RmAV4/DphH0pvK05y9m
tmRhyHGxAoGBAOV2BBocsm6xup575VqmFN+EnIOiTn+haOvfdnVsyQHnth63fOQx
PNtMpTPR1OMKGpJ13e2bV0IgcYRsRkScVkUtoa/17VIgqZXffnJJ0A/HT67uKBq3
8o7RrtyK5N20otw0lZHyqOPhyCdpSsurDhNON1kPVJVYY4N1RiIxfut/AoGBAPHz
HfsJ5ZkyELE9N/r4fce04lprxWH+mQGK0/PfjS9caXPhj/r5ZkVMvzWesF3mmnY8
goE5S35TuTvV1+6rKGizwlCFAQlyXJiFpOryNWpLwCmDDSzLcm+sToAlML3tMgWU
jM3dWHx3C93c3ft4rSWJaUYI9JbHsMzDW6Yh+GbbAoGBANIbKwxh5Hx5XwEJP2yu
kIROYCYkMy6otHLujgBdmPyWl+suZjxoXWoMl2SIqR8vPD+Jj6mmyNJy9J6lqf3f
DRuQ+fEuBZ1i7QWfvJ+XuN0JyovJ5Iz6jC58D1pAD+p2IX3y5FXcVQs8zVJRFjzB
p0TEJOf2oqORaKWRd6ONoMKvAoGALKu6aVMWdQZtVov6/fdLIcgf0pn7Q3CCR2qe
X3Ry2L+zKJYIw0mwvDLDSt8VqQCenB3n6nvtmFFU7ds5lvM67rnhsoQcAOaAehiS
rl4xxoJd5Ewx7odRhZTGmZpEOYzFo4odxRSM9c30/u18fqV1Mm0AZtHYds4/sk6P
aUj0V+kCgYBMpGrJk8RSez5g0XZ35HfpI4ENoWbiwB59FIpWsLl2LADEh29eC455
t9Muq7MprBVBHQo11TMLLFxDIjkuMho/gcKgpYXCt0LfiNm8EZehvLJUXH+3WqUx
we6ywrbFCs6LaxaOCtTiLsN+GbZCatITL0UJaeBmTAbiw0KQjUuZPQ==
-----END RSA PRIVATE KEY-----`
)
