# go-actor [Golang德州扑克游戏案例]

**这是一款分布式的golang游戏服务器框架**

特性：

- [x] 高性能

- [x] 协程安全

- [x] 跨服务同步、异步无感通信

- [x] 毫秒级时间轮

- [x] 游戏状态机管理

- [x] 多节点snowflask uuid

- [x] websocket协议和protobuf编码

- [x] 高性能日志库mlog

- [x] 德州扑克游戏案例

框架示意图：

![pic.jpg](./blob/pic.jpg)

## **快速开始**

### 安装启动

```
安装最新protoc
download for https://github.com/protocolbuffers/protobuf/releases
protoc --version
libprotoc 31.0

安装golang语言 1.24.3+:
https://go.dev/dl/

安装protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

安装protoc-go-inject-tag
go install github.com/favadi/protoc-go-inject-tag@latest

cd go-actor/tools/protoc-gen-xorm
go install

查看安装：
ls $(go env GOPATH)/bin
protoc-gen-go  protoc-gen-xorm  protoc-go-inject-tag

临时添加环境:
export PATH=$PATH:$(go env GOPATH)/bin

安装docker-composer: 

以上准备完毕后:
快速启动所有服务: 
make docker_run && make config && make start_all

快速终止所有服务:
make stop_all && make docker_stop
```

### 德州扑克案例

游戏日志:

```
Texas deal hard cards notify per user : 145 hardcards: 2♥,10♦
Texas deal hard cards notify per user : 146 hardcards: 2♣,J♣
Texas deal hard cards notify per user : 144 hardcards: A♣,J♦
flop_state show public cards : 9♦,3♥,4♦
turn_state show public cards : 9♦,3♥,4♦,6♦
river_state show public cards : 9♦,3♥,4♦,6♦,8♦
texas_end_state show users uid:144 best cards: J♦,9♦,8♦,6♦,4♦ FLUSH
texas_end_state show users uid:145 best cards: 10♦,9♦,8♦,6♦,4♦ FLUSH
texas_end_state show users uid:146 best cards: J♣,9♦,8♦,6♦,4♦ HIGH_CARD
texas_end_state chips:map[144:-100000 145:-100000 146:-100000], srvs:map[144:200], winners:map[144:299800 145:0 146:0]
```

### 服务相关

状态机注册

```
// 注册方法
machine.RegisterState(pb.GameState_DEMO_STAGE_INIT, &demo.Initstate{})

// 启动方法
// onTick()
machine.NewMachine(nowMs, pb.GameState_DEMO_STAGE_INIT, game.(*unsafe.Pointer))

// 状态实例
type Initstate struct {
	BaseState
}


func (d *Initstate) OnEnter(nowMs int64, curState pb.GameState, extra interface{}) {
	game := extra.(*demo.DemoGame)

	// 重置房间状态
	game.Data.Stage = curState

	// init 初始化游戏
	if game.Data.Common.GameFinish {
		game.Reset() //重置房间
	}
}

func (d *Initstate) OnTick(nowMs int64, curState pb.GameState, extra interface{}) pb.GameState {
	return moveToState.(pb.GameState)
}

func (d *Initstate) OnExit(nowMs int64, curState pb.GameState, extra interface{}) {
    // log
}

```



新建一个非网关服务

```
func main() {
    var cfg string
    var nodeId int
    flag.StringVar(&cfg, "config", "config.yaml", "游戏配置文件")
    flag.IntVar(&nodeId, "id", 1, "服务ID")
    flag.Parse()

    // 加载游戏配置
    yamlcfg, node, err := yaml.LoadConfig(cfg, pb.NodeType_NodeTypeRoom, int32(nodeId))
    if err != nil {
        panic(fmt.Sprintf("游戏配置加载失败: %v", err))
    }
    nodeCfg := yamlcfg.Room[node.Id]

    // 初始化日志库
    if err := mlog.Init(yamlcfg.Common.Env, nodeCfg.LogLevel, nodeCfg.LogFile); err != nil {
        panic(fmt.Sprintf("日志库初始化失败: %v", err))
    }
    async.Init(mlog.Errorf)

    // 初始化游戏配置
    mlog.Infof("初始化游戏配置")
    if err := config.Init(yamlcfg.Etcd, yamlcfg.Common); err != nil {
        panic(err)
    }

    // 初始化redis
    mlog.Infof("初始化redis配置")
    if err := dao.InitRedis(yamlcfg.Redis); err != nil {
        panic(fmt.Sprintf("redis初始化失败: %v", err))
    }

    // 初始化框架
    mlog.Infof("启动框架服务: %v", node)
    if err := framework.InitDefault(node, nodeCfg, yamlcfg); err != nil {
        panic(fmt.Sprintf("框架初始化失败: %v", err))
    }

    // 功能模块初始化 todo
    if err := manager.Init(); err != nil {
        panic(fmt.Sprintf("功能模块初始化失败: %v", err))
    }

    // 服务退出
    signal.SignalNotify(func() {
        manager.Close()
        framework.Close()
        mlog.Close()
    })
}
```

