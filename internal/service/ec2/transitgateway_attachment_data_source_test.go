package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayAttachmentDataSource_Filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "resource_id"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "resource_owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_type", "vpc"),
					resource.TestCheckResourceAttrSet(dataSourceName, "state"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "transit_gateway_owner_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayAttachmentDataSource_ID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAttachmentDataSourceConfig_id(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "resource_id"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "resource_owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_type", "vpc"),
					resource.TestCheckResourceAttrSet(dataSourceName, "state"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "transit_gateway_owner_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayAttachmentDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_attachment" "test" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }

  filter {
    name   = "resource-type"
    values = ["vpc"]
  }

  filter {
    name   = "resource-id"
    values = [aws_vpc.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_vpc_attachment.test]
}
`, rName))
}

func testAccTransitGatewayAttachmentDataSourceConfig_id(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_attachment" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`, rName))
}
