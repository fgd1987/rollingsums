package main

import (
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

type Block struct {
	pos int
}

type Patch struct {
	data string
	pos  int
	len  int
}

var step = 5

func min(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func GenerateFile(origin string, patchList *list.List) string {
	result := ""
	for e := patchList.Front(); e != nil; e = e.Next() {
		patch := e.Value.(*Patch)
		if patch.pos == -1 {
			result += patch.data
			//println(patch.data)
		} else {
			//end := min(patch.pos+patch.len, len(origin))
			//fmt.Printf("%d:%d    %d\n", patch.pos, patch.len, len(origin))
			result += origin[patch.pos : patch.pos+patch.len]
		}
	}
	return result
}

func MakePatch(f2 string, blockMap map[string][]int) *list.List {
	dataLen := len(f2)
	patchList := list.New()
	var backItem *Patch

	var bufA = -1
	var bufB = -1
	bflag := false

	i := 0
	for i = 0; i+step <= dataLen; {
		backItem = nil
		if patchList.Back() != nil {
			backItem = patchList.Back().Value.(*Patch)
		}
		// if i+step > dataLen { //不足一个block的剩余
		// 	// if backItem == nil || backItem.pos > -1 {
		// 	// 	backItem = &Patch{pos: -1}
		// 	// 	patchList.PushBack(backItem)
		// 	// }

		// 	// backItem.data += f2[i:len(f2)] //todo  优化
		// 	break
		// }
		h := hash(f2[i : i+step])
		if v, ok := blockMap[h]; ok {
			bflag = true
			if bufA != -1 {
				backItem.data += f2[bufA:bufB]
				bufA = -1
				bufB = -1
			}
			//println(f2[i : i+step])
			//fmt.Printf("find: %s   %d   %s\n", h, v[0], f2[i:i+step])

			//优化 队列上一个元素不是字符串 或  间断
			if backItem == nil || backItem.pos == -1 || v[0] != backItem.pos+backItem.len {
				backItem = &Patch{
					pos: v[0],
					len: step,
				}
				patchList.PushBack(backItem)
			} else {
				backItem.len += step
			}

			i += step

		} else {
			bflag = false
			if backItem == nil || backItem.pos > -1 {
				backItem = &Patch{
					pos: -1,
				}
				patchList.PushBack(backItem)
			}
			//backItem.data += f2[i : i+1] //todo  优化

			if bufA == -1 {
				bufA = i
				bufB = i
			}
			bufB++

			i++
		}

		//println(i)
	}

	if !bflag && bufA > -1 {
		backItem.data += f2[bufA:bufB]
	}

	if i+step > dataLen { //不足一个block的剩余
		if backItem == nil || backItem.pos > -1 {
			backItem = &Patch{pos: -1}
			patchList.PushBack(backItem)
		}

		backItem.data += f2[i:len(f2)] //todo  优化
	}

	return patchList
}

func Diff(f1 string, f2 string) bool {
	//step1  run in server
	blockMap := make(map[string][]int)
	for i := 0; i <= len(f1)-step; i += step {
		h := hash(f1[i : i+step])
		blockMap[h] = append(blockMap[h], i)
	}

	//step2 run in client
	patchList := MakePatch(f2, blockMap)

	//step3 run in server
	result := GenerateFile(f1, patchList)
	println("result: " + result)
	return result == f2
}

var letterRunes = []rune("abc")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// f1 := "aaaaabbbbbcccccdddddaaaaae"
	// f2 := "aeaddbbbbbbdddddaaaa"

	t1 := time.Now() // get current time
	for i := 0; i < 100000; i++ {
		f1 := RandString(rand.Intn(100))
		f2 := RandString(rand.Intn(100))

		//f1 = "bbcbba"
		//f2 = "cbbbabbbbcbbaccababbbacbcccaaacabbbaabccbcaaaaabbbcbcbbaccaaabcabbbcbbbbcbbbaabc"

		//println(f1 + " <-----> " + f2)

		if !Diff(f1, f2) {
			panic(f1 + "  " + f2)
		}

		if i%1000 == 0 {
			println(i)
		}
	}
	elapsed := time.Since(t1)
	fmt.Println("App elapsed: ", elapsed)

}

func hash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	//return hex.EncodeToString(h.Sum(nil))[8:24]
	return hex.EncodeToString(h.Sum(nil))
}