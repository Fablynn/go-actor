package texas

import (
	"go-actor/common/card"
	"go-actor/common/pb"
	"sort"
)

type CardGroup struct {
	value    uint32
	cardType pb.CardType
	cards    []uint32
}

type BackTrackComapre struct {
	cards []uint32
}

func Compare(cards ...uint32) (pb.CardType, uint32, []uint32) {
	ret := &BackTrackComapre{}
	return ret.Get(cards...)
}

func (b *BackTrackComapre) Get(cards ...uint32) (pb.CardType, uint32, []uint32) {
	b.cards = make([]uint32, 0)

	sort.Slice(cards, func(i, j int) bool {
		return card.Card(cards[i]).Rank() > card.Card(cards[j]).Rank()
	})

	combinations := []*CardGroup{}
	// 穷举所有五张牌的组合
	b.backtrack(cards, 0, &combinations)

	// 排序
	sort.Slice(combinations, func(i, j int) bool {
		return combinations[i].value > combinations[j].value
	})
	return combinations[0].cardType, combinations[0].value, combinations[0].cards
}

func (b *BackTrackComapre) backtrack(cards []uint32, start int, data *[]*CardGroup) {
	if len(b.cards) == 5 {
		cardType, value := GetCardType(b.cards...)
		*data = append(*data, &CardGroup{
			value:    value,
			cardType: cardType,
			cards:    append([]uint32{}, b.cards...),
		})
		return
	}

	for i := start; i < len(cards); i++ {
		b.cards = append(b.cards, cards[i])
		b.backtrack(cards, i+1, data)
		b.cards = b.cards[:len(b.cards)-1]
	}
}

