package cloudhsmv2_test

import (
	"testing"
)

func TestAccCloudHSMV2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			"basic":      testAccCluster_basic,
			"disappears": testAccCluster_disappears,
			"tags":       testAccCluster_Tags,
		},
		"Hsm": {
			"availabilityZone":   testAccHSM_AvailabilityZone,
			"basic":              testAccHSM_basic,
			"disappears":         testAccHSM_disappears,
			"disappears_Cluster": testAccHSM_disappears_Cluster,
			"ipAddress":          testAccHSM_IPAddress,
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
