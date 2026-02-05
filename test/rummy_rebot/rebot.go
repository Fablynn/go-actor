package rummy_rebot

import (
	"fmt"
	"go-actor/common/pb"
	"math/bits"
	"slices"
)

type RummyGroup struct {
	group  [][]uint32 //刻子或顺子
	unqiue []uint32   //散牌
}

func dfsWildVaild(cards []uint32, wilds []uint32, pre RummyGroup) RummyGroup { //散牌
	cards = sortHand(cards)
	preNum := len(cards)
	for len(wilds) > 0 { // 没有百搭  或者达成胜利条件确定
		cards, wilds, pre.group = matchWild(cards, wilds, pre.group)

		if len(cards) == preNum || len(cards) <= 1 {
			break
		}
		preNum = len(cards)
	}
	for len(wilds) >= 2 && len(cards) >= 2 { //单牌和多个赖子 组合 转二进制求gap 如果gap >= 4 就不能组成顺子 2个百搭解决一个单牌
		patch := filterUnique(cards)
		var group []uint32
		if len(patch) == 0 || isSeq(patch[0]) >= len(wilds) { // 最优解无法匹配
			group = []uint32{cards[0], wilds[0], wilds[1]}
			wilds = wilds[2:]
			cards = cards[1:]
			pre.group = append(pre.group, group)
		} else {
			group = append(patch[0], wilds[:len(patch[0])]...)
			cards = getDiff(cards, patch[0])
		}
	}
	pre.unqiue = cards
	return pre
}

func explodePair(cards []uint32, seq [][]uint32) (ans [][]uint32, unique []uint32) {
	ranks := map[pb.RankType]int{}
	for _, card := range cards {
		ranks[Card(card).Rank()] |= int(Card(card).Color())
	}
	for rank, color := range ranks { // pairs
		if bits.OnesCount(uint(color)) == 2 {
			ans = append(ans, bitsKernighan(color, int(rank)))
		}
	}
	setMx := 2
	matchSet := make(map[pb.RankType]int)
	for _, group := range seq {
		if isSet(group) {
			setMx--
		} else {
			canUse := explianSeq(group)
			for _, card := range canUse {
				matchSet[Card(card).Rank()] |= int(Card(card).Color())
			}
		}
	}
	cards = append(cards, 0)
	var st []uint32
	var wild = true
	for _, card := range cards {
		if len(st) > 0 && (Card(card).Color() != Card(st[len(st)-1]).Color() || !checkNear(st[len(st)-1], card, wild)) { //跨颜色 或者不连续 清空栈
			if len(st) >= 2 {
				ans = append(ans, slices.Clone(st))
			} else { //特判 利用一个百搭可以形成更长的顺子 和借用一个多余的顺子牌形成刻子
				if checkSeq(st[0], seq) >= 0 || (setMx > 0 && matchSet[Card(st[0]).Rank()] > 0 && (int(Card(st[0]).Color())&matchSet[Card(st[0]).Rank()] == 0)) {
					ans = append(ans, slices.Clone(st))
				}
			}
			st = st[:0]
			wild = true
		}
		if len(st) > 0 && Card(card).Rank()-Card(st[len(st)-1]).Rank() > 1 {
			wild = false
		}

		if Card(card).Rank() == pb.RankType_RANK_2 && ranks[pb.RankType_RANK_A]&int(Card(card).Color()) > 0 {
			st = append(st, uint32(int(Card(card).Color())<<16|int(pb.RankType_RANK_A)))
		}
		st = append(st, card)
	}

	unique = cards[:len(cards)-1]
	if len(ans) != 0 {
		slices.SortFunc(ans, func(i, j []uint32) int {
			if len(j) != len(i) {
				return len(j) - len(i)
			} else {
				a, b := isSeq(i), isSeq(j)
				if abs(a-b) == 13 {
					return rankSum(i) - rankSum(j)
				}
				return a - b
			}
		})

		for _, group := range ans {
			unique = getDiff(unique, group)
		}
	}

	slices.SortFunc(unique, func(i, j uint32) int {
		return int(Card(i).Rank()) - int(Card(j).Rank())
	})

	fmt.Printf("优化出牌逻辑： 成对牌：%v 单牌：%v \n", showGroup(ans), CardList(unique).String())
	return
}

