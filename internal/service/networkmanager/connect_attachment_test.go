// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerConnectAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "GRE"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_basic_NoDependsOn(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic_NoDependsOn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "GRE"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfnetworkmanager.ResourceConnectAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_protocolNoEncap(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_protocolNoEncap(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "CONNECT"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "options.0.protocol", "NO_ENCAP"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_routingPolicyLabel(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_routingPolicyLabel(rName, "testlabel"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", "testlabel"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccNetworkManagerConnectAttachment_routingPolicyLabelUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.ConnectAttachment
	resourceName := "aws_networkmanager_connect_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAttachmentConfig_routingPolicyLabel(rName, "labelv1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", "labelv1"),
				),
			},
			{
				Config: testAccConnectAttachmentConfig_routingPolicyLabel(rName, "labelv2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConnectAttachmentExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "routing_policy_label", "labelv2"),
					testAccCheckConnectAttachmentRecreated(&v1, &v2),
				),
			},
		},
	})
}

func testAccCheckConnectAttachmentRecreated(before, after *awstypes.ConnectAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.Attachment.AttachmentId == after.Attachment.AttachmentId {
			return fmt.Errorf("Connect Attachment was not recreated")
		}
		return nil
	}
}

func testAccCheckConnectAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConnectAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		output, err := tfnetworkmanager.FindConnectAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConnectAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkmanager_connect_attachment" {
				continue
			}

			_, err := tfnetworkmanager.FindConnectAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
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

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.region
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

func testAccConnectAttachmentConfig_baseWithRoutingPolicy(rName, label string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
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

data "aws_region" "current" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.region
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

  routing_policies {
    routing_policy_name      = "policy1"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          type  = "prefix-in-cidr"
          value = "10.0.0.0/8"
        }

        action {
          type = "allow"
        }
      }
    }
  }

  attachment_routing_policy_rules {
    rule_number = 1

    conditions {
      type  = "routing-policy-label"
      value = %[2]q
    }

    action {
      associate_routing_policies = ["policy1"]
    }
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
`, rName, label))
}

func testAccConnectAttachmentConfig_routingPolicyLabel(rName, label string) string {
	return acctest.ConfigCompose(testAccConnectAttachmentConfig_baseWithRoutingPolicy(rName, label), fmt.Sprintf(`
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
  routing_policy_label    = %[1]q
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
`, label))
}
