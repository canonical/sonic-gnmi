package gnmi

import (
	"sync"
	"time"
	"net"
	"regexp"
	log "github.com/golang/glog"

	"github.com/go-redis/redis"
	sdcfg "github.com/sonic-net/sonic-gnmi/sonic_db_config"
)

const table = "TELEMETRY_CONNECTIONS"

var rclient *redis.Client

type ConnectionManager struct {
	connections  map[string]struct{}
	mu           sync.Mutex
	threshold    int
}

func (cm *ConnectionManager) PrepareRedis() {
	ns := sdcfg.GetDbDefaultNamespace()
	rclient = redis.NewClient(&redis.Options{
		Network:     "tcp",
		Addr:        sdcfg.GetDbTcpAddr("STATE_DB", ns),
		Password:    "",
		DB:          sdcfg.GetDbId("STATE_DB", ns),
		DialTimeout: 0,
	})

	res, _ := rclient.HGetAll("TELEMETRY_CONNECTIONS").Result()

	if res == nil {
		return
	}

	for key, _ := range res {
		rclient.HDel(table, key)
	}
}

func (cm *ConnectionManager) Add(addr net.Addr, query string) (string, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if len(cm.connections) >= cm.threshold && cm.threshold != 0 { // 0 is defined as no threshold
		log.V(1).Infof("Cannot add another client connection as threshold is already at limit")
		return "", false
	}
	key := createKey(addr, query)
	log.V(1).Infof("Adding client connection: %s", key)
	cm.connections[key] = struct{}{}
	log.V(1).Infof("Current number of existing connections: %d", len(cm.connections))
	storeKeyRedis(key)
	return key, true
}

func (cm *ConnectionManager) Remove(key string) (bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	_, exists := cm.connections[key]
	if exists {
		log.V(1).Infof("Closing connection: %s", key)
		delete(cm.connections, key)
	}
	deleteKeyRedis(key)
	return exists
}

func createKey(addr net.Addr, query string) string {
	regexStr := "(?:target|element):\"([a-zA-Z0-9-_*]*)\""
	regex := regexp.MustCompile(regexStr)
	matches := regex.FindAllStringSubmatch(query, -1)
	// connectionKeyString will look like "10.0.0.1|OTHERS|proc|uptime|2017-07-04 00:47:20
	connectionKey := addr.String() + "|"
	for i := 0; i < len(matches); i++ {
		if len(matches[i]) < 2 {
			continue
		}
		connectionKey += matches[i][1] // index 1 contains the value we need
		connectionKey += "|"
	}
	connectionKey += time.Now().UTC().Format(time.RFC3339)
	return connectionKey
}

func storeKeyRedis(key string) {
	ret, err := rclient.HSet(table, key, "active").Result()
	if !ret {
		log.V(1).Infof("Subscribe client failed to update telemetry connection key:%s err:%v", key, err)
	}
}

func deleteKeyRedis(key string) {
	ret, err := rclient.HDel(table, key).Result()
	if ret == 0 {
		log.V(1).Infof("Subscribe client failed to delete telemetry connection key:%s err:%v", key, err)
	}
}
