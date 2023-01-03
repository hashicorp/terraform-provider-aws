package kms_test

import (
	"testing"
)

func TestAccKMS_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"CustomKeyStore": {
			"basic":      testAccCustomKeyStore_basic,
			"update":     testAccCustomKeyStore_update,
			"disappears": testAccCustomKeyStore_disappears,
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
