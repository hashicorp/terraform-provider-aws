// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2HostedZoneIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneIDDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.main", names.AttrID, tfelbv2.HostedZoneIDPerRegionALBMap[acctest.Region()]),
				),
			},
			{
				Config: testAccHostedZoneIDDataSourceConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.regional", names.AttrID, "Z32O12XQLNTSW2"),
				),
			},
			{
				Config: testAccHostedZoneIDDataSourceConfig_explicitNetwork,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.network", names.AttrID, tfelbv2.HostedZoneIDPerRegionNLBMap[acctest.Region()]),
				),
			},
			{
				Config: testAccHostedZoneIDDataSourceConfig_explicitNetworkRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_lb_hosted_zone_id.network-regional", names.AttrID, "Z2IFOLAFXWLO4F"),
				),
			},
		},
	})
}

const testAccHostedZoneIDDataSourceConfig_basic = `
data "aws_lb_hosted_zone_id" "main" {}
`

// lintignore:AWSAT003
const testAccHostedZoneIDDataSourceConfig_explicitRegion = `
data "aws_lb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`

const testAccHostedZoneIDDataSourceConfig_explicitNetwork = `
data "aws_lb_hosted_zone_id" "network" {
  load_balancer_type = "network"
}
`

// lintignore:AWSAT003
const testAccHostedZoneIDDataSourceConfig_explicitNetworkRegion = `
data "aws_lb_hosted_zone_id" "network-regional" {
  region             = "eu-west-1"
  load_balancer_type = "network"
}
`