func abs(n int) int {
	if n < 0 {
		n = -n
	}
	return n
}

func matchWild(cards, wilds []uint32, seq [][]uint32) ([]uint32, []uint32, [][]uint32) { // 匹配一个百搭可以成立组合 贪心加反悔
	var ans [][]uint32
	ranks := map[pb.RankType]int{}
	for _, card := range cards {
		ranks[Card(card).Rank()] |= int(Card(card).Color())
	}
	for rank, color := range ranks { // pairs
		if bits.OnesCount(uint(color)) == 2 {
			ans = append(ans, bitsKernighan(color, int(rank)))
		}
	}
	setMx := 2
	matchSet := make(map[pb.RankType]int)
	for _, group := range seq {
		if isSet(group) {
			setMx--
		} else {
			canUse := explianSeq(group)
			for _, card := range canUse {
				matchSet[Card(card).Rank()] |= int(Card(card).Color())
			}
		}
	}
	cards = append(cards, 0)
	var st []uint32
	var wild = true
	for _, card := range cards {
		if len(st) > 0 && (Card(card).Color() != Card(st[len(st)-1]).Color() || !checkNear(st[len(st)-1], card, wild)) { //跨颜色 或者不连续 清空栈
			if len(st) >= 2 {
				ans = append(ans, slices.Clone(st))
			} else { //特判 利用一个百搭可以形成更长的顺子 和借用一个多余的顺子牌形成刻子
				if checkSeq(st[0], seq) >= 0 || (setMx > 0 && matchSet[Card(st[0]).Rank()] > 0 && (int(Card(st[0]).Color())&matchSet[Card(st[0]).Rank()] == 0)) {
					ans = append(ans, slices.Clone(st))
				}
			}
			st = st[:0]
			wild = true
		}
		if len(st) > 0 && Card(card).Rank()-Card(st[len(st)-1]).Rank() > 1 {
			wild = false
		}

		if Card(card).Rank() == pb.RankType_RANK_2 && ranks[pb.RankType_RANK_A]&int(Card(card).Color()) > 0 {
			st = append(st, uint32(int(Card(card).Color())<<16|int(pb.RankType_RANK_A)))
		}
		st = append(st, card)
	}

	if len(ans) != 0 {
		slices.SortFunc(ans, func(i, j []uint32) int {
			if len(j) != len(i) {
				return len(j) - len(i)
			} else {
				return int(Card(i[0]).Rank() - Card(j[0]).Rank())
			}
		})
		match := ans[0]
		cards = getDiff(cards, match)
		if len(match) < 2 {
			m := match[0]
			if i := checkSeq(m, seq); i >= 0 { //搭桥 原地修改seq
				seq[i] = append(seq[i], m, wilds[0])
			} else if setMx > 0 { // 借用多余顺子牌 凑刻子
				for s := range seq {
					canUse := explianSeq(seq[s])
					for _, card := range canUse {
						if isSet([]uint32{card, m}) {
							seq[s] = getDiff(seq[s], []uint32{card})
							match = append(match, card)
						}
					}
				}
				seq = append(seq, append(match, wilds[0]))
			}
		} else {
			seq = append(seq, append(match, wilds[0]))
		}
		wilds = wilds[1:]
	}
	return cards[:len(cards)-1], wilds, seq
}

func checkSeq(card uint32, seq [][]uint32) (result int) {
	for i, group := range seq {
		if len(group) < 3 {
			continue
		}
		if isSeq(append(group, card)) <= 1 {
			return i
		}
	}
	return -1
}

