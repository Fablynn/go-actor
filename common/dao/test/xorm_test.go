package test

import (
	"poker_server/common/config"
	"poker_server/common/dao"
	"poker_server/common/dao/domain"
	"poker_server/common/pb"
	"poker_server/common/yaml"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	dao.RegisterMysqlTable(domain.MYSQL_DB_PLAYER_DATA, &pb.PlayerData{})
}

func TestMain(m *testing.M) {
	if err := config.InitConfig("../../../../poker_gameconf/data", nil); err != nil {
		panic(err)
	}
	cfg, err := yaml.NewConfig("../../../env/local/config.yaml")
	if err != nil {
		panic(err)
	}
	if err := dao.InitMysql(cfg.Mysql); err != nil {
		panic(err)
	}
	m.Run()
}

func TestPlayer(t *testing.T) {
	usr := &pb.PlayerData{
		Uid: 1,
		Base: &pb.PlayerDataBase{
			PlayerInfo: &pb.PlayerInfo{
				Uid:      1,
				NickName: "test_user",
				Avatar:   "avatar.png",
			},
			CreateTime: 1633036800,
			LoginInfo: &pb.PlayerLoginInfo{
				LastLoginTime:  1633036800,
				LastLogoutTime: 1633036800,
				NowLoginTime:   1633036800,
			},
		},
	}
	session := dao.GetMysql(domain.MYSQL_DB_PLAYER_DATA).NewSession()
	defer session.Close()

	// 查询
	ok, err := session.Get(usr)
	t.Log("查询", ok, err, usr)

	// 插入
	if !ok {
		cc, err := session.Insert(usr)
		t.Log("插入", cc, err, usr)
	}

	// 更新
	usr1 := &pb.PlayerData{
		Id:  usr.Id,
		Uid: 1,
		Base: &pb.PlayerDataBase{
			PlayerInfo: &pb.PlayerInfo{
				Uid:      1,
				NickName: "testxxuser",
				Avatar:   "avar.png",
			},
			CreateTime: 1633036800,
			LoginInfo: &pb.PlayerLoginInfo{
				LastLoginTime:  1633036800,
				LastLogoutTime: 1633036800,
				NowLoginTime:   1633036800,
			},
		},
		Bag: &pb.PlayerDataBag{Items: map[uint32]*pb.PbItem{
			1001: {PropId: 1001, Count: 10},
		}},
		Version: usr.Version,
	}
	ss, err := session.ID(usr1.Id).Update(usr1)
	t.Log("更新", ss, err, usr1)

	/*
		// 删除
		newUser := &pb.PlayerData{Uid: usr.Uid}
		cc, err := session.Delete(newUser)
		t.Log("删除", cc, err, newUser, usr)
	*/
}
