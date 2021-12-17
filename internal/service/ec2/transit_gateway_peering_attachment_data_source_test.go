package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayPeeringAttachmentDataSource_Filter_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentFilterDataSourceConfig_sameAccount(rName),
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

func testAccTransitGatewayPeeringAttachmentDataSource_Filter_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentFilterDataSourceConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, "owner_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_ID_sameAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentIDDataSourceConfig_sameAccount(rName),
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

func testAccTransitGatewayPeeringAttachmentDataSource_ID_differentAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentIDDataSourceConfig_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "peer_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(transitGatewayResourceName, "owner_id", dataSourceName, "peer_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "peer_transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayPeeringAttachmentDataSource_Tags(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_peering_attachment.test"
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayPeeringAttachmentTagsDataSourceConfig_sameAccount(rName),
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

func testAccTransitGatewayPeeringAttachmentFilterDataSourceConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentIDDataSourceConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPeeringAttachmentBasicConfig_sameAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  id = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}

func testAccTransitGatewayPeeringAttachmentTagsDataSourceConfig_sameAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPeeringAttachmentTags1Config_sameAccount(rName, "Name", rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  tags = {
    Name = aws_ec2_transit_gateway_peering_attachment.test.tags["Name"]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentFilterDataSourceConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPeeringAttachmentBasicConfig_differentAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"

  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_peering_attachment.test.id]
  }
}
`)
}

func testAccTransitGatewayPeeringAttachmentIDDataSourceConfig_differentAccount(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayPeeringAttachmentBasicConfig_differentAccount(rName),
		`
data "aws_ec2_transit_gateway_peering_attachment" "test" {
  provider = "awsalternate"
  id       = aws_ec2_transit_gateway_peering_attachment.test.id
}
`)
}
