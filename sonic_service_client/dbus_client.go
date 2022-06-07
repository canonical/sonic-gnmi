package host_service

import (
	"time"
	"fmt"
	"reflect"
	log "github.com/golang/glog"
	"github.com/godbus/dbus/v5"
	"github.com/sonic-net/sonic-gnmi/common_utils"
)

type Service interface {
	ConfigReload(fileName string) error
	ConfigSave(fileName string) error
	ApplyPatchYang(fileName string) error
	ApplyPatchDb(fileName string) error
	CreateCheckPoint(cpName string)  error
	DeleteCheckPoint(cpName string) error
}

type DbusClient struct {
	busNamePrefix string
	busPathPrefix string
	intNamePrefix string
	testMode bool
	channel chan struct{}
}

func NewDbusClient(testMode bool) (Service, error) {
	var client DbusClient
	var err error

	client.busNamePrefix = "org.SONiC.HostService."
	client.busPathPrefix = "/org/SONiC/HostService/"
	client.intNamePrefix = "org.SONiC.HostService."
	client.testMode = testMode
	err = nil

	return &client, err
}

func DbusApi(busName string, busPath string, intName string, timeout int, args ...interface{}) error {
	common_utils.IncCounter("DBUS")
	conn, err := dbus.SystemBus()
	if err != nil {
		log.V(2).Infof("Failed to connect to system bus: %v", err)
		common_utils.IncCounter("DBUS fail")
		return err
	}

	ch := make(chan *dbus.Call, 1)
	obj := conn.Object(busName, dbus.ObjectPath(busPath))
	obj.Go(intName, 0, ch, args...)
	select {
	case call := <-ch:
		if call.Err != nil {
			common_utils.IncCounter("DBUS fail")
			return call.Err
		}
		result := call.Body
		if len(result) == 0 {
			common_utils.IncCounter("DBUS fail")
			return fmt.Errorf("Dbus result is empty %v", result)
		}
		if ret, ok := result[0].(int32); ok {
			if ret == 0 {
				return nil
			} else {
				if len(result) != 2 {
					common_utils.IncCounter("DBUS fail")
					return fmt.Errorf("Dbus result is invalid %v", result)
				}
				if msg, check := result[1].(string); check {
					common_utils.IncCounter("DBUS fail")
					return fmt.Errorf(msg)
				} else {
					common_utils.IncCounter("DBUS fail")
					return fmt.Errorf("Invalid result message type %v %v", result[1], reflect.TypeOf(result[1]))
				}
			}
		} else {
			common_utils.IncCounter("DBUS fail")
			return fmt.Errorf("Invalid result type %v %v", result[0], reflect.TypeOf(result[0]))
		}
	case <-time.After(time.Duration(timeout) * time.Second):
		log.V(2).Infof("DbusApi: timeout")
		common_utils.IncCounter("DBUS fail")
		return fmt.Errorf("Timeout %v", timeout)
	}
	return nil
}

func (c *DbusClient) ConfigReload(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS config reload")
		return nil
	}
	modName := "config"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".reload"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}

func (c *DbusClient) ConfigSave(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS config save")
		return nil
	}
	modName := "config"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".save"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}

func (c *DbusClient) ApplyPatchYang(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS apply patch yang")
		return nil
	}
	modName := "gcu"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".apply_patch_yang"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}

func (c *DbusClient) ApplyPatchDb(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS apply patch db")
		return nil
	}
	modName := "gcu"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".apply_patch_db"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}

func (c *DbusClient) CreateCheckPoint(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS create checkpoint")
		return nil
	}
	modName := "gcu"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".create_checkpoint"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}

func (c *DbusClient) DeleteCheckPoint(fileName string) error {
	if c.testMode == true {
		common_utils.IncCounter("DBUS delete checkpoint")
		return nil
	}
	modName := "gcu"
	busName := c.busNamePrefix + modName
	busPath := c.busPathPrefix + modName
	intName := c.intNamePrefix + modName + ".delete_checkpoint"
	err := DbusApi(busName, busPath, intName, 10, fileName)
	return err
}
