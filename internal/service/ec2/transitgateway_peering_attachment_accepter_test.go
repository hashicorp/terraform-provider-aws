package ec2_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccTransitGatewayPeeringAttachmentAccepter_basic_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", sdkacctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepter_Tags_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", sdkacctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tagsSameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_tagsUpdatedSameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_tagsUpdatedSameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepter_basic_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", sdkacctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceNamePeer, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentAccepterBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"

  peer_account_id         = aws_ec2_transit_gateway.test.owner_id
  peer_region             = data.aws_region.current.name
  peer_transit_gateway_id = aws_ec2_transit_gateway.test.id
  transit_gateway_id      = aws_ec2_transit_gateway.peer.id
}
`, rName)
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterBaseConfig(rName),
		`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_tagsSameAccount(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName))
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_tagsUpdatedSameAccount(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		testAccTransitGatewayPeeringAttachmentAccepterBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key3 = "Value3"
    Key2 = "Value2b"
  }
}
`, rName))
}

func testAccAlternateAccountAlternateRegionProviderConfig() string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "awsalternate" {
  access_key = %[1]q
  profile    = %[2]q
  region     = %[3]q
  secret_key = %[4]q
}
`, os.Getenv(conns.EnvVarAlternateAccessKeyId), os.Getenv(conns.EnvVarAlternateProfile), acctest.AlternateRegion(), os.Getenv(conns.EnvVarAlternateSecretAccessKey))
}

func testAccTransitGatewayPeeringAttachmentAccepterConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccAlternateAccountAlternateRegionProviderConfig(),
		testAccTransitGatewayPeeringAttachmentAccepterBaseConfig(rName),
		`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}