func isSeq(cards []uint32) (gap int) {
	var color, wilds int
	var bit uint32
	var hasAce bool
	for _, card := range cards {
		if Card(card).IsWild() {
			wilds++
			continue
		}
		if Card(card).Rank() == pb.RankType_RANK_A {
			hasAce = true
		}
		color |= int(Card(card).Color())
		bit |= Card(card).Bit()
	}
	if bits.OnesCount(uint(color)) != 1 || bits.OnesCount32(bit)+wilds != len(cards) { //检验颜色和重复牌
		return 14
	}
	gap = countZerosBetweenOnes(bit)
	if hasAce {
		bit |= 1
		gap = min(gap, countZerosBetweenOnes(bit&^(1<<13)))
	}
	return max(gap-wilds, 0)
}
func isSet(cards []uint32) (result bool) {
	var pre int
	for _, card := range cards {
		if Card(card).IsWild() {
			continue
		}
		if pre == 0 {
			pre = int(Card(card).Rank())
		} else {
			return pre == int(Card(card).Rank())
		}
	}
	return
}

func filterUnique(cards []uint32) [][]uint32 {
	var match [][]uint32
	var st []uint32
	for _, card := range cards {
		if len(st) > 0 && Card(card).Color() != Card(st[len(st)-1]).Color() {
			if len(st) == 2 && isSeq(st) < 4 {
				match = append(match, slices.Clone(st))
			}
			st = st[:0]
		}
		if len(st) == 2 && isSeq(st) < 4 {
			match = append(match, slices.Clone(st))
			st = st[1:]
		}
		st = append(st, card)
	}

	var fliter [][]uint32
	for _, group := range match {
		if len(fliter) > 0 {
			last := fliter[len(fliter)-1]
			if Card(last[0]).Color() == Card(group[0]).Color() && last[1] == group[0] {
				if isSeq(group) < isSeq(last) {
					fliter = fliter[:len(fliter)-1]
				} else {
					continue
				}
			}
		}
		fliter = append(fliter, group)
	}
	slices.SortFunc(fliter, func(i, j []uint32) int {
		return isSeq(i) - isSeq(j)
	})
	return fliter
}

func explianSeq(cards []uint32) (list []uint32) { //抽取不影响顺子组成的牌
	if len(cards) < 4 {
		return
	}
	for i, card := range cards {
		check := append(append([]uint32{}, cards[:i]...), cards[i+1:]...)
		if isSeq(check) == 0 {
			list = append(list, card)
		}
	}
	return
}

func checkNear(pre, cur uint32, wild bool) bool {
	return Card(cur).Rank()-Card(pre).Rank() == 1 || (wild && Card(cur).Rank()-Card(pre).Rank() == 2) || (Card(pre).Rank() == pb.RankType_RANK_2 && Card(cur).Rank() == pb.RankType_RANK_A) || (wild && Card(pre).Rank() == pb.RankType_RANK_3 && Card(cur).Rank() == pb.RankType_RANK_A)
}

func dfsVaild(cards []uint32) (groupCards []RummyGroup) { // 回溯实现提取顺子剪枝 求刻子和散牌
	groups := bsFindSeq(cards)
	if len(groups) > 0 {
		for _, group := range groups {
			diff := getDiff(cards, group)
			result := dfsVaild(diff)
			for i := range result {
				result[i].group = append(result[i].group, group)
			}
			groupCards = append(groupCards, result...)
		}
	} else {
		cnt := make(map[pb.RankType]int, 13)
		for _, card := range cards {
			cnt[Card(card).Rank()] |= int(Card(card).Color())
		}

		for rank, color := range cnt {
			if !isRummySet(color) {
				continue
			} else {
				var sets [][]uint32
				if color == 15 { //散牌有同花色炸弹元素凑搭子
					setItems := bitsKernighan(color, int(rank))
					sets = append(sets, setItems)

					for i, setItem := range setItems {
						if findNearRank(setItem, cnt) { //刻子某个元素能和其他单牌凑搭子
							item := slices.Clone(setItems)
							sets = append(sets, append(item[:i], item[i+1:]...))
						}
					}
				} else {
					sets = append(sets, bitsKernighan(color, int(rank)))
				}

				for _, group := range sets {
					diff := getDiff(cards, group)
					result := dfsVaild(diff)
					for i := range result {
						result[i].group = append(result[i].group, group)
					}
					groupCards = append(groupCards, result...)
				}
				return
			}
		}

		return []RummyGroup{{unqiue: cards}}
	}
	return
}

