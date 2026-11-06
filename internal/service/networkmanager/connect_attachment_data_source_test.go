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

func TestAccNetworkManagerConnectAttachmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_connect_attachment.test"
	resourceName := "aws_networkmanager_connect_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_policy_rule_number", resourceName, "attachment_policy_rule_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachment_type", resourceName, "attachment_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_arn", resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_id", resourceName, "core_network_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, "edge_location", resourceName, "edge_location"),
					resource.TestCheckResourceAttr(dataSourceName, "options.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options.0.protocol", resourceName, "options.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrProtocol, "GRE"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "segment_name", resourceName, "segment_name"),
					// The resource records the state at creation, before the attachment is accepted.
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "AVAILABLE"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transport_attachment_id", resourceName, "transport_attachment_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "updated_at"),
				),
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachmentDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourcePrefix := "data.aws_networkmanager_connect_attachment."
	greResourceName := "aws_networkmanager_connect_attachment.test"
	noEncapResourceName := "aws_networkmanager_connect_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentDataSourceConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_transport_no_encap", "attachment_id", noEncapResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_transport_no_encap", names.AttrProtocol, "NO_ENCAP"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_transport_no_encap", "options.0.protocol", "NO_ENCAP"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_transport_gre", "attachment_id", greResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_transport_gre", names.AttrProtocol, "GRE"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "attachment_id", noEncapResourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "tags.Filter", "second"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "attachment_id", noEncapResourceName, names.AttrID),
				),
			},
			{
				// The GRE and the Tunnel-less attachment share the transport attachment.
				Config:      testAccConnectAttachmentDataSourceConfig_filtersMultipleMatches(rName),
				ExpectError: regexache.MustCompile(`multiple\s+Network\s+Manager\s+Connect\s+Attachments\s+matched`),
			},
			{
				Config:      testAccConnectAttachmentDataSourceConfig_filtersNoMatch(rName),
				ExpectError: regexache.MustCompile(`no\s+matching\s+Network\s+Manager\s+Connect\s+Attachment\s+found`),
			},
		},
	})
}

func testAccConnectAttachmentDataSourceConfig_transportBase(rName string) string {
	return acctest.ConfigCompose(testAccConnectAttachmentConfig_base(rName), `
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    segment = "shared"
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpc_attachment.test.id
  attachment_type = aws_networkmanager_vpc_attachment.test.attachment_type
}
`)
}

func testAccConnectAttachmentDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConnectAttachmentDataSourceConfig_transportBase(rName),
		fmt.Sprintf(`
resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location

  options {
    protocol = "GRE"
  }

  tags = {
    segment = "shared"
    Name    = %[1]q
  }

  depends_on = [aws_networkmanager_attachment_accepter.test]
}

resource "aws_networkmanager_attachment_accepter" "connect" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}

data "aws_networkmanager_connect_attachment" "test" {
  attachment_id = aws_networkmanager_attachment_accepter.connect.attachment_id
}
`, rName))
}

func testAccConnectAttachmentDataSourceConfig_filtersBase(rName string) string {
	return acctest.ConfigCompose(
		testAccConnectAttachmentDataSourceConfig_transportBase(rName),
		fmt.Sprintf(`
resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location

  options {
    protocol = "GRE"
  }

  tags = {
    segment = "shared"
    Name    = %[1]q
    Filter  = "first"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test]
}

resource "aws_networkmanager_attachment_accepter" "connect" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}

resource "aws_networkmanager_connect_attachment" "test2" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location

  options {
    protocol = "NO_ENCAP"
  }

  tags = {
    segment = "shared"
    Name    = %[1]q
    Filter  = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.test]
}

resource "aws_networkmanager_attachment_accepter" "connect2" {
  attachment_id   = aws_networkmanager_connect_attachment.test2.id
  attachment_type = aws_networkmanager_connect_attachment.test2.attachment_type
}
`, rName))
}

func testAccConnectAttachmentDataSourceConfig_filters(rName string) string {
	return acctest.ConfigCompose(
		testAccConnectAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_connect_attachment" "by_transport_no_encap" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  protocol                = "NO_ENCAP"

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}

data "aws_networkmanager_connect_attachment" "by_transport_gre" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  protocol                = "GRE"

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}

data "aws_networkmanager_connect_attachment" "by_tags" {
  core_network_id = aws_networkmanager_core_network.test.id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}

data "aws_networkmanager_connect_attachment" "by_all_filters" {
  core_network_id         = aws_networkmanager_core_network.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  protocol                = "NO_ENCAP"
  state                   = "AVAILABLE"
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}
`)
}

func testAccConnectAttachmentDataSourceConfig_filtersMultipleMatches(rName string) string {
	return acctest.ConfigCompose(
		testAccConnectAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}
`)
}

func testAccConnectAttachmentDataSourceConfig_filtersNoMatch(rName string) string {
	return acctest.ConfigCompose(
		testAccConnectAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_connect_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_attachment_accepter.connect, aws_networkmanager_attachment_accepter.connect2]
}
`)
}
