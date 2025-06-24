package yaml

import (
	"fmt"
	"os"
	"poker_server/common/pb"
	"poker_server/library/uerror"
	"strings"

	"gopkg.in/yaml.v3"
)

type SlaveConfig struct {
	DbName   string `yaml:"dbname"`
	Db       int    `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
}

type DbConfig struct {
	DbName   string                 `yaml:"dbname"`
	Db       int                    `yaml:"db"`
	User     string                 `yaml:"user"`
	Password string                 `yaml:"password"`
	Host     string                 `yaml:"host"`
	Slave    map[int32]*SlaveConfig `yaml:"slave"`
}

type EtcdConfig struct {
	Topic     string   `yaml:"topic"`
	Endpoints []string `yaml:"endpoints"`
}

type NatsConfig struct {
	Topic     string `yaml:"topic"`
	Endpoints string `yaml:"endpoints"`
}

type CommonConfig struct {
	Env             string `yaml:"env"`
	ConfigIsRemote  bool   `yaml:"config_is_remote"`
	ConfigPath      string `yaml:"config_path"`
	ConfigTopic     string `yaml:"config_topic"`
	RouterExpire    int64  `yaml:"router_expire"`
	DiscoveryExpire int64  `yaml:"discovery_expire"`
	SecretKey       string `yaml:"secret_key"`
}

type ServerConfig struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
	Ip       string `yaml:"ip"`
	Port     int    `yaml:"port"`
	HttpPort int    `yaml:"http_port"`
}

type Config struct {
	Mysql   map[int32]*DbConfig     `yaml:"mysql"`
	Redis   map[int32]*DbConfig     `yaml:"redis"`
	Mongodb map[int32]*DbConfig     `yaml:"mongodb"`
	Etcd    *EtcdConfig             `yaml:"etcd"`
	Nats    *NatsConfig             `yaml:"nats"`
	Common  *CommonConfig           `yaml:"common"`
	Client  map[int32]*ServerConfig `yaml:"client"`
	Gate    map[int32]*ServerConfig `yaml:"gate"`
	Room    map[int32]*ServerConfig `yaml:"room"`
	Match   map[int32]*ServerConfig `yaml:"match"`
	Db      map[int32]*ServerConfig `yaml:"db"`
	Builder map[int32]*ServerConfig `yaml:"builder"`
	Game    map[int32]*ServerConfig `yaml:"game"`
	Gm      map[int32]*ServerConfig `yaml:"gm"`
}

func (c *Config) Unmarshal(buf []byte) error {
	return yaml.Unmarshal(buf, c)
}

func NewConfig(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := new(Config)
	if err = cfg.Unmarshal(content); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadConfig(filename string, nodeType pb.NodeType, nodeId int32) (*Config, *pb.Node, error) {
	cfg, err := NewConfig(filename)
	if err != nil {
		return nil, nil, uerror.New(1, pb.ErrorCode_PARSE_FAILED, "配置文件加载失败: %v", err)
	}
	var ok bool
	var srvCfg *ServerConfig
	switch nodeType {
	case pb.NodeType_NodeTypeGate:
		srvCfg, ok = cfg.Gate[nodeId]
	case pb.NodeType_NodeTypeRoom:
		srvCfg, ok = cfg.Room[nodeId]
	case pb.NodeType_NodeTypeMatch:
		srvCfg, ok = cfg.Match[nodeId]
	case pb.NodeType_NodeTypeDb:
		srvCfg, ok = cfg.Db[nodeId]
	case pb.NodeType_NodeTypeBuilder:
		srvCfg, ok = cfg.Builder[nodeId]
	case pb.NodeType_NodeTypeGame:
		srvCfg, ok = cfg.Game[nodeId]
	case pb.NodeType_NodeTypeGm:
		srvCfg, ok = cfg.Gm[nodeId]
	case pb.NodeType_NodeTypeClient:
		srvCfg, ok = cfg.Client[nodeId]
	}
	if !ok {
		return nil, nil, uerror.New(1, pb.ErrorCode_CONFIG_NOT_FOUND, "配置文件中未找到节点配置: %s", nodeType.String())
	}
	return cfg, &pb.Node{
		Name: strings.TrimPrefix(nodeType.String(), "NodeType"),
		Type: nodeType,
		Id:   nodeId,
		Addr: fmt.Sprintf("%s:%d", srvCfg.Ip, srvCfg.Port),
	}, nil
}
