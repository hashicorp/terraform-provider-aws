package medialive_test

import (
	"testing"
)

func TestAccMediaLive_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Multiplex": {
			"basic":      testAccMultiplex_basic,
			"disappears": testAccMultiplex_disappears,
			"update":     testAccMultiplex_update,
			"updateTags": testAccMultiplex_updateTags,
			"start":      testAccMultiplex_start,
		},
		"MultiplexProgram": {
			"basic":      testAccMultiplexProgram_basic,
			"update":     testAccMultiplexProgram_update,
			"disappears": testAccMultiplexProgram_disappears,
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
