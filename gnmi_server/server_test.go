package gnmi

// server_test covers gNMI get, subscribe (stream and poll) test
// Prerequisite: redis-server should be running.
import (
	"crypto/tls"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"fmt"
	"time"
	"reflect"

	testcert "github.com/sonic-net/sonic-gnmi/testdata/tls"

	"testing"

	_ "github.com/openconfig/gnmi/client"
	_ "github.com/openconfig/ygot/ygot"
	_ "github.com/google/gnxi/utils"
	_ "github.com/jipanyang/gnxi/utils/xpath"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/sonic-net/sonic-gnmi/common_utils"
	// Register supported client types.
	sdc "github.com/sonic-net/sonic-gnmi/sonic_data_client"

	"github.com/agiledragon/gomonkey"
	"github.com/godbus/dbus/v5"
)

func createServer(t *testing.T, port int64) *Server {
	certificate, err := testcert.NewCert()
	if err != nil {
		t.Errorf("could not load server key pair: %s", err)
	}
	tlsCfg := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{certificate},
	}

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsCfg))}
	cfg := &Config{Port: port}
	s, err := NewServer(cfg, opts)
	if err != nil {
		t.Errorf("Failed to create gNMI server: %v", err)
	}
	return s
}

func createInvalidServer(t *testing.T, port int64) *Server {
	certificate, err := testcert.NewCert()
	if err != nil {
		t.Errorf("could not load server key pair: %s", err)
	}
	tlsCfg := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{certificate},
	}

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsCfg))}
	s, err := NewServer(nil, opts)
	if err != nil {
		return nil
	}
	return s
}

func createAuthServer(t *testing.T, port int64) *Server {
	certificate, err := testcert.NewCert()
	if err != nil {
		t.Errorf("could not load server key pair: %s", err)
	}
	tlsCfg := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{certificate},
	}

	opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(tlsCfg))}
	cfg := &Config{Port: port, UserAuth: AuthTypes{"password": true, "cert": true, "jwt": true}}
	JwtRefreshInt = time.Duration(900*uint64(time.Second))
	JwtValidInt = time.Duration(3600*uint64(time.Second))
	GenerateJwtSecretKey()
	s, err := NewServer(cfg, opts)
	if err != nil {
		t.Errorf("Failed to create gNMI server: %v", err)
	}
	return s
}

func runServer(t *testing.T, s *Server) {
	//t.Log("Starting RPC server on address:", s.Address())
	err := s.Serve() // blocks until close
	if err != nil {
		t.Fatalf("gRPC server err: %v", err)
	}
	//t.Log("Exiting RPC server on address", s.Address())
}

func TestAll(t *testing.T) {
	mock1 := gomonkey.ApplyFunc(dbus.SystemBus, func() (conn *dbus.Conn, err error) {
		return &dbus.Conn{}, nil
	})
	defer mock1.Reset()
	mock2 := gomonkey.ApplyMethod(reflect.TypeOf(&dbus.Object{}), "Go", func(obj *dbus.Object, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
		ret := &dbus.Call{}
		ret.Err = nil
		ret.Body = make([]interface{}, 2)
		ret.Body[0] = int32(0)
		ch <- ret
		return &dbus.Call{}
	})
	defer mock2.Reset()
	mockCode := 
`
print('No Yang validation for test mode...')
print('%s')
`
	mock3 := gomonkey.ApplyGlobalVar(&sdc.PyCodeForYang, mockCode)
	defer mock3.Reset()

	s := createServer(t, 8080)
	go runServer(t, s)
	defer s.s.Stop()

	path, _ := os.Getwd()
	path = filepath.Dir(path)

	var cmd *exec.Cmd
	cmd = exec.Command("bash", "-c", "cd "+path+" && "+"pytest -m noauth")
	if result, err := cmd.Output(); err != nil {
		fmt.Println(string(result))
		t.Errorf("Fail to execute pytest: %v", err)
	} else {
		fmt.Println(string(result))
	}

	var counters [len(common_utils.CountersName)]uint64
	err := common_utils.GetMemCounters(&counters)
	if err != nil {
		t.Errorf("Error: Fail to read counters, %v", err)
	}
	for i := 0; i < len(common_utils.CountersName); i++ {
		if common_utils.CountersName[i] == "GNMI set" && counters[i] == 0 {
			t.Errorf("GNMI set counter should not be 0")
		}
		if common_utils.CountersName[i] == "GNMI get" && counters[i] == 0 {
			t.Errorf("GNMI get counter should not be 0")
		}
	}
	return
}

func TestServerPort(t *testing.T) {
	s := createServer(t, -8080)
	port := s.Port()
	if port != 0 {
		t.Errorf("Invalid port: %d", port)
	}
	return
}

func TestInvalidServer(t *testing.T) {
	s := createInvalidServer(t, 8080)
	if s != nil {
		t.Errorf("Should not create invalid server")
	}
	return
}

func TestAuth(t *testing.T) {
	mock1 := gomonkey.ApplyFunc(UserPwAuth, func(username string, passwd string) (bool, error) {
		return true, nil
	})
	defer mock1.Reset()

	s := createAuthServer(t, 8080)
	go runServer(t, s)
	defer s.s.Stop()

	path, _ := os.Getwd()
	path = filepath.Dir(path)

	var cmd *exec.Cmd
	cmd = exec.Command("bash", "-c", "cd "+path+" && "+"pytest -m auth")
	if result, err := cmd.Output(); err != nil {
		fmt.Println(string(result))
		t.Errorf("Fail to execute pytest: %v", err)
	} else {
		fmt.Println(string(result))
	}
	return
}

func TestAuthType(t *testing.T) {
	var ret bool
	var err error
	at := AuthTypes{"password": true, "cert": true, "jwt": true}
	ret = at.Enabled("password")
	if ret != true {
		t.Errorf("Enable ret is wrong: %v", ret)
	}
	ret = at.Enabled("invalid")
	if ret != false {
		t.Errorf("Enable ret is wrong: %v", ret)
	}
	err = at.Set("password")
	if err != nil {
		t.Errorf("Set ret is wrong: %v", err)
	}
	err = at.Set("invalid")
	if err == nil {
		t.Errorf("Set ret is wrong: %v", err)
	}
	err = at.Unset("password")
	if err != nil {
		t.Errorf("Unset ret is wrong: %v", err)
	}
	err = at.Unset("invalid")
	if err == nil {
		t.Errorf("Unset ret is wrong: %v", err)
	}
	fmt.Println(at.String())
}

func TestAuthUser(t *testing.T) {
	var err error
	auth := common_utils.AuthInfo{}
	err = PopulateAuthStruct("root", &auth, nil)
	if err != nil {
		t.Errorf("PopulateAuthStruct failed: %v", err)
	}
	err = PopulateAuthStruct("invalid_user_name", &auth, nil)
	if err == nil {
		t.Errorf("PopulateAuthStruct failed: %v", err)
	}
}

func init() {
	// Enable logs at UT setup
	flag.Lookup("v").Value.Set("10")
	flag.Lookup("log_dir").Value.Set("/tmp/gnmitest")

	// Inform gNMI server to use redis tcp localhost connection
	sdc.UseRedisLocalTcpPort = true
}
