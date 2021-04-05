package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	Env         = flag.String("env", "development", "env")
	ServerID    = flag.String("id", "", "env")
	Host        = flag.String("host", "127.0.0.1", "server mqtt host for cluster") //
	Port        = 0                                                                // flag.Uint("metric.namespace", 0, "mqtt server port for cluster") // 不能定制
	ClientPort  = flag.Uint("clientPort", 0, "ws server port for cluster")         //
	Frontend    = true
	ServerType  = flag.String("serverType", "connector", "usage of this server")
	HostName, _ = os.Hostname()

	Pid = os.Getpid()

	CreatedAt = time.Now()

	// Config is config
	//Config     C
	InternalIP string

	GracefulShutDown = flag.Uint("shutdownTick", 2, "graceful shut down")
	WsPath           = flag.String("ws.path", "/", "path which ws served at")
	ProtoPath        = flag.String("protos.dir", "", "protos dir path")

	RedisAddr = flag.String("redis.addr", "127.0.0.1:6379", "main redis addr")
	RedisPwd  = flag.String("redis.passwd", "", "main redis passwd")
	RedisDB   = flag.Int("redis.db", 3, "main redis db")

	ChannelRedisAddr = flag.String("channel.redis.addr", "127.0.0.1:6379", "global channel redis addr")
	ChannelRedisPwd  = flag.String("channel.redis.passwd", "", "global channel redis passwd")
	ChannelRedisDB   = flag.Int("channel.redis.db", 2, "global channel redis db")

	Amqp        = flag.String("amqp", "amqp://guest:guest@localhost:5672/", "rabbitmq connect url")
	MonitorHost = flag.String("masterHost", "127.0.0.1", "master server Host")
	MonitorPort = flag.Int("masterPort", 3005, "master server Port")
)

// C  config
//type C struct {
//	// Services map[string]Services `yml:"services"`
//	WsPath           string `yml:"wsPath"`
//	GracefulShutDown int    `yml:"gracefulShutDown"`
//	WsPort           string `yml:"wsPort"`
//}

func init() {
	flag.Parse()

	if *ServerID == "" {
		t := time.Now()
		id := fmt.Sprintf("%s-%s-%d%d%d%d%d%d-%05d", *ServerType, HostName, t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), os.Getpid())
		ServerID = &id
	}

	//if *CfgPath != "" {
	//	err := cfg.LoadFile(*CfgPath) //路径是以 main 文件指定的
	//	if err != nil {
	//		os.Exit(-1)
	//	}
	//	e := cfg.Scan(&Config)
	//	fmt.Println(e)
	//}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				InternalIP = ipnet.IP.String()
			}
		}
	}

	fmt.Println("internalIP is", setColor(InternalIP, 0, 0, 32))
	if InternalIP == "" {
		InternalIP = "127.0.0.1"
		fmt.Println("internalIP not found, use", setColor(InternalIP, 0, 0, 31))
	}
}

func Uptime() float64 {
	t := time.Since(CreatedAt).Minutes()
	if n, e := strconv.ParseFloat(fmt.Sprintf("%.2f", t), 10); e == nil {
		return n
	}
	return t
}

func setColor(msg string, conf, bg, text int) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, conf, bg, text, msg, 0x1B)
}
