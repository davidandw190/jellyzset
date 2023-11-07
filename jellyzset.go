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
func (z *zskiplist) getRank(score float64, member string) uint64 {
	var rank uint64 = 0
	x := z.head
	for i := z.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil {
			nextNode := x.level[i].forward

			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
				rank += x.level[i].span
				x = nextNode
			} else {
				break
			}
		}

		if x.member == member {
			return rank
		}
	}

	return rank
}

// func (z *zskiplist) getRank(score float64, member string) int64 {
// 	var rank int64
// 	currentNode := z.head
// 	for level := z.level - 1; level >= 0; level-- {
// 		for currentNode.level[level].forward != nil &&
// 			(currentNode.level[level].forward.score < score ||
// 				(currentNode.level[level].forward.score == score &&
// 					currentNode.level[level].forward.member <= member)) {
// 			rank += int64(currentNode.level[level].span)
// 			currentNode = currentNode.level[level].forward
// 		}

// 		if currentNode.member == member {
// 			return rank
// 		}
// 	}

// 	return -1
// }

// func (z *zskiplist) getRank(score float64, member string) int64 {
// 	var rank uint64 = 0
// 	x := z.head
// 	for i := z.level - 1; i >= 0; i-- {
// 		for x.level[i].forward != nil &&
// 			(x.level[i].forward.score < score ||
// 				(x.level[i].forward.score == score &&
// 					x.level[i].forward.member <= member)) {
// 			rank += x.level[i].span
// 			x = x.level[i].forward
// 		}

// 		if x.member == member {
// 			return int64(rank)
// 		}
// 	}

// 	return 0
// }

// func (z *zskiplist) getRank(score float64, member string) int64 {
// 	var rank int64
// 	currentNode := z.head

// 	for level := z.level - 1; level >= 0; level-- {
// 		for currentNode.level[level].forward != nil {
// 			nextNode := currentNode.level[level].forward

// 			if nextNode.score == score && nextNode.member == member {
// 				return rank
// 			}

// 			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
// 				rank++
// 				currentNode = nextNode
// 			} else {
// 				break
// 			}
// 		}
// 	}

// 	return 0
// }

// func (z *zskiplist) getRank(score float64, member string) int64 {
// 	var rank int64 = -1 // Initialize to -1, which will indicate not found

// 	currentNode := z.head

// 	for level := z.level - 1; level >= 0; level-- {
// 		for currentNode.level[level].forward != nil {
// 			nextNode := currentNode.level[level].forward

// 			if nextNode.score == score && nextNode.member == member {
// 				return rank + 1 // Found, return the rank (starting from 0)
// 			}

// 			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
// 				rank++
// 				currentNode = nextNode
// 			} else {
// 				break
// 			}
// 		}
// 	}

// 	return 0
// }

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

// ZAdd adds a member with a specified score to the sorted set stored at the given key.
//
// If the key does not exist, a new sorted set is created and the member is added with the provided score.
// If the member already exists in the sorted set, its score is updated with the new value.
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - score:   The score to assign to the member.
//   - member:  The member to add or update in the sorted set.
//   - value:   The associated value for the member.
//
// Returns:
//   - 1 if the member is added or updated successfully, 0 otherwise.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.2, "member1", "updatedValue1")
//
// In this example, we create a sorted set "mySortedSet" and add two members, "member1" and "member2," with their respective scores and values. The third ZAdd call updates "member1" with a new value and score.
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

// ZScore returns the score of a member in the sorted set stored at the given key.
//
// If the key or member does not exist in the sorted set, it returns (false, 0.0).
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - member:  The member for which the score is requested.
//
// Returns:
//   - A boolean indicating whether the member exists in the sorted set.
//   - The score of the member if it exists; otherwise, 0.0.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	exists, score := zset.ZScore("mySortedSet", "member1")
//
// In this example, we create a sorted set "mySortedSet" and add "member1" with a score of 3.5. We then retrieve the score for "member1," and exists will be true, while the score will be 3.5.
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

// ZCard returns the number of members in the sorted set stored at the given key.
//
// If the key does not exist, it returns 0, indicating an empty sorted set.
//
// Parameters:
//   - key: The key associated with the sorted set.
//
// Returns:
//   - The number of members in the sorted set or 0 if the key does not exist.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	count := zset.ZCard("mySortedSet")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZCard is then used to determine the count, which will be 2.
func (z *ZSet) ZCard(key string) int {
	set, exists := z.records[key]
	if !exists {
		return 0
	}

	return len(set.records)
}

// ZRank returns the rank of a member in the sorted set stored at the given key.
//
// If the key or member does not exist in the sorted set, it returns -1.
// Ranks are 0-based, with 0 being the rank of the member with the lowest score.
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - member:  The member for which the rank is requested.
//
// Returns:
//   - The rank of the member in the sorted set, or -1 if the key or member does not exist.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	rank := zset.ZRank("mySortedSet", "member2")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZRank is used to find the rank of "member2," which will be 0, as it has the lowest score.
func (z *ZSet) ZRank(key, member string) int64 {
	set, exists := z.records[key]
	if !exists {
		return -1
	}

	node, exist := set.records[member]
	if !exist {
		return -1
	}

	rank := int64(set.zsl.getRank(node.score, member)) // Cast the rank to int64
	return rank
}

// ZRevRank returns the reverse rank of a member in the sorted set stored at the given key.
//
// If the key or member does not exist in the sorted set, it returns -1.
// Reverse ranks are 0-based, with 0 being the rank of the member with the highest score.
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - member:  The member for which the reverse rank is requested.
//
// Returns:
//   - The reverse rank of the member in the sorted set, or -1 if the key or member does not exist.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	revRank := zset.ZRevRank("mySortedSet", "member1")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZRevRank is used to find the reverse rank of "member1," which will be 0, as it has the highest score.
func (z *ZSet) ZRevRank(key, member string) int64 {
	set, exists := z.records[key]
	if !exists {
		return -1
	}

	node, exists := set.records[member]
	if !exists {
		return -1
	}

	// Calculate reverse rank by subtracting the rank from the length
	return int64(set.zsl.length - set.zsl.getRank(node.score, member))
}

// ZRem removes a member from the sorted set stored at the given key.
//
// If the key or member does not exist in the sorted set, it returns false.
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - member:  The member to remove from the sorted set.
//
// Returns:
//   - true if the member is successfully removed, false if the key or member does not exist in the sorted set.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	removed := zset.ZRem("mySortedSet", "member1")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZRem is used to remove "member1," and it returns true, indicating successful removal.
func (z *ZSet) ZRem(key, member string) bool {
	set, exists := z.records[key]
	if !exists {
		return false
	}

	if node, exists := set.records[member]; exists {
		set.zsl.delete(node.score, member)
		delete(set.records, member)
		return true
	}

	return true
}
