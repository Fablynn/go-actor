package domain

import "go-actor/common/pb"

type IPlayerFun interface {
	Load(*pb.PlayerData) error // 加载数据
	Save(*pb.PlayerData) error // 保存数据
	Complete()                 // 加载完成
	Finish()                   // 结束
}
