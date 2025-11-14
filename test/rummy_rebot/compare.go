package rummy_rebot

import "go-actor/library/random"

func GetRummyCardsByCfg(decks, joker uint32) (cards []uint32) {
	for time := uint32(0); time < decks; time++ {
		for i := uint32(0); i <= 3; i++ {
			for j := uint32(2); j <= 14; j++ {
				cards = append(cards, (1<<(16+i))|j)
			}
		}
	}
	for i := uint32(0); i < joker; i++ {
		if i%2 == 0 {
			cards = append(cards, (1<<(31))|15) //small joker
		} else {
			cards = append(cards, (1<<(31))|16) //big joker
		}
	}
	return
}

// 洗牌
func Shuffle(cards []uint32, times int) {
	// 洗牌
	for j := 0; j < times; j++ {
		for i := 0; i < len(cards); i++ {
			pos := random.Intn(len(cards))
			cards[i], cards[pos] = cards[pos], cards[i]
		}
	}
}

func countZerosBetweenOnes(n uint32) (result int) {
	count := 0
	inSequence := false // 是否已经遇到第一个 1

	for i := 0; i < 14; i++ {
		if (n & (1 << i)) != 0 { // 当前位是 1
			if inSequence && count > 0 {
				result += count // 记录 0 的数量
			}
			inSequence = true
			count = 0 // 重置计数
		} else if inSequence { // 当前位是 0，且之前遇到了 1
			count++
		}
	}

	return
}