func findNearRank(card uint32, cnt map[pb.RankType]int) (result bool) {
	for rank := max(2, int(Card(card).Rank())-2); rank <= min(14, int(Card(card).Rank())+2); rank++ {
		if rank != int(Card(card).Rank()) {
			result = result || (int(Card(card).Color())&cnt[pb.RankType(rank)] != 0)
		}
	}

	if Card(card).Rank() == pb.RankType_RANK_A {
		for rank := 2; rank <= 3; rank++ {
			result = result || (int(Card(card).Color())&cnt[pb.RankType(rank)] != 0)
		}
	}
	return
}

func bitsKernighan(color, rank int) (ret []uint32) {
	for color > 0 {
		ret = append(ret, uint32(color&-color<<16|rank))
		color = color & (color - 1)
	}
	return
}

func bsFindSeq(cards []uint32) (ans [][]uint32) {
	n := len(cards)
	if n < 3 {
		return
	}

	ace := map[pb.ColorType]uint32{}
	for _, card := range cards {
		if Card(card).Rank() == pb.RankType_RANK_A {
			ace[Card(card).Color()] = card
		}
	}
	var st []uint32
	var dfs func(int)
	dfs = func(pos int) {
		if len(st) >= 3 {
			prueSeq := slices.Clone(st)
			ans = append(ans, prueSeq)
		}

		if len(st) >= 2 && Card(st[len(st)-1]).Rank() == pb.RankType_RANK_2 && ace[Card(st[len(st)-1]).Color()] > 0 { //或者A23或更大
			prueSeq := slices.Clone(st)
			prueSeq = append(prueSeq, ace[Card(st[len(st)-1]).Color()])
			ans = append(ans, prueSeq)
		}

		if pos == -1 || (len(ans) > 0 && Card(ans[0][0]).Color() != Card(cards[pos]).Color()) { //跨颜色就终止 避免组合嵌套
			return
		}

		for j := pos; j >= 0; j-- {
			if len(st) > 0 && !(Card(cards[j]).Rank() == Card(st[len(st)-1]).Rank()-1 && Card(cards[j]).Color() == Card(st[len(st)-1]).Color()) || (j < n-1 && cards[j] == cards[j+1]) {
				continue
			}
			st = append(st, cards[j])
			dfs(j - 1)

			st = st[:len(st)-1]
		}
	}
	dfs(n - 1)
	return
}
func getDiff(cards []uint32, group []uint32) (unique []uint32) {
	cnt := map[uint32]bool{}
	for _, card := range group {
		cnt[card] = true
	}
	for _, card := range cards {
		if cnt[card] {
			delete(cnt, card)
			continue
		} else {
			unique = append(unique, card)
		}
	}
	return unique
}

func getWilds(cards []uint32) (unique, wilds []uint32) {
	for _, card := range cards {
		if Card(card).IsWild() {
			wilds = append(wilds, card)
		} else {
			unique = append(unique, card)
		}
	}
	return
}

func isRummySet(color int) bool {
	return color == 15 || color == 14 || color == 13 || color == 11 || color == 7
}

func checkRank(a, b []uint32) (sum int) { // 对比rank差值
	for i := range a {
		rankA, rankB := max(int(Card(a[i]).Rank()), 10), max(int(Card(b[i]).Rank()), 10)
		sum += rankA - rankB
	}
	return sum
}

