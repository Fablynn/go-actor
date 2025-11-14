package rummy_rebot

import (
	"fmt"
	"testing"
	"time"
)

func TestTrain(t *testing.T) {
	t.Run("train_auto_declare", func(t *testing.T) {
		start := time.Now() // 记录开始时间
		Train()
		elapsed := time.Since(start) // 计算耗时
		fmt.Printf("该操作耗时: %v\n", elapsed)
	})
}
func Train() {
	cards := GetRummyCardsByCfg(2, 2)
	Shuffle(cards, 3)

	idx := uint32(0)
	ghostCard := Card(cards[idx]).AddWild()
	idx++
	for i := idx; i < uint32(len(cards)); i++ {
		if Card(cards[i]).Rank() == Card(ghostCard).Rank() {
			cards[i] = Card(cards[i]).AddWild() //相同卡翻赖处理
		}
	}
	round := 1
	showCard := cards[idx] //明牌一张
	idx++
	handCards := make([]uint32, 0, 13) //13张手牌
	handCards = append(handCards, cards[idx:idx+13]...)
	idx += 13

	fmt.Printf("Race Init 手牌:%v 可吃明牌:%v 百搭牌：%v \n", CardList(handCards), Card(showCard), Card(ghostCard))
	for {
		fmt.Printf("\n\nNew Round:%d 手牌:%v 可吃明牌:%v \n", round, CardList(handCards), Card(showCard))
		if choose := filterChoose(append(handCards, showCard)); choose.unqiue != nil && len(choose.unqiue) <= 1 {

			if len(choose.unqiue) == 0 { // win
				var item uint32
				choose.group, item = backUnique(choose.group)
				choose.unqiue = []uint32{item}
				fmt.Printf("完美胜利牌型： 组合牌：%v 单牌:%v \n", showGroup(choose.group), CardList(choose.unqiue).String())
				break
			}
			fmt.Printf("吃牌胜利牌型： 组合牌：%v 单牌:%v \n", showGroup(choose.group), CardList(choose.unqiue).String())
			break
		}
		if lose := decision(handCards, showCard); lose > 0 {
			handCards = getDiff(append(handCards, showCard), []uint32{lose})
		} else {
			idx++
			drawCard := cards[idx]
			handCards = append(handCards, drawCard)
			choose := filterChoose(handCards)
			lose = outputDecision(choose)
			if lose == 0 {
				if len(choose.unqiue) == 0 { // win
					var item uint32
					choose.group, item = backUnique(choose.group)
					choose.unqiue = []uint32{item}
					fmt.Printf("摸牌胜利牌型： 组合牌：%v 单牌:%v \n", showGroup(choose.group), CardList(choose.unqiue).String())
					break
				}
			}
			handCards = sortHand(getDiff(handCards, []uint32{lose}))
			fmt.Printf("Round:%d 手牌:%v 抽牌:%v \n", round, CardList(handCards), Card(drawCard))
		}
		round++
		idx++
		showCard = cards[idx]
	}
}
