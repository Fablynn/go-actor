package actor

import (
	"poker_server/common/pb"
	"reflect"
	"testing"
)

type Player struct {
	Actor
}

func (p *Player) SendToClient(head *pb.Head, msg interface{}) error {
	return nil
}

func (d *Player) JoinRoomReq(head *pb.Head, req *pb.RummyJoinRoomReq, rsp *pb.RummyJoinRoomRsp) error {
	return nil
}

func TestActor(t *testing.T) {
	aa := &Player{}
	aa.Register(aa)
	aa.ParseFunc(reflect.TypeOf(aa))

}
