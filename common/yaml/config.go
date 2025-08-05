package yaml

import (
	"go-actor/common/pb"
	"go-actor/library/uerror"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// 主从配置
type SlaveConfig struct {
	DbName   string `yaml:"dbname"`
	Db       int    `yaml:"db"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
}

// 数据库配置
type DbConfig struct {
	DbName   string                 `yaml:"dbname"`
	Db       int                    `yaml:"db"`
	User     string                 `yaml:"user"`
	Password string                 `yaml:"password"`
	Host     string                 `yaml:"host"`
	Slave    map[int32]*SlaveConfig `yaml:"slave"`
}

// 服务注册与发现配置
type EtcdConfig struct {
	Topic     string   `yaml:"topic"`
	Endpoints []string `yaml:"endpoints"`
}

// nats配置
type NatsConfig struct {
	Topic     string `yaml:"topic"`
	Endpoints string `yaml:"endpoints"`
}

// 通用配置
type CommonConfig struct {
	Pprof     bool   `yaml:"pprof"`
	Env       string `yaml:"env"`
	SecretKey string `yaml:"secret_key"`
}

// 游戏配置
type DataConfig struct {
	IsRemote  bool     `yaml:"is_remote"`
	Path      string   `yaml:"path"`
	Topic     string   `yaml:"topic"`
	Endpoints []string `yaml:"endpoints"`
}

// 节点配置
type NodeConfig struct {
	RouterTTL    int64  `yaml:"router_ttl"`
	DiscoveryTTL int64  `yaml:"discovery_ttl"`
	LogLevel     string `yaml:"log_level"`
	LogPath      string `yaml:"log_path"`
	Ip           string `yaml:"ip"`
	Port         int    `yaml:"port"`
	HttpPort     int    `yaml:"http_port"`
}

type Config struct {
	Mysql   map[int32]*DbConfig   `yaml:"mysql"`
	Redis   map[int32]*DbConfig   `yaml:"redis"`
	Mongodb map[int32]*DbConfig   `yaml:"mongodb"`
	Etcd    *EtcdConfig           `yaml:"etcd"`
	Nats    *NatsConfig           `yaml:"nats"`
	Common  *CommonConfig         `yaml:"common"`
	Data    *DataConfig           `yaml:"data"`
	Client  map[int32]*NodeConfig `yaml:"client"`
	Gate    map[int32]*NodeConfig `yaml:"gate"`
	Room    map[int32]*NodeConfig `yaml:"room"`
	Match   map[int32]*NodeConfig `yaml:"match"`
	Db      map[int32]*NodeConfig `yaml:"db"`
	Builder map[int32]*NodeConfig `yaml:"builder"`
	Game    map[int32]*NodeConfig `yaml:"game"`
	Gm      map[int32]*NodeConfig `yaml:"gm"`
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
		return nil, nil, uerror.New(pb.ErrorCode_PARSE_FAILED, "配置文件加载失败: %v", err)
	}
	var ok bool
	var srvCfg *NodeConfig
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
		return nil, nil, uerror.New(pb.ErrorCode_CONFIG_NOT_FOUND, "配置文件中未找到节点配置: %s", nodeType.String())
	}
	return cfg, &pb.Node{
		Name: strings.TrimPrefix(nodeType.String(), "NodeType"),
		Type: nodeType,
		Id:   nodeId,
		Ip:   srvCfg.Ip,
		Port: int32(srvCfg.Port),
	}, nil
}
