package network

import (
	"encoding/binary"
	"go-actor/common/pb"
	"go-actor/framework/domain"
	"go-actor/library/uerror"
	"go-actor/library/util"
	"time"

	"github.com/gorilla/websocket"
)

const (
	CLIENT_MAX_PACKET = 10 * 1024 // 10kb
)

var (
	pools = util.PoolSlice[byte](CLIENT_MAX_PACKET)
)

func get() []byte {
	return pools.Get().([]byte)
}

func put(buf []byte) {
	pools.Put(buf)
}

type Socket struct {
	frame   domain.IFrame   // 帧协议
	maxSize int             // 最大包长限制
	conn    *websocket.Conn // 通信
	wbytes  []byte          // 接受缓存
}

func NewSocket(conn *websocket.Conn, frame domain.IFrame) *Socket {
	return &Socket{
		frame:   frame,
		maxSize: CLIENT_MAX_PACKET,
		conn:    conn,
		wbytes:  get(),
	}
}

func (d *Socket) Close() error {
	defer put(d.wbytes)
	return d.conn.Close()
}

func (d *Socket) newWrite(size int) []byte {
	return d.wbytes[:size]
}

// 设置接受超时时间，避免阻塞
func (d *Socket) SetReadExpire(expire int64) {
	if expire > 0 {
		d.conn.SetReadDeadline(time.Now().Add(time.Duration(expire) * time.Second))
	} else {
		d.conn.SetReadDeadline(time.Time{})
	}
}

// 设置发送超时时间，避免阻塞
func (d *Socket) SetWriteExpire(expire int64) {
	if expire > 0 {
		d.conn.SetWriteDeadline(time.Now().Add(time.Duration(expire) * time.Second))
	} else {
		d.conn.SetWriteDeadline(time.Time{})
	}
}

func (d *Socket) Write(pack *pb.Packet) error {
	// 获取数据帧长度
	ll := d.frame.GetSize(pack)
	if ll+4 > d.maxSize {
		return uerror.New(1, pb.ErrorCode_MAX_SIZE_LIMIT, "超过最大包长限制: %d", d.maxSize)
	}

	// 组包
	buf := d.newWrite(ll + 4)
	binary.BigEndian.PutUint32(buf, uint32(ll))
	if err := d.frame.Encode(pack, buf[4:]); err != nil {
		return err
	}

	// 发送数据包
	return d.conn.WriteMessage(websocket.BinaryMessage, buf)
}

func (d *Socket) Read(recv *pb.Packet) error {
	// 读取总长度
	_, buf, err := d.conn.ReadMessage()
	if err != nil {
		return uerror.New(1, pb.ErrorCode_READ_FAIELD, "读取数据包失败: %v", err)
	}
	return d.frame.Decode(buf[4:], recv)
}
