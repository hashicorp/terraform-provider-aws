// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerAttachmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_attachments.test"
	resourceName := "aws_networkmanager_vpc_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, "attachments.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.attachment_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.attachment_policy_rule_number", resourceName, "attachment_policy_rule_number"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.attachment_type", resourceName, "attachment_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.core_network_arn", resourceName, "core_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.core_network_id", resourceName, "core_network_id"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "attachments.0.created_at"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.edge_location", resourceName, "edge_location"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.owner_account_id", resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.resource_arn", resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.segment_name", resourceName, "segment_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attachments.0.state", resourceName, names.AttrState),
					resource.TestCheckResourceAttr(dataSourceName, "attachments.0.tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "attachments.0.tags.Name", rName),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "attachments.0.updated_at"),
				),
			},
		},
	})
}

func TestAccNetworkManagerAttachmentsDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourcePrefix := "data.aws_networkmanager_attachments."
	resource2Name := "aws_networkmanager_vpc_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttachmentsDataSourceConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourcePrefix+"all", "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"all", "attachments.#", "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_type", "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_other_type", "ids.#", "0"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_other_type", "attachments.#", "0"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_state", "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_other_state", "ids.#", "0"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_edge_location", "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "ids.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "ids.0", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "attachments.#", "1"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "attachments.0.tags.Filter", "second"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_other_tags", "ids.#", "0"),
				),
			},
		},
	})
}

func testAccAttachmentsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, false),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_attachments" "test" {
  core_network_id = aws_networkmanager_vpc_attachment.test.core_network_id
}
`, rName))
}

func testAccAttachmentsDataSourceConfig_filters(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, false),
		fmt.Sprintf(`
resource "aws_vpc" "test2" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test2.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test2.cidr_block, 8, 0)

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    Name   = %[1]q
    Filter = "first"
  }
}

resource "aws_networkmanager_vpc_attachment" "test2" {
  subnet_arns     = [aws_subnet.test2.arn]
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test2.arn

  tags = {
    Name   = %[1]q
    Filter = "second"
  }
}

locals {
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
}

data "aws_networkmanager_attachments" "all" {
  core_network_id = local.core_network_id

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_type" {
  core_network_id = local.core_network_id
  attachment_type = "VPC"

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_other_type" {
  core_network_id = local.core_network_id
  attachment_type = "SITE_TO_SITE_VPN"

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_state" {
  core_network_id = local.core_network_id
  state           = "AVAILABLE"

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_other_state" {
  core_network_id = local.core_network_id
  state           = "PENDING_ATTACHMENT_ACCEPTANCE"

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_edge_location" {
  core_network_id = local.core_network_id
  edge_location   = data.aws_region.current.region

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_tags" {
  core_network_id = local.core_network_id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}

data "aws_networkmanager_attachments" "by_other_tags" {
  core_network_id = local.core_network_id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_vpc_attachment.test, aws_networkmanager_vpc_attachment.test2]
}
`, rName))
}