func backUnique(groups [][]uint32) ([][]uint32, uint32) {
	var pop uint32
	for i, group := range groups {
		if len(group) >= 4 {
			if isSet(group) {
				pop = group[0]
				groups[i] = group[1:]
				return groups, pop
			} else {
				use := explianSeq(group)
				if len(use) > 0 {
					pop = use[0]
					groups[i] = getDiff(group, []uint32{pop})
					return groups, pop
				}
			}
		}
	}
	return groups, pop
}

func showGroup(groups [][]uint32) (ret string) {
	for _, group := range groups {
		ret += CardList(group).String() + " "
	}
	return
}

func filterChoose(real []uint32) (bestChoose RummyGroup) {
	cards := make([]uint32, len(real))
	copy(cards, real)
	cards = sortHand(cards)
	prueSeq := bsFindSeq(cards)
	prueSeq = append(prueSeq, []uint32{})
	for _, seq := range prueSeq {
		var unqiue, wilds []uint32
		unqiue = getDiff(cards, seq)
		unqiue, wilds = getWilds(unqiue)
		groups := dfsVaild(unqiue)
		for _, group := range groups {
			choose := dfsWildVaild(group.unqiue, wilds, group)

			if bestChoose.unqiue == nil || len(choose.unqiue) < len(bestChoose.unqiue) || len(choose.unqiue) == len(bestChoose.unqiue) && checkRank(bestChoose.unqiue, choose.unqiue) > 0 {
				choose.group = append(choose.group, seq)
				if len(choose.unqiue) <= 1 {
					return choose
				}
				bestChoose = choose
			}
		}
	}
	return
}

func sortHand(card []uint32) []uint32 {
	slices.SortFunc(card, func(a, b uint32) int {
		if Card(a).Color() == Card(b).Color() {
			return int(Card(a).Rank()) - int(Card(b).Rank())
		} else {
			return int(Card(a).Color()) - int(Card(b).Color())
		}
	})
	return card
}

func decision(cards []uint32, showCard uint32) (lose uint32) { // 考虑是否抓牌并丢弃价值最低手牌
	bestChoose := filterChoose(cards)
	odd := rankSum(bestChoose.unqiue)
	fmt.Printf("手上牌型： 组合牌：%v 单牌:%v 无效分:%d \n", showGroup(bestChoose.group), CardList(bestChoose.unqiue).String(), odd)
	hands := make([]uint32, len(cards))
	copy(hands, cards)
	decisionChoose := filterChoose(append(hands, showCard))
	after := rankSum(decisionChoose.unqiue)
	fmt.Printf("吃牌牌型： 组合牌：%v 单牌:%v 无效分:%d \n", showGroup(decisionChoose.group), CardList(decisionChoose.unqiue).String(), after)
	lose = outputDecision(decisionChoose)
	if odd-(after-rankSum([]uint32{lose})) < 5 || lose == showCard { // 抓牌收益不大于5分 不抓
		fmt.Printf("预期打牌：%s 分差:%d 本轮： 抓牌收益不大于5分 不抓牌 \n", Card(lose), odd-(after-rankSum([]uint32{lose})))
		return 0
	}
	fmt.Printf("预期打牌：%s 分差:%d  \n", Card(lose), odd-(after-rankSum([]uint32{lose})))
	return lose
}

func outputDecision(choose RummyGroup) (lose uint32) {
	pairs, ones := explodePair(choose.unqiue, choose.group)
	if len(ones) > 0 {
		lose = ones[len(ones)-1]
		return
	}
	if len(pairs) == 0 {
		return 0
	}
	last := pairs[len(pairs)-1]
	lose = last[len(last)-1]

	return
}

func rankSum(cards []uint32) (sum int) {
	for _, card := range cards {
		if Card(card).IsWild() {
			continue
		}
		sum += int(Card(card).Rank())
	}
	return
}
