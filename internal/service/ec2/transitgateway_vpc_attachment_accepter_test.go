package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayVPCAttachmentAccepter_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	vpcAttachmentName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"
	callerIdentityDatasourceName := "data.aws_caller_identity.creator"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", vpcAttachmentName, "id"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", callerIdentityDatasourceName, "account_id"),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepter_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, "key1", "value1"),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway ec2.TransitGateway
	var transitGatewayVpcAttachment ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment_accepter.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckTransitGatewayVPCAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, transitGatewayResourceName, &transitGateway),
					testAccCheckTransitGatewayVPCAttachmentExists(ctx, resourceName, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx, &transitGateway, &transitGatewayVpcAttachment),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.creator.account_id
  resource_share_arn = aws_ram_resource_share.test.id
}

# VPC attachment creator.
data "aws_caller_identity" "creator" {
  provider = "awsalternate"
}

resource "aws_vpc" "test" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  provider = "awsalternate"

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  provider = "awsalternate"

  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), `
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayVPCAttachmentAccepterConfig_defaultRouteTableAssociationAndPropagation(rName string, association, propagation bool) string {
	return acctest.ConfigCompose(testAccTransitGatewayVPCAttachmentAccepterConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }

  transit_gateway_default_route_table_association = %[2]t
  transit_gateway_default_route_table_propagation = %[3]t
}
`, rName, association, propagation))
}
