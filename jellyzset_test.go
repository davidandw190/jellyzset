package jellyzset

import (
	"reflect"
	"testing"
)

func TestZSet_FindRange(t *testing.T) {
	zset := New()

	t.Run("FindRange Non-Existent Key", func(t *testing.T) {
		// Test finding a range for a non-existent key.
		nonExistentKey := "nonexistent_key"
		if z, exists := zset.records[nonExistentKey]; exists {
			result := z.findRange(nonExistentKey, 0, 1, false, false)
			expected := []interface{}{}
			if len(expected) != len(result) {
				t.Errorf("Expected %v but got %v", expected, result)
			}
		}
	})

	t.Run("FindRange Existing Key", func(t *testing.T) {
		// Test finding a range for an existing key.
		zset.ZAdd("existing_key", 1.0, "member1", nil)
		zset.ZAdd("existing_key", 2.0, "member2", nil)
		zset.ZAdd("existing_key", 3.0, "member3", nil)

		result := zset.records["existing_key"].findRange("existing_key", 0, 2, false, false)
		expected := []interface{}{"member1", "member2", "member3"}
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("FindRange Reverse Order", func(t *testing.T) {
		// Test finding a range in reverse order.
		zset.ZAdd("reverse_key", 1.0, "member1", nil)
		zset.ZAdd("reverse_key", 2.0, "member2", nil)
		zset.ZAdd("reverse_key", 3.0, "member3", nil)

		result := zset.records["reverse_key"].findRange("reverse_key", 0, 2, true, false)
		expected := []interface{}{"member3", "member2", "member1"}
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("FindRange With Scores", func(t *testing.T) {
		// Test finding a range with scores.
		zset.ZAdd("score_key", 1.0, "member1", nil)
		zset.ZAdd("score_key", 2.0, "member2", nil)
		zset.ZAdd("score_key", 3.0, "member3", nil)

		result := zset.records["score_key"].findRange("score_key", 0, 2, false, true)
		expected := []interface{}{"member1", 1.0, "member2", 2.0, "member3", 3.0}
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("FindRange With Exclusions", func(t *testing.T) {
		// Test finding a range with exclusions.
		zset.ZAdd("exclusion_key", 1.0, "member1", nil)
		zset.ZAdd("exclusion_key", 2.0, "member2", nil)
		zset.ZAdd("exclusion_key", 3.0, "member3", nil)

		result1 := zset.records["exclusion_key"].findRange("exclusion_key", 0, 2, false, false)
		expected1 := []interface{}{"member1", "member2", "member3"}
		if !reflect.DeepEqual(expected1, result1) {
			t.Errorf("Expected %v but got %v", expected1, result1)
		}

		result2 := zset.records["exclusion_key"].findRange("exclusion_key", 0, 2, false, false)
		expected2 := []interface{}{"member1", "member2", "member3"}
		if !reflect.DeepEqual(expected2, result2) {
			t.Errorf("Expected %v but got %v", expected2, result2)
		}
	})
}

func TestZSet_ZAdd(t *testing.T) {
	zset := New()

	t.Run("Add Single Member to Empty ZSet", func(t *testing.T) {
		// Test adding a single member to an empty sorted set.
		key := "sorted_set"
		score := 3.5
		member := "member1"
		value := "value1"
		count := zset.ZAdd(key, score, member, value)

		assertCountEqual(t, 1, count, "Add a Single Member")

		ok, addedScore := zset.ZScore(key, member)
		assertBoolEqual(t, true, ok, "Add a Single Member - Member Existence Check")
		assertFloatEqual(t, score, addedScore, "Add a Single Member - Score Check")
	})

	t.Run("Add Multimple Members to ZSet", func(t *testing.T) {
		// Test adding multiple members to an existing sorted set.
		key := "sorted_set"
		count := zset.ZAdd(key, 4.0, "member2", "value2")
		count += zset.ZAdd(key, 2.0, "member3", "value3")

		assertCountEqual(t, 2, count, "Add Multiple Members")

		ok, score1 := zset.ZScore(key, "member2")
		assertBoolEqual(t, true, ok, "Member2 Existence Check")
		assertFloatEqual(t, 4.0, score1, "Member2 Score Check")

		ok, score2 := zset.ZScore(key, "member3")
		assertBoolEqual(t, true, ok, "Member3 Existence Check")
		assertFloatEqual(t, 2.0, score2, "Member3 Score Check")

	})

	t.Run("Update the Score of an Existing Member", func(t *testing.T) {
		// Test updating the score of an existing member in a sorted set.
		key := "sorted_set"
		score := 1.0
		member := "member2"
		value := "updated_value"
		count := zset.ZAdd(key, score, member, value)

		assertCountEqual(t, 1, count, "Update Member Score")

		ok, addedScore := zset.ZScore(key, member)
		assertBoolEqual(t, true, ok, "Updated Member Existence Check")
		assertFloatEqual(t, score, addedScore, "Updated Member Score Check")

		ok, _ = zset.ZScore(key, member)
		assertBoolEqual(t, true, ok, "Updated Member Value Existence Check")

	})
}

func TestZSet_ZScore(t *testing.T) {
	zset := New()

	t.Run("Get Score of a Non-Existent Member", func(t *testing.T) {
		// Test getting the score of a member that does not exist in the sorted set.
		key := "sorted_set"
		member := "nonexistent_member"

		ok, score := zset.ZScore(key, member)

		assertBoolEqual(t, false, ok, "Non-Existent Member Existence Check")
		assertFloatEqual(t, 0.0, score, "Non-Existent Member Score Check")
	})

	t.Run("Get Score of an Existing Member", func(t *testing.T) {
		// Test getting the score of an existing member in the sorted set.
		key := "sorted_set"
		score := 5.0
		member := "existing_member"
		value := "value1"

		zset.ZAdd(key, score, member, value)

		ok, retrievedScore := zset.ZScore(key, member)

		assertBoolEqual(t, true, ok, "Existing Member Existence Check")
		assertFloatEqual(t, score, retrievedScore, "Existing Member Score Check")
	})
}

func TestZSet_ZCard(t *testing.T) {
	zset := New()

	t.Run("Get Cardinality of an Empty ZSet", func(t *testing.T) {
		// Test getting the cardinality of an empty sorted set.
		key := "empty_sorted_set"
		cardinality := zset.ZCard(key)

		assertCountEqual(t, 0, cardinality, "Cardinality of Empty Sorted Set")
	})

	t.Run("Get Cardinality of a Non-Empty Sorted Set", func(t *testing.T) {
		// Test getting the cardinality of a non-empty sorted set.
		key := "non_empty_sorted_set"
		zset.ZAdd(key, 1.0, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		cardinality := zset.ZCard(key)

		assertCountEqual(t, 2, cardinality, "Cardinality of Non-Empty Sorted Set")
	})
}

func TestZSet_ZRank(t *testing.T) {
	zset := New()

	t.Run("Rank of Non-Existent Key", func(t *testing.T) {
		// Test getting the rank of a member for a non-existent key.
		rank := zset.ZRank("nonexistent_key", "member")
		assertIntEqual(t, -1, rank, "Rank of Non-Existent Key")
	})

	t.Run("Rank of Non-Existent Member", func(t *testing.T) {
		// Test getting the rank of a non-existent member in an existing key.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")

		rank := zset.ZRank(key, "nonexistent_member")
		assertIntEqual(t, -1, rank, "Rank of Non-Existent Member")
	})

	t.Run("Rank of Single Member", func(t *testing.T) {
		// Test getting the rank of a single member in a sorted set.
		key := "sorted_set"
		member := "member1"

		zset.ZAdd(key, 3.0, member, "value1")
		rank := zset.ZRank(key, member)
		assertIntEqual(t, 0, rank, "Rank of Single Member")
	})

	t.Run("Rank of Members with Same Score", func(t *testing.T) {
		// Test getting the rank of members with the same score in a sorted set.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 3.0, "member2", "value2")

		rank1 := zset.ZRank(key, "member1")
		rank2 := zset.ZRank(key, "member2")

		assertIntEqual(t, 0, rank1, "Rank of Member1 with Same Score")
		assertIntEqual(t, 1, rank2, "Rank of Member2 with Same Score")
	})

	t.Run("Rank of Multiple Members", func(t *testing.T) {
		// Test getting the rank of multiple members with different scores.
		key := "sorted_set"
		zset.ZAdd(key, 5.0, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 7.0, "member3", "value3")

		rank1 := zset.ZRank(key, "member1")
		rank2 := zset.ZRank(key, "member2")
		rank3 := zset.ZRank(key, "member3")

		assertIntEqual(t, 1, rank1, "Rank of Member1")
		assertIntEqual(t, 0, rank2, "Rank of Member2")
		assertIntEqual(t, 2, rank3, "Rank of Member3")

	})
}

func TestZSet_ZRevRank(t *testing.T) {
	zset := New()

	t.Run("Reverse Rank of Non-Existent Key", func(t *testing.T) {
		// Test getting the reverse rank of a member for a non-existent key.
		revRank := zset.ZRevRank("nonexistent_key", "member")
		assertIntEqual(t, -1, revRank, "Reverse Rank of Non-Existent Key")
	})

	t.Run("Reverse Rank of Non-Existent Member", func(t *testing.T) {
		// Test getting the reverse rank of a non-existent member in an existing key.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		revRank := zset.ZRevRank(key, "nonexistent_member")
		assertIntEqual(t, -1, revRank, "Reverse Rank of Non-Existent Member")
	})

	t.Run("Reverse Rank of Single Member", func(t *testing.T) {
		// Test getting the reverse rank of a single member in a sorted set.
		key := "sorted_set"
		member := "member1"
		zset.ZAdd(key, 3.0, member, "value1")
		revRank := zset.ZRevRank(key, member)
		assertIntEqual(t, 0, revRank, "Reverse Rank of Single Member")
	})

	t.Run("Reverse Rank of Members with Same Score", func(t *testing.T) {
		// Test getting the reverse rank of members with the same score in a sorted set.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 3.0, "member2", "value2")

		revRank1 := zset.ZRevRank(key, "member1")
		revRank2 := zset.ZRevRank(key, "member2")

		assertIntEqual(t, 1, revRank1, "Reverse Rank of Member1 with Same Score")
		assertIntEqual(t, 0, revRank2, "Reverse Rank of Member2 with Same Score")

	})

	t.Run("Reverse Rank of Multiple Members", func(t *testing.T) {
		// Test getting the reverse rank of multiple members with different scores.
		key := "sorted_set"
		zset.ZAdd(key, 5.0, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 7.0, "member3", "value3")

		revRank1 := zset.ZRevRank(key, "member1")
		revRank2 := zset.ZRevRank(key, "member2")
		revRank3 := zset.ZRevRank(key, "member3")

		assertIntEqual(t, 1, revRank1, "Reverse Rank of Member1")
		assertIntEqual(t, 2, revRank2, "Reverse Rank of Member2")
		assertIntEqual(t, 0, revRank3, "Reverse Rank of Member3")
	})

	t.Run("RevRank Non-Existent Key", func(t *testing.T) {
		// Test getting reverse rank for a non-existent key.
		revRank := zset.ZRevRank("nonexistent_key", "member1")
		assertInt64Equal(t, -1, revRank, "RevRank Non-Existent Key")
	})

	t.Run("RevRank Non-Existent Member", func(t *testing.T) {
		// Test getting reverse rank for a non-existent member in an existing key.
		key := "sorted_set"
		revRank := zset.ZRevRank(key, "nonexistent_member")
		assertInt64Equal(t, -1, revRank, "RevRank Non-Existent Member")
	})

	t.Run("RevRank Multiple Members", func(t *testing.T) {
		// Test getting reverse rank for a sorted set with multiple members.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 4.0, "member3", "value3")
		revRank := zset.ZRevRank(key, "member2")
		assertInt64Equal(t, 2, revRank, "RevRank Multiple Members")
	})
}

func TestZSet_ZRem(t *testing.T) {

	t.Run("Remove Non-Existent Member", func(t *testing.T) {
		// Test removing a non-existent member from a non-existent key.
		zset := New()

		removed := zset.ZRem("nonexistent_key", "nonexistent_member")
		assertBoolEqual(t, false, removed, "Remove Non-Existent Member")
	})

	t.Run("Remove Existing Member", func(t *testing.T) {
		// Test removing an existing member from a sorted set.
		zset := New()

		key := "sorted_set"
		member := "member1"
		zset.ZAdd(key, 3.0, member, "value1")

		removed := zset.ZRem(key, member)
		assertBoolEqual(t, true, removed, "Remove Existing Member")
		_, exists := zset.records[key].records[member]
		assertBoolEqual(t, false, exists, "Verify Removal of Member")
	})

	t.Run("Remove Non-Existent Member from Existing Key", func(t *testing.T) {
		// Test removing a non-existent member from an existing key.
		zset := New()

		key := "sorted_set"
		removed := zset.ZRem(key, "nonexistent_member")
		assertBoolEqual(t, false, removed, "Remove Non-Existent Member from Existing Key")
		// Verify that the set remains unchanged.
		// _, exists := zset.records[key].records["nonexistent_member"]
		// assertBoolEqual(t, false, exists, "Verify Non-Existence of Non-Existent Member")
	})

	t.Run("Remove Member with Same Score", func(t *testing.T) {
		// Test removing a member with the same score as another member in a sorted set.
		zset := New()

		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 3.0, "member2", "value2")
		removed := zset.ZRem(key, "member1")

		assertBoolEqual(t, true, removed, "Remove Member with Same Score")
		// Verify that the correct member has been removed.
		_, exists1 := zset.records[key].records["member1"]
		_, exists2 := zset.records[key].records["member2"]
		assertBoolEqual(t, false, exists1, "Verify Removal of Member1")
		assertBoolEqual(t, true, exists2, "Verify Retention of Member2")
	})

	t.Run("Remove Last Member", func(t *testing.T) {
		// Test removing the last member from a sorted set.
		zset := New()

		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		removed := zset.ZRem(key, "member1")
		assertBoolEqual(t, true, removed, "Remove Last Member")
		_, exists := zset.records[key]
		assertBoolEqual(t, true, exists, "Verify Empty Set")
	})

	t.Run("Remove Non-Existent Member with Same Score", func(t *testing.T) {
		// Test removing a non-existent member with the same score as another member in a sorted set.
		zset := New()

		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 3.0, "member2", "value2")
		removed := zset.ZRem(key, "nonexistent_member")
		assertBoolEqual(t, false, removed, "Remove Non-Existent Member with Same Score")
		// Verify that the set remains unchanged.
		_, exists1 := zset.records[key].records["member1"]
		_, exists2 := zset.records[key].records["member2"]
		assertBoolEqual(t, true, exists1, "Verify Retention of Member1")
		assertBoolEqual(t, true, exists2, "Verify Retention of Member2")
	})
}

func TestZSet_ZRange(t *testing.T) {
	zset := New()

	t.Run("ZRange Non-Existent Key", func(t *testing.T) {
		// Test retrieving a range for a non-existent key.
		result := zset.ZRange("nonexistent_key", 0, 1)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRange Key Exists, Start > Stop", func(t *testing.T) {
		// Test retrieving a range for an existing key with start greater than stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		result := zset.ZRange(key, 1, 0)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRange Key Exists, Elements in Range", func(t *testing.T) {
		// Test retrieving a range for an existing key with elements in the specified range.
		zset := New()

		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 4.0, "member3", "value3")
		zset.ZAdd(key, 2.2, "member4", "value4")
		zset.ZAdd(key, 3.6, "member5", "value5")
		results := zset.ZRange(key, 2, 2)
		expectedResults := []interface{}{
			"member1",
		}
		assertSliceEqual(t, expectedResults, results, "ZRange Key Exists, Elements in Range")
	})

	t.Run("ZRange Key Exists, Start Equals Stop", func(t *testing.T) {
		// Test retrieving a range for an existing key with start equals stop.
		zset := New()

		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 5.0, "member2", "value2")
		zset.ZAdd(key, 2.2, "member3", "value3")
		zset.ZAdd(key, 4.2, "member4", "value4")
		zset.ZAdd(key, 1.2, "member5", "value5")
		results := zset.ZRange(key, 3, 3)
		expectedResults := []interface{}{"member4"}
		assertSliceEqual(t, expectedResults, results, "ZRange Key Exists, Start Equals Stop")
	})

	t.Run("ZRange Key Exists, Negative Start and Stop", func(t *testing.T) {
		// Test retrieving a range for an existing key with negative start and stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		result := zset.ZRange(key, -1, 1)
		expected := []interface{}{"member1"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRange Key Exists, Out of Bounds Start and Stop", func(t *testing.T) {
		// Test retrieving a range for an existing key with out of bounds start and stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		results := zset.ZRange(key, 0, 5)
		expectedResults := []interface{}{"member2", "member1"}
		assertSliceEqual(t, expectedResults, results, "ZRange Key Exists, Out of Bounds Start and Stop")
	})
}

func TestZSet_ZRevRange(t *testing.T) {
	// Create a new ZSet instance for testing.
	zset := New()

	t.Run("ZRevRange Non-Existent Key", func(t *testing.T) {
		// Test retrieving a reverse range for a non-existent key.
		result := zset.ZRevRange("nonexistent_key", 0, 1)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRevRange Key Exists, Start > Stop", func(t *testing.T) {
		// Test retrieving a reverse range for an existing key with start greater than stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		result := zset.ZRevRange(key, 1, 1)
		expected := []interface{}{"member1"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRevRange Key Exists, Elements in Range", func(t *testing.T) {
		// Test retrieving a reverse range for an existing key with elements in the specified range.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 4.0, "member3", "value3")
		results := zset.ZRevRange(key, 0, 1)
		expectedResults := []interface{}{
			"member3",
			"member1",
		}
		assertSliceEqual(t, expectedResults, results, "ZRevRange Key Exists, Elements in Range")
	})

	t.Run("ZRevRange Key Exists, Start Equals Stop", func(t *testing.T) {
		// Test retrieving a reverse range for an existing key with start equals stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		result := zset.ZRevRange(key, 1, 1)
		expected := []interface{}{"member2"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ZRevRange Key Exists, Negative Start and Stop", func(t *testing.T) {
		// Test retrieving a reverse range for an existing key with negative start and stop.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		result := zset.ZRevRange(key, -1, -1)
		expected := []interface{}{"member2"}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})
}

func TestZSet_ZScoreRange(t *testing.T) {
	zset := New()

	t.Run("ScoreRange Non-Existent Key", func(t *testing.T) {
		// Test retrieving score range for a non-existent key.

		result := zset.ZScoreRange("nonexistent_key", 2.0, 4.0)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ScoreRange Key Exists, No Elements in Range", func(t *testing.T) {
		// Test retrieving score range for an existing key with no elements in the specified range.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")

		result := zset.ZScoreRange(key, 4.0, 5.0)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ScoreRange Key Exists, Elements in Range", func(t *testing.T) {
		// Test retrieving score range for an existing key with elements in the specified range.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.5, "member2", "value2")
		zset.ZAdd(key, 4.2, "member3", "value3")
		result := zset.ZScoreRange(key, 2.4, 4.0)
		expected := []interface{}{
			"member2", 2.5,
			"member1", 3.5,
		}
		assertSliceEqual(t, expected, result, "ScoreRange Key Exists, Elements in Range")
	})

	t.Run("ScoreRange Key Exists, Min > Max", func(t *testing.T) {
		// Test retrieving score range for an existing key with min score greater than max score.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZAdd(key, 4.2, "member3", "value3")
		result := zset.ZScoreRange(key, 4.0, 2.5)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ScoreRange Out of Bounds", func(t *testing.T) {
		// Test score range with bounds outside the available scores.
		zset.ZAdd("set2", 3.5, "member1", "value1")

		// Scores are 3.5, so the range 1.0 to 2.0 is out of bounds.
		result := zset.ZScoreRange("set2", 1.0, 2.0)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("ScoreRange Reversed Bounds", func(t *testing.T) {
		// Test score range with reversed bounds (min > max).
		zset.ZAdd("set3", 3.5, "member1", "value1")

		// Reversed bounds should result in an empty slice.
		result := zset.ZScoreRange("set3", 2.0, 1.0)
		expected := []interface{}{}
		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})
}

func TestZSet_ZRevScoreRange(t *testing.T) {
	zset := New()

	t.Run("RevScoreRange Non-Existent Key", func(t *testing.T) {
		// Test getting a reverse score range for a non-existent key.
		expected := []interface{}{}
		result := zset.ZRevScoreRange("nonexistent_key", 5.0, 0.0)

		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("RevScoreRange Empty Slice", func(t *testing.T) {
		// Test getting a reverse score range for an existing key with an empty sorted set.
		key := "sorted_set"
		expected := []interface{}{}
		result := zset.ZRevScoreRange(key, 4.0, 2.0)

		if len(result) != len(expected) {
			t.Errorf("Expected %v but got %v", expected, result)
		}
	})

	t.Run("RevScoreRange Single Member", func(t *testing.T) {
		// Test getting a reverse score range for a sorted set with a single member.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		result := zset.ZRevScoreRange(key, 5.0, 0.0)
		assertSliceEqual(t, []interface{}{"member1", 3.0}, result, "RevScoreRange Single Member")
	})

	t.Run("RevScoreRange Multiple Members", func(t *testing.T) {
		// Test getting a reverse score range for a sorted set with multiple members.
		key := "sorted_set"
		zset.ZAdd(key, 3.0, "member1", "value1")
		zset.ZAdd(key, 2.5, "member2", "value2")
		zset.ZAdd(key, 4.0, "member3", "value3")
		result := zset.ZRevScoreRange(key, 4.0, 2.0)
		assertSliceEqual(t, []interface{}{"member3", 4.0, "member1", 3.0, "member2", 2.5}, result, "RevScoreRange Multiple Members")
	})

}

func TestZSet_ZKeys(t *testing.T) {
	zset := New()

	t.Run("Keys Empty ZSet", func(t *testing.T) {
		// Test getting keys from an empty ZSet.
		keys := zset.ZKeys()
		if len(keys) != 0 {
			t.Errorf("Expected %v but got %v", 0, keys)
		}
	})

	t.Run("Keys Non-Empty ZSet", func(t *testing.T) {
		// Test getting keys from a non-empty ZSet.
		zset.ZAdd("set1", 3.5, "member1", "value1")
		zset.ZAdd("set2", 2.0, "member2", "value2")
		keys := zset.ZKeys()
		expected := []string{"set1", "set2"}

		if len(expected) != len(keys) {
			t.Errorf("Expected %v but got %v", expected, keys)
		}
	})

	t.Run("Keys After Clear", func(t *testing.T) {
		// Test getting keys after clearing the ZSet.
		zset.ZAdd("set1", 3.5, "member1", "value1")
		zset.ZAdd("set2", 2.0, "member2", "value2")
		zset.ZClear("set1")
		keys := zset.ZKeys()
		expected := []string{"set2"}
		if len(expected) != len(keys) {
			t.Errorf("Expected %v but got %v", expected, keys)
		}
	})
}

func TestZSet_ZKeyExists(t *testing.T) {
	zset := New()

	t.Run("KeyExists Non-Existent Key", func(t *testing.T) {
		// Test checking existence of a non-existent key.
		exists := zset.ZKeyExists("nonexistant_key")
		assertBoolEqual(t, false, exists, "KeyExists Non-Existent Key")
	})

	t.Run("KeyExists Existing Key", func(t *testing.T) {
		// Test checking existence of an existing key.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		exists := zset.ZKeyExists(key)
		assertBoolEqual(t, true, exists, "KeyExists Existing Key")
	})
}

func TestZSet_ZClear(t *testing.T) {
	zset := New()

	t.Run("Clear Non-Existent Key", func(t *testing.T) {
		// Test clearing a non-existent key.
		keys := zset.ZKeys()
		if len(keys) != 0 {
			t.Errorf("Expected %v but got %v", 0, keys)
		}
	})

	t.Run("Clear Existing Key", func(t *testing.T) {
		// Test clearing an existing key.
		key := "sorted_set"
		zset.ZAdd(key, 3.5, "member1", "value1")
		zset.ZAdd(key, 2.0, "member2", "value2")
		zset.ZClear(key)
		keys := zset.ZKeys()
		if len(keys) != 0 {
			t.Errorf("Expected %v but got %v", 0, keys)
		}
	})
}

func assertSliceEqual(t *testing.T, expected, actual []interface{}, message string) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s: Expected %v, got %v", message, expected, actual)
	}
}

func assertCountEqual(t *testing.T, expected, actual int, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected count %d, got %d", message, expected, actual)
	}
}

func assertIntEqual(t *testing.T, expected, actual int64, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected %d, got %d", message, expected, actual)
	}
}

func assertInt64Equal(t *testing.T, expected, actual int64, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected %d, got %d", message, expected, actual)
	}
}

func assertBoolEqual(t *testing.T, expected, actual bool, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected %v, got %v", message, expected, actual)
	}
}

func assertFloatEqual(t *testing.T, expected, actual float64, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected %f, got %f", message, expected, actual)
	}
}
