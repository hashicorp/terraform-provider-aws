package elasticache

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidReplicationGroupAuthToken(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "this-is-valid!#%()^",
			ErrCount: 0,
		},
		{
			Value:    "this-is-not",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid\"",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid@",
			ErrCount: 1,
		},
		{
			Value:    "this-is-not-valid/",
			ErrCount: 1,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(129, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validReplicationGroupAuthToken(tc.Value, "aws_elasticache_replication_group_auth_token")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the ElastiCache Replication Group AuthToken to trigger a validation error")
		}
	}
}
