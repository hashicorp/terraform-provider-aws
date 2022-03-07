package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayPeeringAttachment_basic(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					acctest.CheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_disappears(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckTransitGatewayPeeringAttachmentDisappears(&transitGatewayPeeringAttachment),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_Tags_sameAccount(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentTags1Config_sameAccount(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentTags1Config_sameAccount(rName, "key1", "value1"),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentTags2Config_sameAccount(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachment_differentAccount(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentBasicConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					// Test that the peer account ID != the primary (request) account ID
					func(s *terraform.State) error {
						if acctest.CheckResourceAttrAccountID(resourceName, "peer_account_id") == nil {
							return fmt.Errorf("peer_account_id attribute incorrectly to the requester's account ID")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "peer_region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				Config:            testAccTransitGatewayPeeringAttachmentBasicConfig_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTransitGatewayPeeringAttachmentExists(resourceName string, transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Peering Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		attachment, err := tfec2.DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable && aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStatePendingAcceptance {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment (%s) exists in non-available/pending acceptance (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayPeeringAttachment = *attachment

		return nil
	}
}

func testAccCheckTransitGatewayPeeringAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_peering_attachment" {
			continue
		}

		peeringAttachment, err := tfec2.DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if peeringAttachment == nil {
			continue
		}

		if aws.StringValue(peeringAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(peeringAttachment.State))
		}
	}

	return nil
}

func testAccCheckTransitGatewayPeeringAttachmentDisappears(transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
			TransitGatewayAttachmentId: transitGatewayPeeringAttachment.TransitGatewayAttachmentId,
		}

		if _, err := conn.DeleteTransitGatewayPeeringAttachment(input); err != nil {
			return err
		}

		return tfec2.WaitForTransitGatewayPeeringAttachmentDeletion(conn, aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))
	}
}

func testAccTransitGatewayPeeringAttachmentConfig_base(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName string) string {
	return acctest.ConfigAlternateRegionProvider() + testAccTransitGatewayPeeringAttachmentConfig_base(rName)
}

func testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName string) string {
	return testAccAlternateAccountAlternateRegionProviderConfig() + testAccTransitGatewayPeeringAttachmentConfig_base(rName)
}

func testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName string) string {
	return testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, acctest.AlternateRegion())
}

func testAccTransitGatewayPeeringAttachmentBasicConfig_differentAccount(rName string) string {
	return testAccTransitGatewayPeeringAttachmentConfig_differentAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, acctest.AlternateRegion())
}

func testAccTransitGatewayPeeringAttachmentTags1Config_sameAccount(rName, tagKey1, tagValue1 string) string {
	return testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[2]q

    %[3]s = %[4]q
  }
}
`, acctest.AlternateRegion(), rName, tagKey1, tagValue1)
}

func testAccTransitGatewayPeeringAttachmentTags2Config_sameAccount(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccTransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[2]q

    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, acctest.AlternateRegion(), rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
