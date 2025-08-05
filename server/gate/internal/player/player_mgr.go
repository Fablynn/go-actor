package player

import (
	"fmt"
	"go-actor/common/pb"
	"go-actor/framework/actor"
	"go-actor/framework/recycle"
	"go-actor/library/mlog"
	"go-actor/library/safe"
	"go-actor/server/gate/internal/frame"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type GatePlayerMgr struct {
	actor.Actor
	mgr    *actor.ActorMgr // 玩家管理器
	status int32           // 运行状态
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	WriteBufferPool: &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (d *GatePlayerMgr) start(ip string, port int) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&d.status) <= 0 {
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil || conn == nil {
			mlog.Errorf("WebSocket连接失败: %v", err)
			return
		}
		d.accept(conn)
	})

	// 启动ws服务
	safe.Go(func() {
		mlog.Infof("启动WebSocket服务, 地址: %s:%d", ip, port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			mlog.Errorf("WebSocket服务启动失败:%v", err)
		} else {
			atomic.StoreInt32(&d.status, 0)
		}
	})
	atomic.AddInt32(&d.status, 1)
	mlog.Infof("WebSocket服务(%s:%d)启动成功，等待连接中!!!", ip, port)
}

// 初始化ws
func (d *GatePlayerMgr) Init(ip string, port int) error {
	d.mgr = new(actor.ActorMgr)
	pl := &Player{}
	d.mgr.Register(pl)
	d.mgr.ParseFunc(reflect.TypeOf(pl))
	actor.Register(d.mgr)

	// 初始化Actor
	d.Actor.Register(d)
	d.Actor.ParseFunc(reflect.TypeOf(d))
	d.Actor.Start()
	actor.Register(d)
	d.start(ip, port)
	return nil
}

func (d *GatePlayerMgr) Stop() {
	// 设置状态为停止
	atomic.StoreInt32(&d.status, 0)

	// 停止所有玩家
	d.mgr.Stop()

	// 停止Actor
	d.Actor.Stop()
	mlog.Infof("WebSocket服务已停止")
}

func (d *GatePlayerMgr) accept(conn *websocket.Conn) {
	usr := NewPlayer(conn, &frame.Frame{})
	if err := usr.CheckToken(); err != nil {
		recycle.Destroy(usr)
		mlog.Errorf("玩家登录失败: %v", err)
		return
	}

	// 检查玩家是否已存在
	if act := d.mgr.GetActor(usr.GetId()); act != nil {
		act.SendMsg(&pb.Head{FuncName: "Kick"}, usr.GetExtra())
		mlog.Errorf("玩家被顶号: %v", act.GetId())
		recycle.Destroy(act.(*Player))
	}

	// 登录成功，添加玩家
	uid := usr.GetId()
	d.mgr.AddActor(usr)
	mlog.Infof("客户端: %d(%s)连接成功!!!", uid, conn.RemoteAddr().String())

	// 循环接受消息
	usr.Dispatcher()
	mlog.Infof("关闭websocket协程：%d", uid)
}

// 剔除玩家
func (d *GatePlayerMgr) Kick(head *pb.Head) {
	if act := d.mgr.GetActor(head.Uid); act != nil {
		// 删除玩家
		d.mgr.DelActor(head.Uid)
		// 等待消息处理完成，然后关闭连接
		recycle.Destroy(act.(*Player))
	}
}