跨服务同步通讯 

```
dst := framework.NewGameRouter(playerId, "Player", "ConsumeReq")
newHead := framework.NewHead(dst, pb.RouterType_RouterTypeUid, playerId)
rsp := &pb.ConsumeRsp{}
if err := framework.Request(newHead, req, rsp); err != nil {
    mlog.Infof("Request Error: %v", err)
}
```

跨服务异步通讯

```
newHead := framework.NewHead(dst, pb.RouterType_RouterTypeUid, playerId)
framework.Send(newHead , req)
```

携带自动返回的异步通讯

```
// demoActMgr
dst := framework.NewDemoRouter(uint64(demo.id), "targetMgr", "targetFunc")
head := framework.NewHead(dst, pb.RouterType, uint64(actorId), demoActMgr, FuncName)

// 异步请求方法
func (m *targetMgr) targetFunc(head *pb.Head, req *pb.GetTargetFuncReq, rsp *pb.GetTargetFuncRsp) error {
    // about get rsp ....
    rsp.data = m.data
	return nil
}

// 异步返回处理
func (d *demoActMgr) FuncName(head *pb.Head, rsp *pb.GetTargetFuncRsp) error{
    // get rsp data todo ...
	return nil
}
```

同服务异步通讯

```
actor.SendMsg(head, req, rsp)
```

毫秒级定时器-时间轮，可有效降低golang自带四叉树最小堆计时器高度

```
m.RegisterTimer(&pb.Head{
    SendType:  pb.SendType_POINT,
    ActorName: "DbRummyRoomMgr",
    FuncName:  "OnTick",
}, 5*time.Second, -1)
```

创建一个actor，通过反射自动绑定路由

```创建一个actor
ret.Actor.Register(ret)
ret.Actor.ParseFunc(reflect.TypeOf(ret))
ret.SetId(uint64(pb.DataType_DataTypeReport))
ret.Start()
actor.Register(ret)
```

## **扩展工具**

### pbtool:

通过标签可自动生成pb对象，redis服务序列化、反序列化工具类

```
//@pbtool:[string|hash]|db_name|fieldName:fieldType|#备注
// 示例注释规则 @pbtool 表示protobuf对象参与注释解析 redis工具模板
// [string|hash] 表示protobuf对象序列化存储的两种模板
// db_name 指定存储db
// fieldName1:fieldType1[,fieldName2:fieldType2] 索引字段类型
// #备注 标签

@pbtool:string|poker|generator|#房间id生成器
@pbtool:hash|poker|user_info|uid@uint64|#玩家永久缓存信息
@pbtool:hash|poker|texas|RoomId@uint64|#德州游戏房间信息数据
```

### cfgtool:

解析文件table对象为指定pb文件

```
枚举类型说明：
E|道具类型-金币|PropertType|Coin|1    

配置规则说明：
@config|sheet@结构名|map:字段名[,字段名]:别名|group:字段名[,字段名]:别名
map: 工具类依据字段名筛选配置数据 多个字段名符复合筛选

example:
@config:table_cfg|网关接口路由表:RouterConfig|map:Cmd|map:NodeType,ActorName,FuncName

result make file content :
func MGetCmd(Cmd uint32) *pb.RouterConfig {
    obj, ok := obj.Load().(*RouterConfigData)
    if !ok {
        return nil
    }
    if val, ok := obj._Cmd[Cmd]; ok {
        return val
    }
    return nil
}

func MGetNodeTypeActorNameFuncName(NodeType pb.NodeType, ActorName string, FuncName string) *pb.RouterConfig {
    obj, ok := obj.Load().(*RouterConfigData)
    if !ok {
        return nil
    }
    if val, ok := obj._NodeTypeActorNameFuncName[pb.Index3[pb.NodeType, string, string]{NodeType, ActorName, FuncName}]; ok {
        return val
    }
    return nil
}

@struct|sheet@结构名
@enum|sheet
```