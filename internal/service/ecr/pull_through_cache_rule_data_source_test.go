// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRPullThroughCacheRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource := "data.aws_ecr_pull_through_cache_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPullThroughCacheRuleDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, "upstream_registry_url", "public.ecr.aws"),
					acctest.CheckResourceAttrAccountID(dataSource, "registry_id"),
				),
			},
		},
	})
}

func testAccPullThroughCacheRuleDataSourceConfig_basic() string {
	return `
resource "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
}

data "aws_ecr_pull_through_cache_rule" "test" {
  ecr_repository_prefix = aws_ecr_pull_through_cache_rule.test.ecr_repository_prefix
}
`
}
