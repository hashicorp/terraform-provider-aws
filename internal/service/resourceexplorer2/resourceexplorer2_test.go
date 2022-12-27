package resourceexplorer2_test

import (
	"testing"
)

func TestAccResourceExplorer2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Index": {
			"basic":      testAccIndex_basic,
			"disappears": testAccIndex_disappears,
			"tags":       testAccIndex_tags,
			"type":       testAccIndex_type,
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