// 从大到小排列
func GetCardType(vals ...uint32) (pb.CardType, uint32) {
	sort.Slice(vals, func(i, j int) bool {
		return card.Card(vals[i]).Rank() > card.Card(vals[j]).Rank()
	})
	c0, c1, c2, c3, c4 := card.Card(vals[0]), card.Card(vals[1]), card.Card(vals[2]), card.Card(vals[3]), card.Card(vals[4])
	r0, r1, r2, r3, r4 := c0.Rank(), c1.Rank(), c2.Rank(), c3.Rank(), c4.Rank()
	b0, b1, b2, b3, b4 := c0.Bit(), c1.Bit(), c2.Bit(), c3.Bit(), c4.Bit()
	color := pb.ColorType(c0.Color() | c1.Color() | c2.Color() | c3.Color() | c4.Color())
	rank := (b0 | b1 | b2 | b3 | b4) >> (c4.Rank() - 1)
	// 同花顺
	if (color == pb.ColorType_Diamond || color == pb.ColorType_Club || color == pb.ColorType_Heart || color == pb.ColorType_Spade) && (rank == 0x1F || rank == 0x100F) {
		if c0.Rank() == pb.RankType_RANK_A && rank == 0x1F {
			if rank == 0x100F {
				return pb.CardType_STRAIGHT_FLUSH, uint32(pb.CardType_STRAIGHT_FLUSH)<<(4*7) + uint32(1) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
			}
			return pb.CardType_ROYAL_FLUSH, uint32(pb.CardType_ROYAL_FLUSH)<<(4*7) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
		}
		if rank == 0x100F {
			return pb.CardType_STRAIGHT_FLUSH, uint32(pb.CardType_STRAIGHT_FLUSH)<<(4*7) + uint32(1) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
		}
		return pb.CardType_STRAIGHT_FLUSH, uint32(pb.CardType_STRAIGHT_FLUSH)<<(4*7) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
	}
	// 四条
	if b0|b1|b2|b3 == b3 {
		return pb.CardType_FOUR_OF_A_KIND, uint32(pb.CardType_FOUR_OF_A_KIND)<<(4*7) + uint32(r0|r1|r2|r3)<<(4*5) + uint32(r4)
	}
	if b1|b2|b3|b4 == b4 {
		return pb.CardType_FOUR_OF_A_KIND, uint32(pb.CardType_FOUR_OF_A_KIND)<<(4*7) + uint32(r1|r2|r3|r4)<<(4*5) + uint32(r0)
	}
	// 三带二
	if b0|b1|b2 == b2 && b3|b4 == b4 {
		return pb.CardType_FULL_HOUSE, uint32(pb.CardType_FULL_HOUSE)<<(4*7) + uint32(r0|r1|r2)<<(4*5) + uint32(r3|r4)<<(4*3)
	}
	if b2|b3|b4 == b4 && b0|b1 == b1 {
		return pb.CardType_FULL_HOUSE, uint32(pb.CardType_FULL_HOUSE)<<(4*7) + uint32(r2|r3|r4)<<(4*5) + uint32(r0|r1)<<(4*3)
	}
	// 同花
	if color == pb.ColorType_Diamond || color == pb.ColorType_Club || color == pb.ColorType_Heart || color == pb.ColorType_Spade {
		return pb.CardType_FLUSH, uint32(pb.CardType_FLUSH)<<(4*7) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
	}
	// 顺子
	if rank == 0x1F {
		return pb.CardType_STRAIGHT, uint32(pb.CardType_STRAIGHT)<<(4*7) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
	}
	if rank == 0x100F {
		return pb.CardType_STRAIGHT, uint32(pb.CardType_STRAIGHT)<<(4*7) + uint32(1) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
	}
	// 三条
	if b0|b1|b2 == b2 {
		return pb.CardType_THREE_OF_A_KIND, uint32(pb.CardType_THREE_OF_A_KIND)<<(4*7) + uint32(r0|r1|r2)<<(4*5) + uint32(1<<r3) + uint32(1<<r4)
	}
	if b1|b2|b3 == b3 {
		return pb.CardType_THREE_OF_A_KIND, uint32(pb.CardType_THREE_OF_A_KIND)<<(4*7) + uint32(r1|r2|r3)<<(4*5) + uint32(1<<r0) + uint32(1<<r4)
	}
	if b2|b3|b4 == b4 {
		return pb.CardType_THREE_OF_A_KIND, uint32(pb.CardType_THREE_OF_A_KIND)<<(4*7) + uint32(r2|r3|r4)<<(4*5) + uint32(1<<r0) + uint32(1<<r1)
	}
	// 两队
	if b0|b1 == b1 && b2|b3 == b3 {
		return pb.CardType_TWO_PAIR, uint32(pb.CardType_TWO_PAIR)<<(4*7) + uint32(r0|r1)<<(4*5) + uint32(r2|r3)<<(4*5) + uint32(r4)
	}
	if b0|b1 == b1 && b3|b4 == b4 {
		return pb.CardType_TWO_PAIR, uint32(pb.CardType_TWO_PAIR)<<(4*7) + uint32(r0|r1)<<(4*5) + uint32(r3|r4)<<(4*5) + uint32(r2)
	}
	if b1|b2 == b2 && b3|b4 == b4 {
		return pb.CardType_TWO_PAIR, uint32(pb.CardType_TWO_PAIR)<<(4*7) + uint32(r1|r2)<<(4*5) + uint32(r3|r4)<<(4*5) + uint32(r0)
	}
	// 一对
	if b0|b1 == b1 {
		return pb.CardType_ONE_PAIR, uint32(pb.CardType_ONE_PAIR)<<(4*7) + uint32(r0|r1)<<(4*5) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4) //uint32(r2|r3|r4)
	}
	if b1|b2 == b2 {
		return pb.CardType_ONE_PAIR, uint32(pb.CardType_ONE_PAIR)<<(4*7) + uint32(r1|r2)<<(4*5) + uint32(1<<r0) + uint32(1<<r3) + uint32(1<<r4) //| uint32(r0|r3|r4)
	}
	if b2|b3 == b3 {
		return pb.CardType_ONE_PAIR, uint32(pb.CardType_ONE_PAIR)<<(4*7) + uint32(r2|r3)<<(4*5) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r4) //| uint32(r0|r1|r4)
	}
	if b3|b4 == b4 {
		return pb.CardType_ONE_PAIR, uint32(pb.CardType_ONE_PAIR)<<(4*7) + uint32(r3|r4)<<(4*5) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) //| uint32(r0|r1|r2)
	}
	// 高牌
	return pb.CardType_HIGH_CARD, uint32(pb.CardType_HIGH_CARD)<<(4*7) + uint32(1<<r0) + uint32(1<<r1) + uint32(1<<r2) + uint32(1<<r3) + uint32(1<<r4)
}
