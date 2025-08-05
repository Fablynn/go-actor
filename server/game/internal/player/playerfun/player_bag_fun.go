package playerfun

import (
	"go-actor/common/pb"
	"go-actor/library/uerror"
	"go-actor/server/game/internal/player/domain"

	"google.golang.org/protobuf/proto"
)

// PlayerBagFun 背包管理
type PlayerBagFun struct {
	*PlayerFun
	data *pb.PlayerDataBag
}

func NewPlayerBagFun(fun *PlayerFun) domain.IPlayerFun {
	return &PlayerBagFun{PlayerFun: fun}
}

func (p *PlayerBagFun) Load(msg *pb.PlayerData) error {
	if msg == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家数据为空")
	}
	if msg.Bag == nil {
		msg.Bag = &pb.PlayerDataBag{Items: make(map[uint32]*pb.PbItem)}
	}
	p.data = msg.Bag
	return nil
}

func (p *PlayerBagFun) Save(msg *pb.PlayerData) error {
	if msg == nil {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "玩家数据为空")
	}
	buf, err := proto.Marshal(p.data)
	if err != nil {
		return err
	}
	newBase := &pb.PlayerDataBag{}
	if err := proto.Unmarshal(buf, newBase); err != nil {
		return err
	}
	msg.Bag = newBase
	return nil
}

func (p *PlayerBagFun) Finish() {
	//todo save data to disk
}

// 查询道具
func (p *PlayerBagFun) GetItem(propID uint32) *pb.PbItem {
	if item, ok := p.data.Items[propID]; ok {
		return item
	}
	return nil
}

// 添加道具
func (p *PlayerBagFun) AddItem(propID uint32, val int64) {
	if item := p.GetItem(propID); item != nil {
		item.Count += val
	} else {
		p.data.Items[propID] = &pb.PbItem{
			PropId: propID,
			Count:  val,
		}
	}
	p.Change()
}

// 消耗道具
func (p *PlayerBagFun) SubItem(propID uint32, val int64) error {
	if val == 0 {
		return nil
	}
	if val < 0 {
		return uerror.New(pb.ErrorCode_PARAM_INVALID, "消耗道具参数负值")
	}
	item := p.GetItem(propID)
	if item == nil || item.Count < val {
		return uerror.New(pb.ErrorCode_GAME_PROP_NOT_ENOUGH, "游戏道具不足")
	}
	item.Count -= val
	p.Change()
	return nil
}

// 背包查询接口
func (p *PlayerBagFun) GetBagReq(head *pb.Head, req *pb.GetBagReq, rsp *pb.GetBagRsp) error {
	for _, item := range p.data.Items {
		rsp.List = append(rsp.List, item)
	}
	return nil
}

// 背包奖励接口
func (p *PlayerBagFun) RewardReq(head *pb.Head, req *pb.RewardReq, rsp *pb.RewardRsp) error {
	tmps := map[uint32]*pb.Reward{}
	for _, rw := range req.RewardList {
		if rw.Incr < 0 {
			return uerror.New(pb.ErrorCode_PARAM_INVALID, "奖励增加数量不能小于0")
		}
		if val, ok := tmps[rw.PropId]; ok {
			val.Incr += rw.Incr
		} else {
			tmps[rw.PropId] = rw
		}
	}
	for _, rw := range tmps {
		p.AddItem(rw.PropId, rw.Incr)
		p.Change()
		rw.Total = p.GetItem(rw.PropId).Count
		rsp.RewardList = append(rsp.RewardList, rw)
	}
	return nil
}

// 道具消耗请求
func (p *PlayerBagFun) ConsumeReq(head *pb.Head, req *pb.ConsumeReq, rsp *pb.ConsumeRsp) error {
	tmps := map[uint32]*pb.Reward{}
	for _, rw := range req.RewardList {
		if rw.Incr < 0 {
			return uerror.New(pb.ErrorCode_PARAM_INVALID, "消耗的数量不能小于0")
		}
		if val, ok := tmps[rw.PropId]; ok {
			val.Incr += rw.Incr
		} else {
			tmps[rw.PropId] = rw
		}
		item := p.GetItem(rw.PropId)
		if item == nil || item.Count < tmps[rw.PropId].Incr {
			return uerror.New(pb.ErrorCode_GAME_PROP_NOT_ENOUGH, "玩家道具数量不足%v", rw.PropId)
		}
	}
	for _, rw := range tmps {
		p.SubItem(rw.PropId, rw.Incr)
		p.Change()
		rw.Total = p.GetItem(rw.PropId).Count
		rsp.RewardList = append(rsp.RewardList, rw)
	}
	return nil
}
