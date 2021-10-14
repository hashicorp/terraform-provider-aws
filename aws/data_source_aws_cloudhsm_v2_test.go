package aws

import (
	"testing"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDataSourceCloudHsmV2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			"basic": testAccDataSourceCloudHsmV2Cluster_basic,
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
