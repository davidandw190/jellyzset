package jellyzset

// Package jellyzset provides a high-performance implementation of a Redis-compatible ZSet (sorted set)
// data structure in Go. This library is designed to efficiently manage ordered collections of elements
// associated with scores, offering logarithmic-time complexity for key operations.
//
// Overview:
//
//	The jellyzset package employs a skip list-based indexing structure to optimize the insertion,
//	deletion, and range query operations inherent to sorted sets. Each ZSet instance manages multiple
//	sorted sets, uniquely identified by keys. The underlying skip list, represented by the zskiplist
//	struct, serves as the backbone for quick and effective access to elements within a sorted set.
//
// Implementation Details:
//   - The skip list dynamically balances itself, ensuring rapid insertion and deletion operations.
//   - Element order is maintained based on scores, facilitating efficient range queries.
//   - Skip list levels expedite searches, resulting in logarithmic time complexity for key operations.
//
// Credits/Resources:
//
//   - https://www.epaperpress.com/sortsearch/download/skiplist.pdf
//
//   - http://blog.wjin.org/posts/redis-internal-data-structure-skiplist.html
//
//   - https://developpaper.com/skip-list-lookup-tree-btree-in-redis-mysql/
//
//   - https://github.com/redis/redis/blob/unstable/src/server.h
//
//   - https://github.com/arriqaaq/zset
//
//   - https://github.com/redis/redis/blob/unstable/src/server.h
//
//   - https://www.youtube.com/watch?v=UGaOXaXAM5M
//
//   - https://www.youtube.com/watch?v=NDGpsfwAaqo

import (
	"errors"
	"math"
	"math/rand"
)

const (
	SkipListMaxLvl  = 32   // Maximum level for the skip list, 2^32 elements
	SkipProbability = 0.25 // Probability for the skip list, 1/4
)

// ZSet represents a collection of sorted sets, each identified by a unique key.
// It uses a map to store references to individual sorted sets.
type ZSet struct {
	records map[string]*zset
}

// ZRangeConfig specifies the configuration for ZRangeByScore method to customize the range query.
type ZRangeConfig struct {
	Limit        int  // Limit the max nodes to be returned
	ExcludeStart bool // Exclude start value, so it searches in the interval (start, end] or (start, end)
	ExcludeEnd   bool // Exclude end value, so it searches in the interval [start, end) or (start, end)
}

// zset represents an individual sorted set in the ZSet data structure.
// It contains references to the skip list and a map of elements.
type zset struct {
	records map[string]*zslNode
	zsl     *zskiplist
}

// zskiplist is a skip list-based data structure used to maintain order in the sorted set.
type zskiplist struct {
	head   *zslNode
	tail   *zslNode
	length uint64
	level  int
}

// zslNode represents a node in the skip list, containing information about the element,
// its score, and references to the next nodes in different levels.
type zslNode struct {
	member    string
	value     interface{}
	score     float64
	backwards *zslNode
	level     []*zslLevel
}

// zslLevel represents a level in the skip list, containing references to the forward node and
// the span, which is the number of elements between the current node and the next node in that level.
type zslLevel struct {
	forward *zslNode
	span    uint64
}

// New creates a new instance of the ZSet data structure.
func New() *ZSet {
	return &ZSet{
		make(map[string]*zset),
	}
}

// createNode creates a new zslNode with the given parameters.
// It initializes the levels based on the specified level.
func createNode(level int, score float64, member string, value interface{}) *zslNode {
	newNode := &zslNode{
		score:  score,
		member: member,
		value:  value,
		level:  make([]*zslLevel, level),
	}

	for i := range newNode.level {
		newNode.level[i] = new(zslLevel)
	}

	return newNode
}

// newZSkipList creates a new instance of the zskiplist with an initial head node.
func newZSkipList() *zskiplist {
	head := createNode(SkipListMaxLvl, 0, "", nil)
	return &zskiplist{
		level: 1,
		head:  head,
		tail:  head,
	}
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
	return int64(set.zsl.length - set.zsl.getRank(node.score, member) - 1)
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

	return false
}

