package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayMulticastDomain_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	mCastSenders := "224.0.0.1"
	mCastMembers := "224.0.0.1"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainDataSourceConfig(rName, mCastSenders, mCastMembers),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "aws_ec2_transit_gateway_multicast_domain", dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "sources", dataSourceName, mCastSenders),
					resource.TestCheckResourceAttrPair(resourceName, "members", dataSourceName, mCastMembers),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainDataSourceConfig(rName string, mCastSenders string, mCastMembers string) string {
	return fmt.Sprintf(`

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]
	
  filter {
    name = "name"
	values = [
	  "amzn-ami-hvm-*-x86_64-gp2",
	]
  }
	
  filter {
    name = "owner-alias"
    values = [
  	"amazon",
    ]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}
	
resource "aws_subnet" "test" {
  vpc_id      = aws_vpc.test.id
  cidr_block  = "10.0.1.0/24"
  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test1.id
    tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test1.id
    subnet_ids                    = [aws_subnet.test.id]
  }

  members {
    group_ip_address = "224.0.0.1"
	network_interface_ids = [aws_instance.test.primary_network_interface_id]
	}
	
  sources {
    group_ip_address = "224.0.0.1"
	  network_interface_ids = [aws_instance.test.primary_network_interface_id]
	}

	tags = {
		Name = %[1]q
	  }
}

data "aws_ec2_transit_gateway_multicast_domain" "test" {
	transit_gateway_id = aws_ec2_transit_gateway.test.id
}

`, rName, mCastSenders, mCastMembers)
}
