// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudWatchContributorManagedInsightRulesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix("tfacctest")
	dataSourceName := "data.aws_cloudwatch_contributor_managed_insight_rules.test"
	vpcResource := "aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccContributorManagedInsightRulesDataSourceConfig_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "managed_rules.#", "4"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, vpcResource, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "managed_rules.0.template_name", "VpcEndpointService-NewConnectionsByEndpointId-v1"),
					resource.TestCheckResourceAttr(dataSourceName, "managed_rules.1.template_name", "VpcEndpointService-BytesByEndpointId-v1"),
					resource.TestCheckResourceAttr(dataSourceName, "managed_rules.2.template_name", "VpcEndpointService-RstPacketsByEndpointId-v1"),
					resource.TestCheckResourceAttr(dataSourceName, "managed_rules.3.template_name", "VpcEndpointService-ActiveConnectionsByEndpointId-v1"),
				),
			},
		},
	})
}

func testAccContributorManagedInsightRulesDataSourceConfig_basic(rName string, count int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  count = %[2]d

  load_balancer_type = "network"
  name               = "%[1]s-${count.index}"

  subnets = aws_subnet.test[*].id

  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  network_load_balancer_arns = aws_lb.test[*].arn
}

data "aws_cloudwatch_contributor_managed_insight_rules" "test" {
  resource_arn = aws_vpc_endpoint_service.test.arn
}
`, rName, count))
}