// ZScoreRange retrieves a range of elements with scores within the specified range from the sorted set stored at the given key.
//
// If the key does not exist or the provided minimum score is greater than the maximum score, it returns nil.
//
// Parameters:
//   - key:  The key associated with the sorted set.
//   - min:  The minimum score of the range (inclusive).
//   - max:  The maximum score of the range (inclusive).
//
// Returns:
//   - A slice of interfaces containing elements with scores within the specified range.
//   - The slice is empty if there are no elements within the range or if the key does not exist.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.2, "member3", "value3")
//	results := zset.ZScoreRange("mySortedSet", 2.5, 4.0)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZScoreRange is then used to retrieve elements within the score range of 2.5 to 4.0, and the results slice will contain the elements "member1" and "member2" with their respective scores.
func (z *ZSet) ZScoreRange(key string, min, max float64) []interface{} {
	if _, exists := z.records[key]; !exists || min > max {
		return nil
	}

	item := z.records[key].zsl
	minScore, maxScore := z.limitScores(item, min, max)

	return z.collectElementsInRange(item, minScore, maxScore)
}

// ZRevScoreRange returns all the elements in the sorted set at the given key with scores falling within the range [max, min].
//
// This function returns elements ordered from high to low scores within the specified range, including elements with scores equal to max or min.
//
// If the key does not exist or if the provided max score is less than the min score, the function returns an empty slice.
//
// Parameters:
//   - key: The key associated with the sorted set.
//   - max: The maximum score for the range.
//   - min: The minimum score for the range.
//
// Returns:
//   - A slice of interfaces containing elements with scores within the specified range, ordered from high to low scores.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.0, "member3", "value3")
//	result := zset.ZRevScoreRange("mySortedSet", 4.0, 2.0)
//
// In this example, we create a sorted set "mySortedSet" and add three members with different scores. ZRevScoreRange is used to retrieve elements within the score range [4.0, 2.0]. The result will be a slice containing the elements "member3" with a score of 4.0 and "member2" with a score of 2.0, ordered from high to low scores.
func (z *ZSet) ZRevScoreRange(key string, max, min float64) []interface{} {
	if _, exists := z.records[key]; !exists || min > max {
		return nil
	}

	item := z.records[key].zsl
	minScore, maxScore := z.limitScores(item, min, max)

	return z.collectElementsInReverseRange(item, maxScore, minScore)
}

// ZKeyExists checks if a sorted set exists with the given key.
//
// Parameters:
//   - key: The key to check for the existence of a sorted set.
//
// Returns:
//   - true if a sorted set exists with the provided key, false otherwise.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	exists := zset.ZKeyExists("mySortedSet")
//
// In this example, we create a sorted set "mySortedSet" and use ZKeyExists to check if it exists. The result will be true.
func (z *ZSet) ZKeyExists(key string) bool {
	_, exists := z.records[key]
	return exists
}

// ZClear removes a sorted set with the given key from the ZSet.
//
// If the sorted set with the provided key does not exist, the function has no effect.
//
// Parameters:
//   - key: The key associated with the sorted set to be removed.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZClear("mySortedSet")
//
// In this example, we create a sorted set "mySortedSet" and then use ZClear to remove it. After this operation, ZKeyExists("mySortedSet") will return false.
func (z *ZSet) ZClear(key string) {
	if z.ZKeyExists(key) {
		delete(z.records, key)
	}
}

// ZKeys returns a slice of all the keys in the ZSet, representing individual sorted sets.
//
// This function provides a list of all the unique keys present in the ZSet.
//
// Returns:
//   - A slice of strings containing all the keys in the ZSet.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("set1", 3.5, "member1", "value1")
//	zset.ZAdd("set2", 2.0, "member2", "value2")
//	keys := zset.ZKeys()
//
// In this example, we create a ZSet and add two sorted sets with keys "set1" and "set2." The Keys function is used to retrieve a slice containing the keys ["set1", "set2"].
func (z *ZSet) ZKeys() []string {
	keys := make([]string, 0, len(z.records))
	for key := range z.records {
		keys = append(keys, key)
	}
	return keys
}

