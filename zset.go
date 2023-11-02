package jellyzset

import (
	"math/rand"
)

const (
	SkipListMaxLvl  = 32   // Maximum level for the skip list, 2^32 elements
	SkipProbability = 0.25 // Probability for the skip list, 1/4
)

type ZSet struct {
	records map[string]*zset
}

type zset struct {
	records map[string]*zslNode
	zsl     *zskiplist
}

type zskiplist struct {
	head   *zslNode
	tail   *zslNode
	length uint64
	level  int
}

type zslNode struct {
	member    string
	value     interface{}
	score     float64
	backwards *zslNode
	level     []*zslLevel
}

type zslLevel struct {
	forward *zslNode
	span    uint64
}

func createNode(level int, score float64, member string, value interface{}) *zskiplistNode {
	node := &zslNode{
		score:  score,
		member: member,
		value:  value,
		level:  make([]*zslLevel, level),
	}

	for i := range node.level {
		node.level[i] = new(zslLevel)
	}

	return node
}

func newZSkipList() *zskiplist {
	return &zskiplist{
		level: 1,
		head:  createNode(SkipListMaxLvl, 0, "", nil),
	}
}

func getRandomLevel() int {
	level := 1
	for rand.Float64() < SkipProbability && level < SkipListMaxLvl {
		level++
	}

	return level
}

func (z *zskiplist) insert(score float64, member string, value interface{}) *zslNode {
	updates := make([]*zslNode, SkipListMaxLvl)
	rank := make([]uint64, SkipListMaxLvl)

	currentNode := z.head
	for level := z.level - 1; level >= 0; level-- {
		if level == z.level-1 {
			rank[level] = 0
		} else {
			rank[level] = rank[level+1]
		}

		for currentNode.level[level].forward != nil {
			nextNode := currentNode.level[level].forward

			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
				rank[level] += currentNode.level[level].span
			} else {
				break
			}
		}

		updates[level] = currentNode
	}

	level := getRandomLevel()
	if level > z.level {
		for newLevel := z.level; newLevel < level; newLevel++ {
			rank[newLevel] = 0
			updates[newLevel] = z.head
			updates[newLevel].level[newLevel].span = uint64(z.length)
		}

		z.level = level
	}

	newNode := &zslNode{
		score:  score,
		member: member,
		value:  value,
		level:  make([]*zslLevel, level),
	}

	for currentLevel := 0; currentLevel < level; currentLevel++ {
		newNode.level[currentLevel].forward = updates[currentLevel].level[currentLevel].forward
		updates[currentLevel].level[currentLevel].forward = newNode
		newNode.level[currentLevel].span = updates[currentLevel].level[currentLevel].span - (rank[0] - rank[currentLevel])
		updates[currentLevel].level[currentLevel].span = (rank[0] - rank[currentLevel]) + 1
	}

	for currentLevel := level; currentLevel < z.level; currentLevel++ {
		updates[currentLevel].level[currentLevel].span++
	}

	if updates[0] == z.head {
		newNode.backwards = nil
	} else {
		newNode.backwards = updates[0]
	}

	if newNode.level[0].forward != nil {
		newNode.level[0].forward.backwards = newNode
	} else {
		z.tail = newNode
	}

	z.length++
	return newNode
}

func (z *zskiplist) getRank(score float64, member string) int64 {
	var rank uint64 = 0
	currentNode := z.head

	for level := z.level - 1; level >= 0; level-- {
		for currentNode.level[level].forward != nil {
			nextNode := currentNode.level[level].forward
			if nextNode.score < score || (nextNode.score == score && nextNode.member <= member) {
				rank += currentNode.level[level].span
				currentNode = nextNode
			} else {
				break
			}
		}

		if currentNode.member == member {
			return int64(rank)
		}
	}

	return 0
}
