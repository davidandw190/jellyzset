package jellyzset

import (
	"math/rand"
)

const (
	SkipListMaxLvl  = 32   // Maximum level for the skip list, 2^32 elements
	SkipProbability = 0.25 // Probability for the skip list, 1/4
)

// ZSet represents a sorted set data structure with multiple records.
type ZSet struct {
	records map[string]*zset
}

// zset represents a single sorted set in the ZSet.
type zset struct {
	records map[string]*zslNode
	zsl     *zskiplist
}

// zskiplist represents the skip list structure for sorted sets.
type zskiplist struct {
	head   *zslNode
	tail   *zslNode
	length uint64
	level  int
}

// zslNode represents a node in the skip list.
type zslNode struct {
	member    string
	value     interface{}
	score     float64
	backwards *zslNode
	level     []*zslLevel
}

// zslLevel represents a level in the skip list node.
type zslLevel struct {
	forward *zslNode
	span    uint64
}

// createNode creates a new skip list node with the given parameters.
func createNode(level int, score float64, member string, value interface{}) *zslNode {
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

// newZSkipList initializes and returns a new empty skip list.
func newZSkipList() *zskiplist {
	return &zskiplist{
		level: 1,
		head:  createNode(SkipListMaxLvl, 0, "", nil),
	}
}

// getRandomLevel returns a random level for a skip list node.
func getRandomLevel() int {
	level := 1
	for rand.Float64() < SkipProbability && level < SkipListMaxLvl {
		level++
	}

	return level
}

// insert adds a new node with the specified score, member, and value to the skip list.
// It returns the inserted node.
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

// getRank returns the rank of a member in the skip list based on its score.
// If the member is not found, it returns 0.
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

// deleteNode deletes a node from the skip list based on the provided node and updates.
func (z *zskiplist) deleteNode(nodeToDelete *zslNode, updates []*zslNode) {
	for level := 0; level < z.level; level++ {
		if updates[level].level[level].forward == nodeToDelete {
			updates[level].level[level].span += nodeToDelete.level[level].span - 1
			updates[level].level[level].forward = nodeToDelete.level[level].forward
		} else {
			updates[level].level[level].span--
		}
	}

	if nodeToDelete.level[0].forward != nil {
		nodeToDelete.level[0].forward.backwards = nodeToDelete.backwards
	} else {
		z.tail = nodeToDelete.backwards
	}

	for z.level > 1 && z.head.level[z.level-1].forward == nil {
		z.level--
	}

	z.length--
}

// delete removes a member with the specified score from the skip list.
func (z *zskiplist) delete(score float64, member string) {
	updates := make([]*zslNode, SkipListMaxLvl)
	currentNode := z.head

	for level := z.level - 1; level >= 0; level-- {
		for currentNode.level[level].forward != nil {
			nextNode := currentNode.level[level].forward
			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
				currentNode = nextNode
			} else {
				break
			}
		}
		updates[level] = currentNode
	}

	currentNode = currentNode.level[0].forward
	if currentNode != nil && currentNode.score == score && currentNode.member == member {
		z.deleteNode(currentNode, updates)
	}
}

// getNodeByRank returns the node in the skip list at the specified rank.
func (z *zskiplist) getNodeByRank(rank uint64) *zslNode {
	if rank == 0 || rank > z.length {
		return nil
	}

	var traversed uint64
	currentNode := z.head

	for level := z.level - 1; level >= 0; level-- {
		for (currentNode.level[level].forward != nil) && (traversed+currentNode.level[level].span <= rank) {
			traversed += currentNode.level[level].span
			currentNode = currentNode.level[level].forward
		}

		if traversed == rank {
			return currentNode
		}
	}

	return nil
}

// findRange retrieves a range of elements from the zset.
// It starts at the 'start' rank and goes up to the 'stop' rank.
// If 'reverse' is true, it fetches the elements in reverse order.
// If 'scoresEnabled' is true, the results will include scores along with members.
// The function returns a slice of interfaces containing the selected elements.
func (z *zset) findRange(key string, start, stop int64, reverse, scoresEnabled bool) []interface{} {

	length := int64(z.zsl.length)
	results := make([]interface{}, 0)

	if start < 0 {
		start += length
		if start < 0 {
			start = 0
		}
	}

	if stop < 0 {
		stop += length
	}

	if start > stop || start >= length {
		return results
	}

	// Calculate the number of elements to fetch
	span := (stop - start) + 1
	node := z.getStartNode(start, reverse)

	// Fetch the elements
	for span > 0 {
		if scoresEnabled {
			results = append(results, node.member, node.score)
		} else {
			results = append(results, node.member)
		}

		span--
		node = z.getNextNode(node, reverse)
	}

	return results

}

// getStartNode retrieves the starting node for a given rank.
// If 'reverse' is true, it adjusts the rank for fetching in reverse order.
func (z *zset) getStartNode(rank int64, reverse bool) *zslNode {
	if reverse {
		rank = int64(z.zsl.length) - rank
	} else {
		rank++
	}

	return z.zsl.getNodeByRank(uint64(rank))
}

// getNextNode retrieves the next node based on the current node in the zset.
// If 'reverse' is true, it returns the previous node (in reverse order).
func (z *zset) getNextNode(currentNode *zslNode, reverse bool) *zslNode {
	if reverse {
		return currentNode.backwards
	}
	return currentNode.level[0].forward
}

// ZAdd adds a member with a specified score to the sorted set stored at key.
func (z *ZSet) ZAdd(key string, score float64, member string, value interface{}) int {
	set, exists := z.records[key]
	if !exists {
		set = &zset{
			records: make(map[string]*zslNode),
			zsl:     newZSkipList(),
		}
		z.records[key] = set
	}

	existingNode, memberExists := set.records[member]

	if memberExists && existingNode.score == score {
		// The member already exists with the same score; update the value.
		existingNode.value = value
	} else {
		// The member is new or has a different score; insert it.
		if memberExists {
			set.zsl.delete(existingNode.score, member)
		}

		newNode := set.zsl.insert(score, member, value)
		set.records[member] = newNode
	}

	return 1
}