// ZRange returns a range of elements from the sorted set at the given key.
//
// It starts at the 'start' index and goes up to the 'stop' index (inclusive).
// If 'start' is greater than 'stop' or the key does not exist, an empty slice is returned.
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - start: The starting index of the range.
//   - stop:  The ending index of the range.
//
// Returns:
//   - A slice of interfaces containing the selected elements within the specified range.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.0, "member3", "value3")
//	result := zset.ZRange("mySortedSet", 0, 1)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZRange is used to retrieve elements within the range [0, 1]. The result will be a slice containing the elements "member2" and "member1".
func (z *ZSet) ZRange(key string, start, stop int) []interface{} {
	if !z.ZKeyExists(key) {
		return []interface{}{}
	}

	return z.records[key].findRange(key, int64(start), int64(stop), false, false)
}

// ZRangeWithScore returns a range of elements with scores from the sorted set at the given key.
//
// It starts at the 'start' index and goes up to the 'stop' index (inclusive).
// If 'start' is greater than 'stop' or the key does not exist, an empty slice is returned.
// The results include scores along with members in the format [member1, score1, member2, score2, ...].
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - start: The starting index of the range.
//   - stop:  The ending index of the range.
//
// Returns:
//   - A slice of interfaces containing the selected elements with scores within the specified range.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.0, "member3", "value3")
//	result := zset.ZRangeWithScore("mySortedSet", 0, 1)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZRangeWithScores is used to retrieve elements with scores within the range [0, 1]. The result will be a slice containing the elements "member2," its score 2.0, "member1," and its score 3.5.
func (z *ZSet) ZRangeWithScore(key string, start, stop int) []interface{} {
	if !z.ZKeyExists(key) || start > stop {
		return []interface{}{}
	}

	return z.records[key].findRange(key, int64(start), int64(stop), false, true)
}

// ZRevRange returns a range of elements in reverse order from the sorted set at the given key.
//
// It starts at the 'start' index and goes down to the 'stop' index (inclusive).
// If 'start' is greater than 'stop' or the key does not exist, an empty slice is returned.
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - start: The starting index of the range.
//   - stop:  The ending index of the range.
//
// Returns:
//   - A slice of interfaces containing the selected elements within the specified range, in reverse order.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.0, "member3", "value3")
//	result := zset.ZRevRange("mySortedSet", 1, 0)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZRevRange is used to retrieve elements in reverse order within the range [1, 0]. The result will be a slice containing the elements "member1" and "member2" in reverse order.
func (z *ZSet) ZRevRange(key string, start, stop int) []interface{} {
	if !z.ZKeyExists(key) || start > stop {
		return []interface{}{}
	}

	return z.records[key].findRange(key, int64(start), int64(stop), true, false)
}

// ZRevRangeWithScore returns a range of elements with scores in reverse order from the sorted set at the given key.
//
// It starts at the 'start' index and goes down to the 'stop' index (inclusive).
// If 'start' is greater than 'stop' or the key does not exist, nil is returned.
// The results include scores along with members in the format [member1, score1, member2, score2, ...], in reverse order.
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - start: The starting index of the range.
//   - stop:  The ending index of the range.
//
// Returns:
//   - A slice of interfaces containing the selected elements with scores within the specified range, in reverse order.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.0, "member3", "value3")
//	result := zset.ZRevRangeWithScore("mySortedSet", 1, 0)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZRevRangeWithScores is used to retrieve elements with scores in reverse order within the range [1, 0]. The result will be a slice containing the elements "member2" with its score 2.0 and "member1" with its score 3.5, in reverse order.
func (z *ZSet) ZRevRangeWithScore(key string, start, stop int) []interface{} {
	if !z.ZKeyExists(key) || start > stop {
		return nil
	}

	return z.records[key].findRange(key, int64(start), int64(stop), true, true)
}

