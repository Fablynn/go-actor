package reward

import "go-actor/common/pb"

func ToReward(coinType pb.CoinType, count int64) *pb.Reward {
	return &pb.Reward{
		PropId: uint32(coinType),
		Incr:   count,
	}
}

func ToConsumeRequest(coinType pb.CoinType, count int64) *pb.ConsumeReq {
	return &pb.ConsumeReq{
		RewardList: []*pb.Reward{
			{
				PropId: uint32(coinType),
				Incr:   count,
			},
		},
	}
}
