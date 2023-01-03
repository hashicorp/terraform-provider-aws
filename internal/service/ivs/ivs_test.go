package ivs_test

import (
	"testing"
)

func TestAccIVS_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"PlaybackKeyPair": {
			"basic":      testAccPlaybackKeyPair_basic,
			"update":     testAccPlaybackKeyPair_update,
			"tags":       testAccPlaybackKeyPair_tags,
			"disappears": testAccPlaybackKeyPair_disappears,
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
