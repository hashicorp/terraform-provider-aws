package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayVPCAttachmentsDataSource_Filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ec2_transit_gateway_vpc_attachments.by_attachment_id", "ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_ec2_transit_gateway_vpc_attachments.by_gateway_id", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test1" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test1.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"
  vpc_id            = aws_vpc.test2.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test1" {
  subnet_ids         = [aws_subnet.test1.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test2" {
  subnet_ids         = [aws_subnet.test2.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test2.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_vpc_attachments" "by_attachment_id" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_vpc_attachment.test1.id]
  }

  depends_on = [aws_ec2_transit_gateway_vpc_attachment.test1, aws_ec2_transit_gateway_vpc_attachment.test2]
}

data "aws_ec2_transit_gateway_vpc_attachments" "by_gateway_id" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_vpc_attachment.test1, aws_ec2_transit_gateway_vpc_attachment.test2]
}
`, rName))
}
