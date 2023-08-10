// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsAssetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_asset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, outposts.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostAssetDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "outposts", regexp.MustCompile(`outpost/.+`)),
					resource.TestMatchResourceAttr(dataSourceName, "asset_id", regexp.MustCompile(`^(\w+)$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "asset_type"),
					resource.TestMatchResourceAttr(dataSourceName, "rack_elevation", regexp.MustCompile(`^[\S \n]+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "rack_id", regexp.MustCompile(`^[\S \n]+$`)),
				),
			},
		},
	})
}

func testAccOutpostAssetDataSourceConfig_basic() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_assets" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_asset" "test" {
  arn      = tolist(data.aws_outposts_outposts.test.arns)[0]
  asset_id = data.aws_outposts_assets.test.asset_ids[0]
}

`
}
