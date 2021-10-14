package aws

import (
	"testing"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSCloudHsmV2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			"basic":      testAccAWSCloudHsmV2Cluster_basic,
			"disappears": testAccAWSCloudHsmV2Cluster_disappears,
			"tags":       testAccAWSCloudHsmV2Cluster_Tags,
		},
		"Hsm": {
			"availabilityZone":   testAccAWSCloudHsmV2Hsm_AvailabilityZone,
			"basic":              testAccAWSCloudHsmV2Hsm_basic,
			"disappears":         testAccAWSCloudHsmV2Hsm_disappears,
			"disappears_Cluster": testAccAWSCloudHsmV2Hsm_disappears_Cluster,
			"ipAddress":          testAccAWSCloudHsmV2Hsm_IpAddress,
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
