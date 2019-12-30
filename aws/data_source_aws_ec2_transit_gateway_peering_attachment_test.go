package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "peer_account_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_region", dataSourceName, "peer_region"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter() string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"

}

resource "aws_ec2_transit_gateway" "first" {
  provider = "aws"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-Filter"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-second-Filter"
  }
}

// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[1]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-peering-attachment-Filter"
  }
}

data "aws_ec2_transit_gateway_peering_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["${aws_ec2_transit_gateway_peering_attachment.test.id}"]
  }
}
`, testAccGetAlternateRegion())
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID() string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "second" {
  provider = "aws.alternate"

}

resource "aws_ec2_transit_gateway" "first" {
  provider = "aws"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-first-ID"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.alternate"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-second-ID"
  }
}

// Create the Peering attachment in the first account...
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = "${data.aws_caller_identity.second.account_id}"
  peer_region             = %[1]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.second.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.first.id}"
  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-peering-attachment-ID"
  }
}

data "aws_ec2_transit_gateway_peering_attachment" "test" {
  id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`, testAccGetAlternateRegion())
}
