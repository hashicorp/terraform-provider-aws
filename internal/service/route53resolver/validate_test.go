// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
)

func TestValidResolverName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing - 123__",
			ErrCount: 0,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(65, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
		{
			Value:    acctest.Ct1,
			ErrCount: 1,
		},
		{
			Value:    acctest.Ct10,
			ErrCount: 0,
		},
		{
			Value:    "A",
			ErrCount: 0,
		},
	}
	for _, tc := range cases {
		_, errors := tfroute53resolver.ValidResolverName(tc.Value, "aws_route53_resolver_endpoint")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS Route53 Resolver Endpoint Name to not trigger a validation error for %q", tc.Value)
		}
	}
}
