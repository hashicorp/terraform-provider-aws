package medialive_test

import (
	"testing"
)

func TestAccMediaLive_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Multiplex": {
			"basic":      testAccMediaLiveMultiplex_basic,
			"disappears": testAccMediaLiveMultiplex_disappears,
			"update":     testAccMediaLiveMultiplex_update,
			"updateTags": testAccMediaLiveMultiplex_updateTags,
			"start":      testAccMediaLiveMultiplex_start,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
