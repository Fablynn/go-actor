package manager

import (
	"fmt"
	"net/http"
	"poker_server/common/pb"
	"poker_server/framework/actor"
	"poker_server/library/async"
	"poker_server/library/mlog"
	"poker_server/server/gate/internal/player"
	"reflect"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type GatePlayerMgr struct {
	actor.Actor
	mgr    *actor.ActorMgr // 玩家管理器
	status int32           // 运行状态
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10 * 1024,
	WriteBufferSize: 10 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 初始化ws
func (d *GatePlayerMgr) Init(ip string, port int) error {
	// 初始化ActorMgr
	d.mgr = new(actor.ActorMgr)
	player := &player.Player{}
	d.mgr.Register(player)
	d.mgr.ParseFunc(reflect.TypeOf(player))
	actor.Register(d.mgr)

	// 初始化Actor
	d.Actor.Register(d)
	d.Actor.ParseFunc(reflect.TypeOf(d))
	d.Actor.Start()
	actor.Register(d)

	// 启动ws服务
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

	async.SafeGo(mlog.Errorf, func() {
		mlog.Infof("启动WebSocket服务, 地址: %s:%d", ip, port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			mlog.Errorf("WebSocket服务启动失败:%v", err)
		}
		atomic.StoreInt32(&d.status, 0)
	})
	atomic.AddInt32(&d.status, 1)
	mlog.Infof("WebSocket服务(%s:%d)启动成功，等待连接中!!!", ip, port)
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
	// 新建玩家
	usr := player.NewPlayer(conn, &player.Frame{})
	if err := usr.CheckToken(); err != nil {
		mlog.Errorf("玩家登录失败: %v", err)
		conn.Close()
		return
	}

	// 检查玩家是否已存在
	if act := d.mgr.GetActor(usr.GetId()); act != nil {
		d.Add(1)
		async.SafeGo(mlog.Errorf, func() {
			act.Stop()
			d.Done()
		})
	}

	// 登录成功，添加玩家
	usr.Start()
	d.mgr.AddActor(usr)
	mlog.Infof("客户端: %d(%s)连接成功!!!", usr.GetId(), conn.RemoteAddr().String())

	// 循环接受消息
	usr.Dispatcher()
}

// 剔除玩家
func (d *GatePlayerMgr) Kick(head *pb.Head) {
	if act := d.mgr.GetActor(head.Uid); act != nil {
		// 删除玩家
		d.mgr.DelActor(head.Uid)

		// 等待消息处理完成，然后关闭连接
		d.Add(1)
		async.SafeGo(mlog.Errorf, func() {
			act.Stop()
			d.Done()
		})
	}
}
