// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerConnectAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "GRE"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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

func TestAccNetworkManagerConnectAttachment_basic_NoDependsOn(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic_NoDependsOn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "GRE"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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

func TestAccNetworkManagerConnectAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkmanager.ResourceConnectAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_protocolNoEncap(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_protocolNoEncap(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "NO_ENCAP"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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

func TestAccNetworkManagerConnectAttachment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkmanager.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
				),
			},
			{
				Config: testAccConnectAttachmentConfig_tags2(rName, "segment", "shared", "Name", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
				),
			},
			{
				Config: testAccConnectAttachmentConfig_tags1(rName, "segment", "shared"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
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

func testAccCheckConnectAttachmentExists(ctx context.Context, n string, v *networkmanager.ConnectAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Connect Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		output, err := tfnetworkmanager.FindConnectAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_connect_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindConnectAttachmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Manager Connect Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConnectAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.name
      asn      = 64512
    }
  }
  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = true
  }
  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
    share_with = ["*"]
  }
  attachment_policies {
    rule_number     = 1
    condition_logic = "or"
    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "segment"
      value    = "shared"
    }
    action {
      association_method = "constant"
      segment            = "shared"
    }
  }
}

`, rName))
}

func testAccConnectAttachmentConfig_basic(rName string) string {
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

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = "GRE"
  }
  tags = {
    segment = "shared"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`)
}

func testAccConnectAttachmentConfig_basic_NoDependsOn(rName string) string {
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

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = "GRE"
  }
  tags = {
    segment = "shared"
  }
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`)
}

func testAccConnectAttachmentConfig_protocolNoEncap(rName string) string {
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

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = "NO_ENCAP"
  }
  tags = {
    segment = "shared"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`)
}

func testAccConnectAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccConnectAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
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

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = "GRE"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`, tagKey1, tagValue1))
}

func testAccConnectAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccConnectAttachmentConfig_base(rName), fmt.Sprintf(`
resource "aws_networkmanager_vpc_attachment" "test" {
  subnet_arns     = [aws_subnet.test[0].arn]
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

resource "aws_networkmanager_connect_attachment" "test" {
  core_network_id         = aws_networkmanager_core_network.test.id
  transport_attachment_id = aws_networkmanager_vpc_attachment.test.id
  edge_location           = aws_networkmanager_vpc_attachment.test.edge_location
  options {
    protocol = "GRE"
  }
  depends_on = [
    "aws_networkmanager_attachment_accepter.test"
  ]
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test2" {
  attachment_id   = aws_networkmanager_connect_attachment.test.id
  attachment_type = aws_networkmanager_connect_attachment.test.attachment_type
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
