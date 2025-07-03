package test

import (
	"math"
	"testing"
)

func TestCheck(t *testing.T) {
	t.Run("maxFreeTime", func(t *testing.T) {
		ret := maxFreeTime(5, 1, []int{1, 3}, []int{2, 5})
		t.Logf("maxFreeTime: %v", ret)
	})

	t.Run("minSwaps", func(t *testing.T) {
		ret := minSwaps([]int{0, 1, 0, 1, 1, 0, 0})
		t.Logf("minSwaps: %v", ret)
	})

	t.Run("maxFreq", func(t *testing.T) {
		ret := maxFreq("abcabababacabcabc", 3, 3, 10)
		t.Logf("maxFreq: %v", ret)
	})

	t.Run("getSubarrayBeauty", func(t *testing.T) {
		ret := getSubarrayBeauty([]int{1, -1, -3, -2, 3}, 3, 2)
		t.Logf("getSubarrayBeauty: %v", ret)
	})

	t.Run("minFlips", func(t *testing.T) {
		ret := minFlips(`111000`)
		t.Logf("minFlips: %v", ret)
	})

	t.Run("checkInclusion", func(t *testing.T) {
		ret := checkInclusion("ab", "eidbaooo")
		t.Logf("minFlips: %v", ret)
	})

	t.Run("findAnagrams", func(t *testing.T) {
		ret := findAnagrams("baa", "aa")
		t.Logf("minFlips: %v", ret)
	})

	t.Run("findSubstring", func(t *testing.T) {
		ret := findSubstring("wordgoodgoodgoodbestword", []string{"word", "good", "best", "good"})
		t.Logf("minFlips: %v", ret)
	})
}

func findSubstring(s string, words []string) (ans []int) {
	ls, m, n := len(s), len(words), len(words[0])
	for i := 0; i < n && i+m*n <= ls; i++ {
		differ := map[string]int{}
		for j := 0; j < m; j++ {
			differ[s[i+j*n:i+(j+1)*n]]++
		}
		for _, word := range words {
			differ[word]--
			if differ[word] == 0 {
				delete(differ, word)
			}
		}
		for start := i; start < ls-m*n+1; start += n {
			if start != i {
				word := s[start+(m-1)*n : start+m*n]
				differ[word]++
				if differ[word] == 0 {
					delete(differ, word)
				}
				word = s[start-n : start]
				differ[word]--
				if differ[word] == 0 {
					delete(differ, word)
				}
			}
			if len(differ) == 0 {
				ans = append(ans, start)
			}
		}
	}
	return
}

func findAnagrams(s string, p string) (ans []int) {
	n, m := len(p), len(s)
	if n > m {
		return
	}
	var cnt1, cnt2 [26]int
	for i, ch := range p {
		cnt1[ch-'a']++
		cnt2[s[i]-'a']++
	}
	if cnt1 == cnt2 {
		ans = append(ans, 0)
	}

	for i := n; i < m; i++ {
		cnt2[s[i]-'a']++
		cnt2[s[i-n]-'a']--
		if cnt1 == cnt2 {
			ans = append(ans, i-n+1)
		}
	}
	return
}

func checkInclusion(s1, s2 string) bool {
	n, m := len(s1), len(s2)
	if n > m {
		return false
	}
	var cnt1, cnt2 [26]int
	for i, ch := range s1 {
		cnt1[ch-'a']++
		cnt2[s2[i]-'a']++
	}
	if cnt1 == cnt2 {
		return true
	}
	for i := n; i < m; i++ {
		cnt2[s2[i]-'a']++
		cnt2[s2[i-n]-'a']--
		if cnt1 == cnt2 {
			return true
		}
	}
	return false
}

// 环形字符数组 使二进制字符串字符交替的最少反转次数
func minFlips(s string) (ans int) {
	// s 111000
	// loop s => 11100011100
	// window size
	n := len(s)
	ans = n
	cnt := 0
	for i := range 2*n - 1 {
		if int(s[i%n]%2) == i%2 {
			cnt++
		}
		left := i - n + 1
		if left < 0 {
			continue
		}
		ans = min(ans, cnt, n-cnt) //正向和反向
		if int(s[left]%2) == left%2 {
			cnt--
		}
	}

	return
}

// 求美丽值 滑动窗口 计数排序
func getSubarrayBeauty(nums []int, k int, x int) (ans []int) {
	const numWeight = 50
	cnt := [numWeight*2 + 1]int{}

	for i := 0; i < k-1; i++ {
		cnt[numWeight+nums[i]]++ //-50 - -1 => 0-49
	}

	for i, val := range nums[k-1:] {
		cnt[numWeight+val]++
		left := x
		for key, n := range cnt[:numWeight] {
			left -= n
			if left <= 0 {
				ans = append(ans, key-numWeight)
				break
			}
		}

		if i+1 > len(ans) {
			ans = append(ans, 0)
		}

		cnt[numWeight+nums[i]]--
	}

	return
}

func maxFreq(s string, maxLetters int, minSize int, maxSize int) (ret int) {
	cnt := map[string]int{}
	strCnt := map[string]int{}

	for i := 0; i < len(s); i++ {
		cnt[string(s[i])]++
		if i < minSize-1 {
			continue
		}

		if len(cnt) <= maxLetters {
			strCnt[s[i+1-minSize:i+1]]++
			ret = max(ret, strCnt[s[i+1-minSize:i+1]])
		}

		cnt[string(s[i+1-minSize])]--
		if cnt[string(s[i+1-minSize])] == 0 {
			delete(cnt, string(s[i+1-minSize]))
		}
	}

	return
}

func minSwaps(nums []int) (ret int) {
	// nums 按1滑动窗口 环形 算区间最小的0 按1算窗口
	oneLen := 0
	for i := 0; i < len(nums); i++ {
		if nums[i] == 1 {
			oneLen++
		}
	}

	ret = math.MaxInt
	volem := 0 // 1区间最小的0
	for i := 0; i < len(nums)+oneLen; i++ {

		if nums[i%len(nums)] == 0 {
			volem++
		}

		if i < oneLen-1 {
			continue
		}
		ret = min(ret, volem)
		if i+1-oneLen != len(nums) && nums[i+1-oneLen] == 0 {
			volem--
		}
	}

	return
}

func maxFreeTime(eventTime int, k int, startTime []int, endTime []int) (ret int) {
	n := len(startTime)
	freeTime := make([]int, 0, n+1)
	before := 0
	for i := 0; i < n; i++ {
		if i > 0 {
			before = endTime[i-1]
		}
		freeTime = append(freeTime, startTime[i]-before)
	}
	freeTime = append(freeTime, eventTime-endTime[n-1])

	volem := 0
	for i := 0; i < n+1; i++ {
		volem += freeTime[i]
		if i < k {
			continue
		}
		ret = max(ret, volem)
		volem -= freeTime[i-k]
	}
	return
}
