package eks

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidClusterName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "my-valid-eks-cluster_1_dev",
			ErrCount: 0,
		},
		{
			Value:    `_invalid`,
			ErrCount: 1,
		},
		{
			Value:    `-invalid`,
			ErrCount: 1,
		},
		{
			Value:    `invalid@`,
			ErrCount: 1,
		},
		{
			Value:    `invalid*`,
			ErrCount: 1,
		},
		{
			Value:    `invalid:`,
			ErrCount: 1,
		},
		{
			Value:    `invalid$`,
			ErrCount: 1,
		},
		{
			Value:    ``,
			ErrCount: 2,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(101, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validClusterName(tc.Value, "cluster_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the EKS Cluster Name to trigger a validation error: %s, expected %d, got %d errors", tc.Value, tc.ErrCount, len(errors))
		}
	}
}
