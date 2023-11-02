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
