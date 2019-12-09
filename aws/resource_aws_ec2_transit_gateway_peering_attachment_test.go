package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"testing"
)

func init() {
	resource.AddTestSweepers("aws_ec2_transit_gateway_peering_attachment", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_peering_attachment",
		F:    testSweepEc2TransitGatewayPeeringAttachments,
	})
}
func testSweepEc2TransitGatewayPeeringAttachments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeTransitGatewayAttachmentsInput{}
	for {
		output, err := conn.DescribeTransitGatewayAttachments(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway Peering Attachment sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateway Peering Attachments: %s", err)
		}
		for _, attachment := range output.TransitGatewayAttachments {
			if aws.StringValue(attachment.ResourceType) != ec2.TransitGatewayAttachmentResourceTypeTgwPeering {
				continue
			}
			if aws.StringValue(attachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
				continue
			}
			id := aws.StringValue(attachment.TransitGatewayAttachmentId)
			input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
				TransitGatewayAttachmentId: aws.String(id),
			}
			log.Printf("[INFO] Deleting EC2 Transit Gateway Peering Attachment: %s", id)
			_, err := conn.DeleteTransitGatewayPeeringAttachment(input)
			if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
				continue
			}
			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment (%s): %s", id, err)
			}
			if err := waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %s", id, err)
			}
		}
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}
	return nil
}
func TestAccAWSEc2TransitGatewayPeeringAttachment_basic(t *testing.T) {
	var transitGatewayPeeringAttachment1 ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.first"
	transitGateway2ResourceName := "aws_ec2_transit_gateway.second"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment1),
					resource.TestCheckResourceAttr(resourceName, "peer_account_id", "true"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGateway2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachment_disappears(t *testing.T) {
	var transitGatewayPeeringAttachment1 ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment1),
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentDisappears(&transitGatewayPeeringAttachment1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
func TestAccAWSEc2TransitGatewayPeeringAttachment_Tags(t *testing.T) {
	var transitGatewayPeeringAttachment1, transitGatewayPeeringAttachment2, transitGatewayPeeringAttachment3 ec2.TransitGatewayPeeringAttachment
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment2),
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentNotRecreated(&transitGatewayPeeringAttachment1, &transitGatewayPeeringAttachment2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment3),
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentNotRecreated(&transitGatewayPeeringAttachment2, &transitGatewayPeeringAttachment3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}
func testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName string, transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Peering Attachment ID is set")
		}
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		attachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)
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
func testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}
		peeringAttachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)
		if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
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
func testAccCheckAWSEc2TransitGatewayPeeringAttachmentDisappears(transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
			TransitGatewayAttachmentId: transitGatewayPeeringAttachment.TransitGatewayAttachmentId,
		}
		if _, err := conn.DeleteTransitGatewayPeeringAttachment(input); err != nil {
			return err
		}
		return waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))
	}
}
func testAccCheckAWSEc2TransitGatewayPeeringAttachmentNotRecreated(i, j *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayAttachmentId) != aws.StringValue(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway Peering Attachment was recreated")
		}
		return nil
	}
}
func testAccAWSEc2TransitGatewayPeeringAttachmentConfig() string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"

}

resource "aws_ec2_transit_gateway" "first" {
  provider = "aws"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-Config"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-second-Config"
  }
}

// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[1]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-peering-attachment-Config"
  }
}
`, testAccGetAlternateRegion())
}
func testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1(tagKey1, tagValue1 string) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"

}

resource "aws_ec2_transit_gateway" "first" {
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-ConfigTags1"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-second-ConfigTags1"
  }
}

// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[3]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1, testAccGetAlternateRegion())
}
func testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"

}

resource "aws_ec2_transit_gateway" "first" {
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-ConfigTags2"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway1-second-ConfigTags2"
  }
}

// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[5]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2, testAccGetAlternateRegion())
}
