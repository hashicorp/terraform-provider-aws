package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_basic_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_Tags_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tags_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Side", "Accepter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachmentAccepter_basic_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment_accepter.test"
	peeringAttachmentName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"
	rName := fmt.Sprintf("tf-testacc-tgwpeerattach-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", transitGatewayResourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", peeringAttachmentName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  provider = "aws.alternate"
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`))
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tags_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  provider = "aws.alternate"
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName))
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_tagsUpdated_sameAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  provider = "aws.alternate"
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
  tags = {
    Name = %[1]q
    Side = "Accepter"
    Key3 = "Value3"
    Key2 = "Value2b"
  }
}
`, rName))
}

func testAccAWSEc2TransitGatewayPeeringAttachmentAccepterConfig_basic_differentAccount(rName string) string {
	return composeConfig(
		testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "test" {
  provider = "aws.alternate"
  transit_gateway_attachment_id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`))
}