// ZRetrieveByRank retrieves the member and score at the specified rank from the sorted set stored at the given key.
//
// If the key does not exist or the provided rank is out of bounds, it returns an empty slice.
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - rank:  The rank of the member to retrieve (0-based).
//
// Returns:
//   - A slice of interfaces containing the member and score at the specified rank.
//   - The slice is empty if the key does not exist or if the rank is out of bounds.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	result := zset.ZRetrieveByRank("mySortedSet", 0)
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZRetrieveByRank is then used to retrieve the member and score at rank 0, resulting in the slice ["member2", 2.0].
func (z *ZSet) ZRetrieveByRank(key string, rank int) []interface{} {
	zset, exists := z.records[key]
	if !exists {
		return []interface{}{}
	}

	member, score := zset.getNodeByRank(key, int64(rank), false)
	return []interface{}{member, score}
}

// ZRevRetrieveByRank retrieves the member and score at the specified reverse rank from the sorted set stored at the given key.
//
// If the key does not exist or the provided rank is out of bounds, it returns an empty slice.
//
// Parameters:
//   - key:   The key associated with the sorted set.
//   - rank:  The reverse rank of the member to retrieve (0-based).
//
// Returns:
//   - A slice of interfaces containing the member and score at the specified reverse rank.
//   - The slice is empty if the key does not exist or if the rank is out of bounds.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	result := zset.ZRevRetrieveByRank("mySortedSet", 0)
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZRevRetrieveByRank is then used to retrieve the member and score at reverse rank 0, resulting in the slice ["member1", 3.5].
func (z *ZSet) ZRevRetrieveByRank(key string, rank int) []interface{} {
	zset, exists := z.records[key]
	if !exists {
		return []interface{}{}
	}

	member, score := zset.getNodeByRank(key, int64(rank), true)

	return []interface{}{member, score}
}

// ZPopMin retrieves and removes the member with the lowest score from the sorted set stored at the given key.
//
// If the key does not exist or the sorted set is empty, it returns (nil, errors.New("key does not exist")).
//
// Parameters:
//   - key: The key associated with the sorted set.
//
// Returns:
//   - A pointer to the zslNode containing information about the popped member and score.
//   - An error if the key does not exist or the sorted set is empty.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	poppedNode, err := zset.ZPopMin("mySortedSet")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZPopMin is then used to retrieve and remove the member with the lowest score, resulting in the poppedNode containing information about "member2" and its score of 2.0.
func (z *ZSet) ZPopMin(key string) (*zslNode, error) {
	if !z.ZKeyExists(key) {
		return nil, errors.New("key does not exist")
	}

	zset := z.records[key]
	firstNode := zset.zsl.head.level[0].forward

	if firstNode != nil {
		z.ZRem(key, firstNode.member)
	}

	return firstNode, nil
}

// ZPopMax retrieves and removes the member with the highest score from the sorted set stored at the given key.
//
// If the key does not exist or the sorted set is empty, it returns (nil, errors.New("key does not exist")).
//
// Parameters:
//   - key: The key associated with the sorted set.
//
// Returns:
//   - A pointer to the zslNode containing information about the popped member and score.
//   - An error if the key does not exist or the sorted set is empty.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	poppedNode, err := zset.ZPopMax("mySortedSet")
//
// In this example, we create a sorted set "mySortedSet" and add two members. ZPopMax is then used to retrieve and remove the member with the highest score, resulting in the poppedNode containing information about "member1" and its score of 3.5.
func (z *ZSet) ZPopMax(key string) (*zslNode, error) {
	if !z.ZKeyExists(key) {
		return nil, errors.New("key does not exist")
	}

	zset := z.records[key]
	lastNode := zset.zsl.tail

	if lastNode != nil {
		z.ZRem(key, lastNode.member)
	}

	return lastNode, nil
}

