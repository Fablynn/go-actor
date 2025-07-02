package card

import (
	"fmt"
	"go-actor/common/pb"
	"strings"
)

var (
	WildFlag   uint32 = 1 << 31
	color_name        = map[pb.ColorType]string{
		pb.ColorType_ColorTypeNone: "$", // 花色类型-无 → 空
		pb.ColorType_Club:          "♣", // 花色类型-梅花 → ♣
		pb.ColorType_Heart:         "♥", // 花色类型-红桃 → ♥
		pb.ColorType_Spade:         "♠", // 花色类型-黑桃 → ♠
		pb.ColorType_Diamond:       "♦", // 花色类型-方块 → ♦
	}

	Color_value = map[string]uint32{
		"♦": 8,
		"♣": 1,
		"♥": 2,
		"♠": 4,
	}

	str_to_color = map[string]uint32{
		"1": 1,
		"2": 2,
		"3": 4,
		"4": 8,
	}
	rank_name = map[pb.RankType]string{
		pb.RankType_RANK_0:        "0",  // 牌值-0（新增的零值）
		pb.RankType_RANK_1:        "A",  // 牌值-1 → A
		pb.RankType_RANK_2:        "2",  // 牌值-2 → 2
		pb.RankType_RANK_3:        "3",  // 牌值-3 → 3
		pb.RankType_RANK_4:        "4",  // 牌值-4 → 4
		pb.RankType_RANK_5:        "5",  // 牌值-5 → 5
		pb.RankType_RANK_6:        "6",  // 牌值-6 → 6
		pb.RankType_RANK_7:        "7",  // 牌值-7 → 7
		pb.RankType_RANK_8:        "8",  // 牌值-8 → 8
		pb.RankType_RANK_9:        "9",  // 牌值-9 → 9
		pb.RankType_RANK_10:       "10", // 牌值-10 → 10
		pb.RankType_RANK_J:        "J",  // 牌值-J → J
		pb.RankType_RANK_Q:        "Q",  // 牌值-Q → Q
		pb.RankType_RANK_K:        "K",  // 牌值-K → K
		pb.RankType_RANK_A:        "A",  // 牌值-A → A（与RankType_One相同）
		pb.RankType_RANK_JOKER:    "$",
		pb.RankType_RANK_BIGJOKER: "￥",
	}

	Rank_value = map[string]uint32{
		"2":  2,
		"3":  3,
		"4":  4,
		"5":  5,
		"6":  6,
		"7":  7,
		"8":  8,
		"9":  9,
		"10": 10,
		"J":  11,
		"Q":  12,
		"K":  13,
		"A":  14,
		"$":  15,
		"￥":  16,
	}
)

// 扑克牌通用定义
type Card uint32

type CardList []uint32

func (c Card) Color() pb.ColorType {
	return pb.ColorType(c >> 16 & 0x0F)
}

func (c Card) Rank() pb.RankType {
	return pb.RankType(c & 0x0F)
}

func (c Card) Bit() uint32 {
	move := c.Rank() - 1
	if move <= 0 { // 大王
		return 0
	}
	return 1 << move
}

func (c Card) Value() uint32 {
	return uint32(c)
}

func (c Card) String() string {
	if c.IsWild() {
		return fmt.Sprintf("%s%s%s", rank_name[c.Rank()], color_name[c.ClearWild().Color()], color_name[0])
	}
	return fmt.Sprintf("%s%s", rank_name[c.Rank()], color_name[c.Color()])
}

func (c Card) AddWild() uint32 {
	return uint32(c) | WildFlag
}

func (c Card) IsWild() bool {
	return (uint32(c) & WildFlag) != 0
}

func (c Card) ClearWild() Card {
	c = Card(uint32(c) &^ WildFlag)
	return c
}

func (c CardList) String() string {
	strs := []string{}
	for _, v := range c {
		strs = append(strs, Card(v).String())
	}
	return strings.Join(strs, ",")
}

func StrToCard(strs ...string) (rets []uint32) {
	for _, item := range strs {
		item = strings.ToUpper(item)
		ll := len(item)
		val, ok1 := Rank_value[item[:ll-1]]
		col, ok2 := str_to_color[item[ll-1:]]
		if !ok1 || !ok2 {
			continue
		}
		rets = append(rets, val|(col<<16))
	}
	if len(rets) != len(strs) {
		rets = rets[:0]
	}
	return
}
