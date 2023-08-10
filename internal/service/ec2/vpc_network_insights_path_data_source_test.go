// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCNetworkInsightsPathDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_network_insights_path.test"
	datasourceName := "data.aws_ec2_network_insights_path.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetworkInsightsPathDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "destination", resourceName, "destination"),
					resource.TestCheckResourceAttrPair(datasourceName, "destination_ip", resourceName, "destination_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "destination_port", resourceName, "destination_port"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_insights_path_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "protocol", resourceName, "protocol"),
					resource.TestCheckResourceAttrPair(datasourceName, "source", resourceName, "source"),
					resource.TestCheckResourceAttrPair(datasourceName, "source_ip", resourceName, "source_ip"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccVPCNetworkInsightsPathDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_network_interface" "test" {
  count = 2

  subnet_id = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_network_insights_path" "test" {
  source           = aws_network_interface.test[0].id
  destination      = aws_network_interface.test[1].id
  destination_port = 443
  protocol         = "tcp"

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_network_insights_path" "test" {
  network_insights_path_id = aws_ec2_network_insights_path.test.id
}
`, rName))
}