// ZRangeByScore retrieves elements with scores within the specified range from the sorted set stored at the given key.
//
// If the key does not exist, it returns an empty slice.
//
// Parameters:
//   - key:     The key associated with the sorted set.
//   - start:   The minimum score of the range (inclusive).
//   - end:     The maximum score of the range (inclusive).
//   - config:  Configuration options for the range query (optional).
//
// Returns:
//   - A slice of pointers to zslNode containing information about elements within the specified score range.
//   - The slice is empty if the key does not exist or if no elements are within the specified range.
//
// Example:
//
//	zset := jellyzset.New()
//	zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
//	zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
//	zset.ZAdd("mySortedSet", 4.2, "member3", "value3")
//	config := &ZRangeConfig{Limit: 2, ExcludeEnd: true}
//	result := zset.ZRangeByScore("mySortedSet", 2.0, 4.0, config)
//
// In this example, we create a sorted set "mySortedSet" and add three members. ZRangeByScore is then used to retrieve elements within the score range of 2.0 to 4.0, excluding the end, and the result will contain pointers to zslNode for "member2" and "member1".
func (z *ZSet) ZRangeByScore(key string, start, end float64, config *ZRangeConfig) []*zslNode {
	result := []*zslNode{}
	if z.ZKeyExists(key) {
		return result
	}

	node := z.records[key]
	zsl := node.zsl

	limit := int(^uint(0) >> 1)
	if config != nil && config.Limit > 0 {
		limit = config.Limit
	}

	excludeStart, excludeEnd := false, false
	if config != nil {
		excludeStart, excludeEnd = config.ExcludeStart, config.ExcludeEnd
	}

	reverse := start > end
	if reverse {
		start, end = end, start
		excludeStart, excludeEnd = excludeEnd, excludeStart
	}

	// Determine if out of range
	if zsl.length == 0 {
		return result
	}

	compare := func(a, b float64) bool {
		if reverse {
			return a > b
		}
		return a < b
	}

	updateNode := func(n *zslNode, level int) *zslNode {
		for n.level[level].forward != nil && compare(n.level[level].forward.score, end) {
			n = n.level[level].forward
		}
		return n
	}

	if reverse {
		// Search from end to start
		currentNode := zsl.head
		if excludeEnd {
			for level := zsl.level - 1; level >= 0; level-- {
				currentNode = updateNode(currentNode, level)
			}
		} else {
			for level := zsl.level - 1; level >= 0; level-- {
				currentNode = updateNode(currentNode, level)
			}
		}

		for currentNode != nil && limit > 0 {
			if excludeStart && !compare(currentNode.score, start) {
				break
			}

			nextNode := currentNode.backwards

			result = append(result, currentNode)
			limit--

			currentNode = nextNode
		}
	} else {
		// Search from start to end
		currentNode := zsl.head
		if excludeStart {
			for level := zsl.level - 1; level >= 0; level-- {
				currentNode = updateNode(currentNode, level)
			}
		} else {
			for level := zsl.level - 1; level >= 0; level-- {
				currentNode = updateNode(currentNode, level)
			}
		}

		currentNode = currentNode.level[0].forward

		for currentNode != nil && limit > 0 {
			if excludeEnd && !compare(currentNode.score, end) {
				break
			}

			nextNode := currentNode.level[0].forward

			result = append(result, currentNode)
			limit--

			currentNode = nextNode
		}
	}

	return result
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

	for level := z.level - 1; level >= 0; level-- {
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

	newNodeLevel := getRandomLevel()

	if newNodeLevel > z.level {
		for i := z.level; i < newNodeLevel; i++ {
			rankValues[i] = 0
			updateNodes[i] = z.head
			updateNodes[i].level[i].span = uint64(z.length)
		}
		z.level = newNodeLevel
	}

	newNode := createNode(newNodeLevel, score, member, value)

	for level := 0; level < newNodeLevel; level++ {
		newNode.level[level].forward = updateNodes[level].level[level].forward
		updateNodes[level].level[level].forward = newNode

		newNode.level[level].span = updateNodes[level].level[level].span - (rankValues[0] - rankValues[level])
		updateNodes[level].level[level].span = (rankValues[0] - rankValues[level]) + 1
	}

	for level := newNodeLevel; level < z.level; level++ {
		updateNodes[level].level[level].span++
	}

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
	currentNode := z.head
	for level := z.level - 1; level >= 0; level-- {
		for currentNode.level[level].forward != nil {
			nextNode := currentNode.level[level].forward

			if nextNode.score < score || (nextNode.score == score && nextNode.member < member) {
				rank += currentNode.level[level].span
				currentNode = nextNode
			} else {
				break
			}
		}

		if currentNode.member == member {
			return rank
		}
	}

	return rank
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

func (z *zset) getNodeByRank(key string, rank int64, reverse bool) (string, float64) {
	if rank < 0 || rank > int64(z.zsl.length) {
		return "", math.MinInt64
	}

	if reverse {
		rank = int64(z.zsl.length) - rank
	} else {
		rank++
	}

	n := z.zsl.getNodeByRank(uint64(rank))
	if n == nil {
		return "", math.MinInt64
	}

	node := z.records[n.member]
	if node == nil {
		return "", math.MinInt64
	}

	return node.member, node.score

}

// getNodeByRank returns the node in the skip list at the specified rank.
func (zsl *zskiplist) getNodeByRank(rank uint64) *zslNode {
	if rank == 0 || rank > zsl.length {
		return nil
	}

	var traversed uint64
	currentNode := zsl.head

	for level := zsl.level - 1; level >= 0; level-- {
		for currentNodeHasForward(currentNode, level) && traversed+currentNode.level[level].span <= rank {
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
// If 'reverseEnabled' is true, it fetches the elements in reverse order.
// If 'scoresEnabled' is true, the results will include scores along with members.
// The function returns a slice of interfaces containing the selected elements.
func (zset *zset) findRange(key string, start, stop int64, reverse, withScores bool) (result []interface{}) {
	length := int64(zset.zsl.length)

	start = adjustRange(start, length)
	stop = adjustRange(stop, length)

	if start > stop {
		return
	}

	span := stop - start + 1
	node := zset.getStartNode(start, reverse)

	for span > 0 && node != nil {
		span--
		if withScores {
			result = append(result, node.member, node.score)
		} else {
			result = append(result, node.member)
		}
		node = zset.getNextNode(node, reverse)
	}

	return result
}

// Helper function to adjust range values
func adjustRange(value, length int64) int64 {
	if value < 0 {
		value += length
		if value < 0 {
			value = 0
		}
	}
	return value
}

// Helper function to check if the current node has a forward node at a given level
func currentNodeHasForward(node *zslNode, level int) bool {
	return node.level[level].forward != nil
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

// limitScores ensures that min and max scores fall within the valid score range.
//
// If min is below the lowest score, it is set to the lowest score.
// If max is above the highest score, it is set to the highest score.
func (z *ZSet) limitScores(item *zskiplist, min, max float64) (float64, float64) {
	minScore := item.head.level[0].forward.score
	if min < minScore {
		min = minScore
	}

	maxScore := item.tail.score
	if max > maxScore {
		max = maxScore
	}

	return min, max
}

// collectElementsInRange collects all elements with scores between min and max in the sorted set.
func (z *ZSet) collectElementsInRange(item *zskiplist, min, max float64) []interface{} {
	var result []interface{}
	currentNode := item.head
	for level := item.level - 1; level >= 0; level-- {
		for currentNode.level[level].forward != nil && currentNode.level[level].forward.score < min {
			currentNode = currentNode.level[level].forward
		}
	}

	currentNode = currentNode.level[0].forward
	for currentNode != nil && currentNode.score <= max {
		result = append(result, currentNode.member, currentNode.score)
		currentNode = currentNode.level[0].forward
	}

	return result
}

// collectElementsInReverseRange collects all elements with scores between max and min in the sorted set, in reverse order.
func (z *ZSet) collectElementsInReverseRange(item *zskiplist, max, min float64) []interface{} {
	var result []interface{}
	currentNode := item.head
	for level := item.level - 1; level >= 0; level-- {
		for currentNode.level[level].forward != nil && currentNode.level[level].forward.score <= max {
			currentNode = currentNode.level[level].forward
		}
	}

	for currentNode != nil && currentNode.score >= min {
		result = append(result, currentNode.member, currentNode.score)
		currentNode = currentNode.backwards
	}

	return result
}
