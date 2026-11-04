// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerTransitGatewayRouteTableAttachmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_transit_gateway_route_table_attachment.test"
	resourceName := "aws_networkmanager_transit_gateway_route_table_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_id", resourceName, names.AttrID),
					// The API returns no policy rule number or segment for this attachment, and the SDKv2 resource stores them as "0" and "".
					resource.TestCheckNoResourceAttr(dataSourceName, "attachment_policy_rule_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_type", resourceName, "attachment_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_arn", resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_id", resourceName, "core_network_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, "edge_location", resourceName, "edge_location"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, "peering_id", resourceName, "peering_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckNoResourceAttr(dataSourceName, "segment_name"),
					// The resource records the state at creation, before the attachment is accepted.
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "AVAILABLE"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_route_table_arn", resourceName, "transit_gateway_route_table_arn"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccNetworkManagerTransitGatewayRouteTableAttachmentDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourcePrefix := "data.aws_networkmanager_transit_gateway_route_table_attachment."
	resource2Name := "aws_networkmanager_transit_gateway_route_table_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayRouteTableAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_route_table_arn", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_route_table_arn", "transit_gateway_route_table_arn", resource2Name, "transit_gateway_route_table_arn"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "tags.Filter", "second"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "peering_id", resource2Name, "peering_id"),
				),
			},
			{
				// Both attachments use the same peering.
				Config:      testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersMultipleMatches(rName),
				ExpectError: regexache.MustCompile(`multiple\s+Network\s+Manager\s+Transit\s+Gateway\s+Route\s+Table\s+Attachments\s+matched`),
			},
			{
				Config:      testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersNoMatch(rName),
				ExpectError: regexache.MustCompile(`no\s+matching\s+Network\s+Manager\s+Transit\s+Gateway\s+Route\s+Table\s+Attachment\s+found`),
			},
		},
	})
}

func testAccTransitGatewayRouteTableAttachmentDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayRouteTableAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.test.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.test.attachment_type
}

data "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  attachment_id = aws_networkmanager_attachment_accepter.test.attachment_id
}
`, rName))
}

func testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersBase(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayRouteTableAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_route_table" "test2" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test.arn

  tags = {
    Name   = %[1]q
    Filter = "first"
  }

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.test.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.test.attachment_type
}

resource "aws_networkmanager_transit_gateway_route_table_attachment" "test2" {
  peering_id                      = aws_networkmanager_transit_gateway_peering.test.id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test2.arn

  tags = {
    Name   = %[1]q
    Filter = "second"
  }

  depends_on = [aws_ec2_transit_gateway_policy_table_association.test]
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_transit_gateway_route_table_attachment.test2.id
  attachment_type = aws_networkmanager_transit_gateway_route_table_attachment.test2.attachment_type
}
`, rName))
}

func testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filters(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_transit_gateway_route_table_attachment" "by_route_table_arn" {
  core_network_id                 = aws_networkmanager_transit_gateway_route_table_attachment.test2.core_network_id
  transit_gateway_route_table_arn = aws_ec2_transit_gateway_route_table.test2.arn

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}

data "aws_networkmanager_transit_gateway_route_table_attachment" "by_tags" {
  core_network_id = aws_networkmanager_transit_gateway_route_table_attachment.test2.core_network_id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}

data "aws_networkmanager_transit_gateway_route_table_attachment" "by_all_filters" {
  core_network_id = aws_networkmanager_transit_gateway_route_table_attachment.test2.core_network_id
  edge_location   = data.aws_region.current.region
  peering_id      = aws_networkmanager_transit_gateway_peering.test.id
  state           = "AVAILABLE"

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`)
}

func testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersMultipleMatches(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  core_network_id = aws_networkmanager_transit_gateway_route_table_attachment.test2.core_network_id
  peering_id      = aws_networkmanager_transit_gateway_peering.test.id

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`)
}

func testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersNoMatch(rName string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayRouteTableAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_transit_gateway_route_table_attachment" "test" {
  core_network_id = aws_networkmanager_transit_gateway_route_table_attachment.test2.core_network_id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`)
}
