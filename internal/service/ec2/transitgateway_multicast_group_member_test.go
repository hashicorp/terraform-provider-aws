package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccTransitGatewayMulticastGroupMember_basic(t *testing.T) {
	var v ec2.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_member.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayMulticastGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupMemberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupMemberExists(resourceName, &v),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastGroupMember_disappears(t *testing.T) {
	var v ec2.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_member.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayMulticastGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupMemberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupMemberExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayMulticastGroupMember(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayMulticastGroupMember_Disappears_domain(t *testing.T) {
	var v ec2.TransitGatewayMulticastGroup
	resourceName := "aws_ec2_transit_gateway_multicast_group_member.test"
	domainResourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayMulticastGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupMemberConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupMemberExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGatewayMulticastDomain(), domainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayMulticastGroupMember_twoMembers(t *testing.T) {
	var v1, v2 ec2.TransitGatewayMulticastGroup
	resource1Name := "aws_ec2_transit_gateway_multicast_group_member.test1"
	resource2Name := "aws_ec2_transit_gateway_multicast_group_member.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayMulticastGroupMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastGroupMemberConfig_twoMembers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayMulticastGroupMemberExists(resource1Name, &v1),
					testAccCheckTransitGatewayMulticastGroupMemberExists(resource2Name, &v2),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayMulticastGroupMemberExists(n string, v *ec2.TransitGatewayMulticastGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Multicast Group Member ID is set")
		}

		multicastDomainID, groupIPAddress, eniID, err := tfec2.TransitGatewayMulticastGroupMemberParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindTransitGatewayMulticastGroupMemberByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayMulticastGroupMemberDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_multicast_group_member" {
			continue
		}

		multicastDomainID, groupIPAddress, eniID, err := tfec2.TransitGatewayMulticastGroupMemberParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindTransitGatewayMulticastGroupMemberByThreePartKey(conn, multicastDomainID, groupIPAddress, eniID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Transit Gateway Multicast Group Member %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccTransitGatewayMulticastGroupMemberConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

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
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "test" {
  subnet_id                           = aws_subnet.test.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_group_member" "test" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}
`, rName))
}

func testAccTransitGatewayMulticastGroupMemberConfig_twoMembers(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

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
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "test" {
  subnet_id                           = aws_subnet.test.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.test.id
}

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test2" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_group_member" "test1" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test1.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}

resource "aws_ec2_transit_gateway_multicast_group_member" "test2" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test2.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}
`, rName))
}
