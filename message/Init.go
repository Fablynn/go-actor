package message

import (
	"go-actor/common/pb"
	"go-actor/framework"
)

// Init 路由消息表 打到不同节点
func Init() {
	framework.RegisterCmd(pb.NodeType_NodeTypeGate, pb.RouterType_UID, pb.CMD_CMDGATE_LOGIN_REQ, "Player.Login")        // 登录请求
	framework.RegisterCmd(pb.NodeType_NodeTypeGate, pb.RouterType_UID, pb.CMD_CMDKICK_PLAYER_NTF, "Player.Kick")        // 剔除玩家通知
	framework.RegisterCmd(pb.NodeType_NodeTypeGame, pb.RouterType_UID, pb.CMD_CMDGATE_HEART_REQ, "Player.HeartRequest") // 心跳请求
}
