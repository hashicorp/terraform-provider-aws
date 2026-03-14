// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
)

func TestValidReplicationGroupAuthToken(t *testing.T) {
	t.Parallel()

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
			Value:    acctest.RandStringFromCharSet(t, 129, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfelasticache.ValidReplicationGroupAuthToken(tc.Value, "aws_elasticache_replication_group_auth_token")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the ElastiCache Replication Group AuthToken to trigger a validation error")
		}
	}
}
