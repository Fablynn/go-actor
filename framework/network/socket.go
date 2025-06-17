package network

import (
	"encoding/binary"
	"poker_server/common/pb"
	"poker_server/framework/domain"
	"poker_server/library/uerror"
	"time"

	"github.com/gorilla/websocket"
)

// frame组成部分说明： len(四个字节表示包长) | cmd（4byte） | uid(8byte) | routeId(8byte) | seq(4byte) |version(4byte) | extra(4byte) | body(数据包体)
// frame组成： 4+head+body    第一个4字节存储 head+body的总长度长度
type Socket struct {
	frame   domain.IFrame   // 帧协议
	maxSize int             // 最大包长限制
	conn    *websocket.Conn // 通信
	rbytes  []byte          // 接受缓存
	sbytes  []byte          // 接受缓存
}

func NewSocket(conn *websocket.Conn, maxSize int) *Socket {
	return &Socket{
		maxSize: maxSize,
		conn:    conn,
		rbytes:  make([]byte, maxSize/2),
		sbytes:  make([]byte, maxSize/2),
	}
}

func (d *Socket) Register(frame domain.IFrame) {
	d.frame = frame
}

func (d *Socket) Close() error {
	return d.conn.Close()
}

func (d *Socket) newRead(size int) (ret []byte) {
	if cap(d.sbytes) < size {
		d.sbytes = make([]byte, size)
	}
	ret = d.sbytes[:size]
	return
}

func (d *Socket) newSend(size int) (ret []byte) {
	if cap(d.rbytes) < size {
		d.rbytes = make([]byte, size)
	}
	ret = d.rbytes[:size]
	return
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
	if ll > d.maxSize {
		return uerror.New(1, pb.ErrorCode_MAX_SIZE_LIMIT, "超过最大包长限制: %d", d.maxSize)
	}

	// 组包
	buf := d.newSend(ll + 4)
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
