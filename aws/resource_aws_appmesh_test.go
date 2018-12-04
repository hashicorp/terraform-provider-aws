package aws

import (
	"testing"
)

func TestAccAWSAppmesh(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Mesh": {
			"basic": testAccAwsAppmeshMesh_basic,
		},
		"VirtualRouter": {
			"basic": testAccAwsAppmeshVirtualRouter_basic,
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
