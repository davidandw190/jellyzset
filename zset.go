package jellyzset

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
	lvl    int
}

type zslNode struct {
	member    string
	value     interface{}
	score     float64
	backwards *zslNode
	lvl       []*zslLevel
}

type zslLevel struct {
	forward *zslNode
	span    uint64
}

func createNode(level int, score float64, member string, value interface{}) *zslNode {
	node := &zslNode{
		score:  score,
		member: member,
		value:  value,
		lvl:    make([]*zslLevel, level),
	}

	for i := range node.lvl {
		node.lvl[i] = new(zslLevel)
	}

	return node
}
