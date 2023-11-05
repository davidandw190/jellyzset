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
	head := createNode(SkipListMaxLvl, 0, "", nil)
	return &zskiplist{
		level: 1,
		head:  head,
		tail:  head,
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
	// Initialize arrays for update nodes and rank values
	updateNodes := make([]*zslNode, SkipListMaxLvl)
	rankValues := make([]uint64, SkipListMaxLvl)

	currentNode := z.head

	// Traverse the levels of the skip list
	for level := z.level - 1; level >= 0; level-- {
		// Initialize rank and update node information
		if level == z.level-1 {
			rankValues[level] = 0
		} else {
			rankValues[level] = rankValues[level+1]
		}

		for currentNode.level[level].forward != nil &&
			(currentNode.level[level].forward.score < score ||
				(currentNode.level[level].forward.score == score && currentNode.level[level].forward.member < member)) {

			rankValues[level] += currentNode.level[level].span
			currentNode = currentNode.level[level].forward
		}

		updateNodes[level] = currentNode
	}

	// Generate a random level for the new node
	newNodeLevel := getRandomLevel()

	// Add a new level if needed
	if newNodeLevel > z.level {
		for i := z.level; i < newNodeLevel; i++ {
			rankValues[i] = 0
			updateNodes[i] = z.head
			updateNodes[i].level[i].span = uint64(z.length)
		}
		z.level = newNodeLevel
	}

	// Create a new node
	newNode := createNode(newNodeLevel, score, member, value)

	// Insert the new node according to the update nodes and rank values
	for level := 0; level < newNodeLevel; level++ {
		newNode.level[level].forward = updateNodes[level].level[level].forward
		updateNodes[level].level[level].forward = newNode

		newNode.level[level].span = updateNodes[level].level[level].span - (rankValues[0] - rankValues[level])
		updateNodes[level].level[level].span = (rankValues[0] - rankValues[level]) + 1
	}

	// Increment span for untouched levels
	for level := newNodeLevel; level < z.level; level++ {
		updateNodes[level].level[level].span++
	}

	// Update the backward and forward pointers
	if updateNodes[0] == z.head {
		newNode.backwards = nil
	} else {
		newNode.backwards = updateNodes[0]
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

func New() *ZSet {
	return &ZSet{
		make(map[string]*zset),
	}
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

// ZScore returns the score of a member in the sorted set at key.
func (z *ZSet) ZScore(key string, member string) (ok bool, score float64) {
	set, exists := z.records[key]
	if !exists {
		return false, 0.0
	}

	node, exists := set.records[member]
	if !exists {
		return false, 0.0
	}

	return true, node.score
}

// ZCard returns the number of members in the sorted set at key.
func (z *ZSet) ZCard(key string) int {
	set, exists := z.records[key]
	if !exists {
		return 0
	}

	return len(set.records)
}

// ZRank returns the rank of a member in the sorted set at key.
func (z *ZSet) ZRank(key, member string) int64 {
	set, exists := z.records[key]
	if !exists {
		return -1
	}

	node, exists := set.records[member]
	if !exists {
		return -1
	}

	return int64(set.zsl.getRank(node.score, member))
}

// ZRevRank returns the reverse rank of a member in the sorted set at key.
func (z *ZSet) ZRevRank(key, member string) int64 {
	set, exists := z.records[key]
	if !exists {
		return -1
	}

	node, exists := set.records[member]
	if !exists {
		return -1
	}

	return int64(set.zsl.length) - int64(set.zsl.getRank(node.score, member))
}
