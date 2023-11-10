package jellyzset

import (
	"testing"
)

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
}

func slicesEqualIgnoreOrder(t *testing.T, slice1, slice2 []interface{}) bool {
	t.Helper()
	if len(slice1) != len(slice2) {
		return false
	}

	matched := make(map[interface{}]bool)

	for _, item := range slice1 {
		matched[item] = true
	}

	for _, item := range slice2 {
		if !matched[item] {
			return false
		}
	}

	return true
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

func assertStringEqual(t *testing.T, expected, actual string, message string) {
	t.Helper()
	if actual != expected {
		t.Errorf("%s: Expected %s, got %s", message, expected, actual)
	}
}
