package jellyzset

import "testing"

func slicesEqualIgnoreOrder(slice1, slice2 []interface{}) bool {
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
