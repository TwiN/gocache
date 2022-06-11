package gocache

import "testing"

func TestMatchPattern(t *testing.T) {
	scenarios := []struct {
		pattern         string
		key             string
		expectedToMatch bool
	}{
		{
			pattern:         "*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "**",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "living*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*living*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*123",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*_*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "living*_*3",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "living*room_*3",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "living*room_*3",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*vin*om*2*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "livingroom_123",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "*livingroom_123*",
			key:             "livingroom_123",
			expectedToMatch: true,
		},
		{
			pattern:         "livingroom",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
		{
			pattern:         "livingroom123",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
		{
			pattern:         "what",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
		{
			pattern:         "*what*",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
		{
			pattern:         "*.*",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
		{
			pattern:         "room*123",
			key:             "livingroom_123",
			expectedToMatch: false,
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.pattern+"---"+scenario.key, func(t *testing.T) {
			matched := MatchPattern(scenario.pattern, scenario.key)
			if scenario.expectedToMatch {
				if !matched {
					t.Errorf("%s should've matched pattern '%s'", scenario.key, scenario.pattern)
				}
			} else {
				if matched {
					t.Errorf("%s shouldn't have matched pattern '%s'", scenario.key, scenario.pattern)
				}
			}
		})
	}
}
