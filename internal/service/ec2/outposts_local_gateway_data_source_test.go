// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2OutpostsLocalGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					acctest.CheckResourceAttrAccountID(dataSourceName, "owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayDataSourceConfig_id() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway" "test" {
  id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}
