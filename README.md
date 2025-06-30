# go-actor

**这是一款分布式的golang游戏服务器框架 基于golang + actor model技术构建 它具备高性能、可伸缩、分布式、协程分组管理等特点。并且上手简单、易学**

框架示意图：

![pic.jpg](./blob/pic.jpg)

## **快速开始**

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
head := framework.NewHead(dst, pb.RouterType, uint64(actorId), actorName, FuncName)
```

同服务异步通讯

```
actor.SendMsg(head, req, rsp)
```

时间轮毫秒级定时器，有效降低golang自带四叉树最小堆计时器高度

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
