// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBHostedZoneIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneIDDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.main", names.AttrID, tfelb.HostedZoneIDPerRegionMap[acctest.Region()]),
				),
			},
			{
				Config: testAccHostedZoneIDDataSourceConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", names.AttrID, "Z32O12XQLNTSW2"),
				),
			},
		},
	})
}

const testAccHostedZoneIDDataSourceConfig_basic = `
data "aws_elb_hosted_zone_id" "main" {}
`

// lintignore:AWSAT003
const testAccHostedZoneIDDataSourceConfig_explicitRegion = `
data "aws_elb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`
