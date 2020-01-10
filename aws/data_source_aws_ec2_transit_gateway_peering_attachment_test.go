package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Filter(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter(rName),
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
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID(rName),
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

func TestAccAWSEc2TransitGatewayPeeringAttachmentDataSource_Tags(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigTags(rName),
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

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfig_base(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.peer.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.test.id}"

  tags = {
    Name = %[2]q
  }
}
`, testAccGetAlternateRegion(), rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigFilter(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfig_base(rName) + fmt.Sprintf(`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["${aws_ec2_transit_gateway_peering_attachment.test.id}"]
  }
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigID(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfig_base(rName) + fmt.Sprintf(`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  id = "${aws_ec2_transit_gateway_peering_attachment.test.id}"
}
`)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfigTags(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentDataSourceConfig_base(rName) + fmt.Sprintf(`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  tags = {
    Name = "${aws_ec2_transit_gateway_peering_attachment.test.tags["Name"]}"
  }
}
`)
}
