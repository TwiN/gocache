package gocache

import "testing"

func TestMatchPattern(t *testing.T) {
	testMatchPattern(t, "*", "livingroom_123", true)
	testMatchPattern(t, "**", "livingroom_123", true)
	testMatchPattern(t, "living*", "livingroom_123", true)
	testMatchPattern(t, "*living*", "livingroom_123", true)
	testMatchPattern(t, "*123", "livingroom_123", true)
	testMatchPattern(t, "*_*", "livingroom_123", true)
	testMatchPattern(t, "living*_*3", "livingroom_123", true)
	testMatchPattern(t, "living*room_*3", "livingroom_123", true)
	testMatchPattern(t, "living*room_*3", "livingroom_123", true)
	testMatchPattern(t, "*vin*om*2*", "livingroom_123", true)
	testMatchPattern(t, "livingroom_123", "livingroom_123", true)
	testMatchPattern(t, "*livingroom_123*", "livingroom_123", true)
	testMatchPattern(t, "livingroom", "livingroom_123", false)
	testMatchPattern(t, "livingroom123", "livingroom_123", false)
	testMatchPattern(t, "what", "livingroom_123", false)
	testMatchPattern(t, "*what*", "livingroom_123", false)
	testMatchPattern(t, "*.*", "livingroom_123", false)
	testMatchPattern(t, "room*123", "livingroom_123", false)
}

func testMatchPattern(t *testing.T, pattern, key string, expectedToMatch bool) {
	matched := MatchPattern(pattern, key)
	if expectedToMatch {
		if !matched {
			t.Errorf("%s should've matched pattern '%s'", key, pattern)
		}
	} else {
		if matched {
			t.Errorf("%s shouldn't have matched pattern '%s'", key, pattern)
		}
	}
}
