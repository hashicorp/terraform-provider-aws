// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerSiteToSiteVPNAttachmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	dataSourceName := "data.aws_networkmanager_site_to_site_vpn_attachment.test"
	resourceName := "aws_networkmanager_site_to_site_vpn_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentDataSourceConfig_basic(rName, bgpASN, vpnIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_policy_rule_number", resourceName, "attachment_policy_rule_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_type", resourceName, "attachment_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_arn", resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_id", resourceName, "core_network_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, "edge_location", resourceName, "edge_location"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "segment_name", resourceName, "segment_name"),
					// The resource records the state at creation, before the attachment is accepted.
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "AVAILABLE"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpn_connection_arn", resourceName, "vpn_connection_arn"),
				),
			},
		},
	})
}

func TestAccNetworkManagerSiteToSiteVPNAttachmentDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	bgpASN := acctest.RandIntRange(t, 64512, 65534)
	vpnIP, err := sdkacctest.RandIpAddress("172.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	vpnIP2, err := sdkacctest.RandIpAddress("172.0.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	dataSourcePrefix := "data.aws_networkmanager_site_to_site_vpn_attachment."
	resource2Name := "aws_networkmanager_site_to_site_vpn_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSiteToSiteVPNAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteToSiteVPNAttachmentDataSourceConfig_filters(rName, bgpASN, vpnIP, vpnIP2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_vpn_connection_arn", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_vpn_connection_arn", "vpn_connection_arn", resource2Name, "vpn_connection_arn"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "tags.Filter", "second"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "attachment_id", resource2Name, names.AttrID),
				),
			},
			{
				Config:      testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersMultipleMatches(rName, bgpASN, vpnIP, vpnIP2),
				ExpectError: regexache.MustCompile(`multiple\s+Network\s+Manager\s+Site\s+To\s+Site\s+VPN\s+Attachments\s+matched`),
			},
			{
				Config:      testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersNoMatch(rName, bgpASN, vpnIP, vpnIP2),
				ExpectError: regexache.MustCompile(`no\s+matching\s+Network\s+Manager\s+Site\s+To\s+Site\s+VPN\s+Attachment\s+found`),
			},
		},
	})
}

func testAccSiteToSiteVPNAttachmentDataSourceConfig_basic(rName string, bgpASN int, vpnIP string) string {
	return acctest.ConfigCompose(
		testAccSiteToSiteVPNAttachmentConfig_base(rName, bgpASN, vpnIP),
		fmt.Sprintf(`
resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id    = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn = aws_vpn_connection.test.arn

  tags = {
    segment = "shared"
    Name    = %[1]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}

data "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  attachment_id = aws_networkmanager_attachment_accepter.test.attachment_id
}
`, rName))
}

func testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersBase(rName string, bgpASN int, vpnIP, vpnIP2 string) string {
	return acctest.ConfigCompose(
		testAccSiteToSiteVPNAttachmentConfig_base(rName, bgpASN, vpnIP),
		fmt.Sprintf(`
resource "aws_customer_gateway" "test2" {
  bgp_asn    = %[2]d
  ip_address = %[3]q
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test2" {
  customer_gateway_id = aws_customer_gateway.test2.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id    = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn = aws_vpn_connection.test.arn

  tags = {
    segment = "shared"
    Name    = %[1]q
    Filter  = "first"
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test.attachment_type
}

resource "aws_networkmanager_site_to_site_vpn_attachment" "test2" {
  core_network_id    = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpn_connection_arn = aws_vpn_connection.test2.arn

  tags = {
    segment = "shared"
    Name    = %[1]q
    Filter  = "second"
  }
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_site_to_site_vpn_attachment.test2.id
  attachment_type = aws_networkmanager_site_to_site_vpn_attachment.test2.attachment_type
}
`, rName, bgpASN, vpnIP2))
}

func testAccSiteToSiteVPNAttachmentDataSourceConfig_filters(rName string, bgpASN int, vpnIP, vpnIP2 string) string {
	return acctest.ConfigCompose(
		testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersBase(rName, bgpASN, vpnIP, vpnIP2), `
data "aws_networkmanager_site_to_site_vpn_attachment" "by_vpn_connection_arn" {
  core_network_id    = aws_networkmanager_site_to_site_vpn_attachment.test2.core_network_id
  vpn_connection_arn = aws_networkmanager_site_to_site_vpn_attachment.test2.vpn_connection_arn

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}

data "aws_networkmanager_site_to_site_vpn_attachment" "by_tags" {
  core_network_id = aws_networkmanager_site_to_site_vpn_attachment.test2.core_network_id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}

data "aws_networkmanager_site_to_site_vpn_attachment" "by_all_filters" {
  core_network_id = aws_networkmanager_site_to_site_vpn_attachment.test2.core_network_id
  edge_location   = data.aws_region.current.region
  state           = "AVAILABLE"

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`)
}

func testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersMultipleMatches(rName string, bgpASN int, vpnIP, vpnIP2 string) string {
	return acctest.ConfigCompose(
		testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersBase(rName, bgpASN, vpnIP, vpnIP2),
		fmt.Sprintf(`
data "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id = aws_networkmanager_site_to_site_vpn_attachment.test2.core_network_id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`, rName))
}

func testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersNoMatch(rName string, bgpASN int, vpnIP, vpnIP2 string) string {
	return acctest.ConfigCompose(
		testAccSiteToSiteVPNAttachmentDataSourceConfig_filtersBase(rName, bgpASN, vpnIP, vpnIP2), `
data "aws_networkmanager_site_to_site_vpn_attachment" "test" {
  core_network_id = aws_networkmanager_site_to_site_vpn_attachment.test2.core_network_id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test, aws_networkmanager_attachment_accepter.test2]
}
`)
}
