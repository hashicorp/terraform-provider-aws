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

func TestAccNetworkManagerVPCAttachmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_vpc_attachment.test"
	resourceName := "aws_networkmanager_vpc_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentDataSourceConfig_basic(rName),
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
					resource.TestCheckResourceAttrPair(dataSourceName, "options.0.appliance_mode_support", resourceName, "options.0.appliance_mode_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options.0.dns_support", resourceName, "options.0.dns_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options.0.ipv6_support", resourceName, "options.0.ipv6_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, "options.0.security_group_referencing_support", resourceName, "options.0.security_group_referencing_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerAccountID, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrResourceARN, resourceName, names.AttrResourceARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "segment_name", resourceName, "segment_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_arns.#", resourceName, "subnet_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_arn", resourceName, "vpc_arn"),
				),
			},
		},
	})
}

func TestAccNetworkManagerVPCAttachmentDataSource_filters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourcePrefix := "data.aws_networkmanager_vpc_attachment."
	resource2Name := "aws_networkmanager_vpc_attachment.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVPCAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAttachmentDataSourceConfig_filters(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_vpc_arn", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_vpc_arn", "vpc_arn", resource2Name, "vpc_arn"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_tags", "attachment_id", resource2Name, names.AttrID),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(dataSourcePrefix+"by_tags", "tags.Filter", "second"),
					resource.TestCheckResourceAttrPair(dataSourcePrefix+"by_all_filters", "attachment_id", resource2Name, names.AttrID),
				),
			},
			{
				Config:      testAccVPCAttachmentDataSourceConfig_filtersMultipleMatches(rName),
				ExpectError: regexache.MustCompile(`multiple\s+Network\s+Manager\s+VPC\s+Attachments\s+matched`),
			},
			{
				Config:      testAccVPCAttachmentDataSourceConfig_filtersNoMatch(rName),
				ExpectError: regexache.MustCompile(`no\s+matching\s+Network\s+Manager\s+VPC\s+Attachment\s+found`),
			},
		},
	})
}

func testAccVPCAttachmentDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentConfig_base(rName, false),
		fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = aws_subnet.test[*].arn
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test.arn

  tags = {
    key1 = %[1]q
  }
}

data "aws_networkmanager_vpc_attachment" "test" {
  attachment_id = aws_networkmanager_vpc_attachment.test.id
}
`, rName))
}

func testAccVPCAttachmentDataSourceConfig_filtersBase(rName string) string {
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
    Extra  = "a"
  }
}

resource "aws_networkmanager_vpc_attachment" "test2" {
  subnet_arns     = [aws_subnet.test2.arn]
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  vpc_arn         = aws_vpc.test2.arn

  tags = {
    Name   = %[1]q
    Filter = "second"
    Extra  = "b"
  }
}
`, rName))
}

func testAccVPCAttachmentDataSourceConfig_filters(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_vpc_attachment" "by_vpc_arn" {
  core_network_id = aws_networkmanager_vpc_attachment.test2.core_network_id
  vpc_arn         = aws_networkmanager_vpc_attachment.test2.vpc_arn

  depends_on = [aws_networkmanager_vpc_attachment.test]
}

data "aws_networkmanager_vpc_attachment" "by_tags" {
  core_network_id = aws_networkmanager_vpc_attachment.test2.core_network_id

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_vpc_attachment.test]
}

data "aws_networkmanager_vpc_attachment" "by_all_filters" {
  core_network_id = aws_networkmanager_vpc_attachment.test2.core_network_id
  edge_location   = data.aws_region.current.region
  state           = "AVAILABLE"

  tags = {
    Filter = "second"
  }

  depends_on = [aws_networkmanager_vpc_attachment.test]
}
`)
}

func testAccVPCAttachmentDataSourceConfig_filtersMultipleMatches(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentDataSourceConfig_filtersBase(rName),
		fmt.Sprintf(`
data "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_vpc_attachment.test2.core_network_id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_networkmanager_vpc_attachment.test]
}
`, rName))
}

func testAccVPCAttachmentDataSourceConfig_filtersNoMatch(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCAttachmentDataSourceConfig_filtersBase(rName), `
data "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_vpc_attachment.test2.core_network_id

  tags = {
    Filter = "third"
  }

  depends_on = [aws_networkmanager_vpc_attachment.test]
}
`)
}
