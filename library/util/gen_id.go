package util

import (
	"github.com/sony/sonyflake"
	"go-actor/common/pb"
	"time"
)

type TIDGen struct {
	idGen *sonyflake.Sonyflake
}

func NewIDGen(machineId uint16) (idg *TIDGen, err error) {
	idg = &TIDGen{
		idGen: newSonyFlake(machineId),
	}
	return
}
func (i *TIDGen) GenID() (uint64, error) {
	return i.idGen.NextID()
}

func newSonyFlake(MachineID uint16) *sonyflake.Sonyflake {
	return sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID: func() (uint16, error) {
			return MachineID, nil
		},
		CheckMachineID: nil,
	})
}

func DestructRoomId(roomId uint64) *pb.DefaultRoomId {
	return &pb.DefaultRoomId{
		GameType:  pb.GameType(roomId >> 32 & 0xFF),
		CoinType:  pb.CoinType(roomId >> 24 & 0xFF),
		MatchType: pb.MatchType(roomId >> 40 & 0xFF),
		Incr:      uint32(roomId & 0xFFFFFF),
	}
}

func GenMatchId(types *pb.DefaultRoomId) uint64 {
	switch types.GetGameType() {
	default: // rummy texas类型
		return uint64(types.GameType)<<32 | uint64(types.CoinType)
	}
}
