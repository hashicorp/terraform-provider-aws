// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerDirectConnectGatewayAttachmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	edgeLocation1 := endpoints.UsEast1RegionID
	edgeLocation2 := endpoints.UsWest2RegionID
	dataSourceName := "data.aws_networkmanager_dx_gateway_attachment.test"
	resourceName := "aws_networkmanager_dx_gateway_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentDataSourceConfig_basic(rName, edgeLocation1, edgeLocation2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_policy_rule_number", resourceName, "attachment_policy_rule_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_type", resourceName, "attachment_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_arn", resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_id", resourceName, "core_network_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, "direct_connect_gateway_arn", resourceName, "direct_connect_gateway_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "edge_locations.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "edge_locations.*", edgeLocation1),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "edge_locations.*", edgeLocation2),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, "segment_name", resourceName, "segment_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccNetworkManagerDirectConnectGatewayAttachmentDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	edgeLocation1 := endpoints.UsEast1RegionID
	edgeLocation2 := endpoints.UsWest2RegionID
	dataSourcePrefix := "data.aws_networkmanager_dx_gateway_attachment."
	resource2Name := "aws_networkmanager_dx_gateway_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectConnectGatewayAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectConnectGatewayAttachmentDataSourceConfig_filters(rName, edgeLocation1, edgeLocation2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_dx_gateway_arn", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_dx_gateway_arn", "direct_connect_gateway_arn", resource2Name, "direct_connect_gateway_arn"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_edge_location_subset", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_edge_location_subset", "edge_locations.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "tags.Filter", "second"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "attachment_id", resource2Name, names.AttrID),
				),
			},
			{
				// Both attachments cover the first edge location.
				Config:      testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersMultipleMatches(rName, edgeLocation1, edgeLocation2),
				ExpectError: regexache.MustCompile(`multiple\s+Network\s+Manager\s+Direct\s+Connect\s+Gateway\s+Attachments\s+matched`),
			},
			{
				Config:      testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersNoMatch(rName, edgeLocation1, edgeLocation2),
				ExpectError: regexache.MustCompile(`no\s+matching\s+Network\s+Manager\s+Direct\s+Connect\s+Gateway\s+Attachment\s+found`),
			},
		},
	})
}

func testAccDirectConnectGatewayAttachmentDataSourceConfig_basic(rName, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(
		testAccDirectConnectGatewayAttachmentConfig_multiRegionBase(rName, edgeLocation1, edgeLocation2),
		fmt.Sprintf(`
resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [%[2]q, %[3]q]

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_dx_gateway_attachment" "test" {
  attachment_id = aws_networkmanager_dx_gateway_attachment.test.id
}
`, rName, edgeLocation1, edgeLocation2))
}

func testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersBase(rName, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(
		testAccDirectConnectGatewayAttachmentConfig_multiRegionBase(rName, edgeLocation1, edgeLocation2),
		fmt.Sprintf(`
resource "aws_dx_gateway" "test2" {
  name            = "%[1]s-2"
  amazon_side_asn = 65001
}

resource "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test.arn
  edge_locations             = [%[2]q]

  tags = {
    Name   = %[1]q
    Filter = "first"
  }
}

resource "aws_networkmanager_dx_gateway_attachment" "test2" {
  core_network_id            = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test2.arn
  edge_locations             = [%[2]q, %[3]q]

  tags = {
    Name   = %[1]q
    Filter = "second"
  }
}
`, rName, edgeLocation1, edgeLocation2))
}

func testAccDirectConnectGatewayAttachmentDataSourceConfig_filters(rName, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(
		testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersBase(rName, edgeLocation1, edgeLocation2),
		fmt.Sprintf(`
data "aws_networkmanager_dx_gateway_attachment" "by_dx_gateway_arn" {
  core_network_id            = aws_networkmanager_dx_gateway_attachment.test2.core_network_id
  direct_connect_gateway_arn = aws_dx_gateway.test2.arn

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}

data "aws_networkmanager_dx_gateway_attachment" "by_edge_location_subset" {
  core_network_id = aws_networkmanager_dx_gateway_attachment.test2.core_network_id
  edge_locations  = [%[2]q]

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}

data "aws_networkmanager_dx_gateway_attachment" "by_tags" {
  core_network_id = aws_networkmanager_dx_gateway_attachment.test2.core_network_id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}

data "aws_networkmanager_dx_gateway_attachment" "by_all_filters" {
  core_network_id = aws_networkmanager_dx_gateway_attachment.test2.core_network_id
  edge_locations  = [%[2]q, %[1]q]
  state           = "AVAILABLE"

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}
`, edgeLocation1, edgeLocation2))
}

func testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersMultipleMatches(rName, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(
		testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersBase(rName, edgeLocation1, edgeLocation2),
		fmt.Sprintf(`
data "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id = aws_networkmanager_dx_gateway_attachment.test2.core_network_id
  edge_locations  = [%[1]q]

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}
`, edgeLocation1))
}

func testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersNoMatch(rName, edgeLocation1, edgeLocation2 string) string {
	return acctest.ConfigCompose(
		testAccDirectConnectGatewayAttachmentDataSourceConfig_filtersBase(rName, edgeLocation1, edgeLocation2), `
data "aws_networkmanager_dx_gateway_attachment" "test" {
  core_network_id = aws_networkmanager_dx_gateway_attachment.test2.core_network_id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_dx_gateway_attachment.test]
}
`)
}
