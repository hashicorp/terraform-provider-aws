package appsync_test

import (
	"testing"
)

func TestAccAppSync_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"APIKey": {
			"basic":       testAccAppSyncAPIKey_basic,
			"description": testAccAppSyncAPIKey_description,
			"expires":     testAccAppSyncAPIKey_expires,
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
