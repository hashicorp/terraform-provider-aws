// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sts_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSTSCallerIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.STSEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCallerIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckCallerIdentityAccountID("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

const testAccCallerIdentityConfig_basic = `
data "aws_caller_identity" "current" {}
`
