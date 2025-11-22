// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworkPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	originalSegmentValue := "segmentValue1"
	updatedSegmentValue := "segmentValue2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(originalSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), originalSegmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(updatedSegmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), updatedSegmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachment(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	segmentValue := "segmentValue"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(fmt.Sprintf(`{"core-network-configuration":{"asn-ranges":\["65022-65534"\],"edge-locations":\[{"location":"%s"}\],"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["0.0.0.0/0"\],"destinations":\["attachment-.+"\],"segment":"segment"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":true}\],"version":"2021.12"}`, acctest.Region()))),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(segmentValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), segmentValue)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachmentMultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(fmt.Sprintf(`{"core-network-configuration":{"asn-ranges":\["65022-65534"\],"edge-locations":\[{"location":"%s"},{"location":"%s"}\],"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["10.0.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment"},{"action":"create-route","destination-cidr-blocks":\["10.1.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment2"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":true},{"isolate-attachments":false,"name":"segment2","require-attachment-acceptance":true}\],"version":"2021.12"}`, acctest.Region(), acctest.AlternateRegion()))),
					// use test below if the order of locations is unordered
					// resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"65022-65534\"],\"edge-locations\":[{\"location\":\"%s\"},{\"location\":\"%s\"}],\"vpn-ecmp-support\":true},\"segments\":[{\"description\":\"base-policy\",\"isolate-attachments\":false,\"name\":\"segment\",\"require-attachment-acceptance\":false}],\"version\":\"2021.12\"}", acctest.AlternateRegion(), acctest.Region())),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
			{
				Config:            testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectPolicyErrorInvalidASNRange(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_expectPolicyErrorInvalidASNRange(),
				ExpectError: regexache.MustCompile("CoreNetworkPolicyException: Incorrect policy"),
			},
		},
	})
}

func testAccCheckCoreNetworkPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager Core Network ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerClient(ctx)

		// pass in latestPolicyVersionId to get the latest version id by default
		const latestPolicyVersionId = -1
		_, err := tfnetworkmanager.FindCoreNetworkPolicyByTwoPartKey(ctx, conn, rs.Primary.ID, aws.Int32(latestPolicyVersionId))

		return err
	}
}

func testAccCoreNetworkPolicyAttachmentConfig_basic(segmentValue string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[2]q
    }
  }

  segments {
    name = %[1]q
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, segmentValue, acctest.Region())
}

func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentCreate() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "0.0.0.0/0"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.test.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id  = aws_networkmanager_global_network.test.id
  create_base_policy = true
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

resource "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.test[*].arn
  vpc_arn         = aws_vpc.test.arn
}
`, acctest.Region()))
}

func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentMultiRegionCreate() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }

    edge_locations {
      location = %[2]q
    }
  }

  segments {
    name = "segment"
  }

  segments {
    name = "segment2"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.0.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.test.id,
    ]
  }

  segment_actions {
    action  = "create-route"
    segment = "segment2"
    destination_cidr_blocks = [
      "10.1.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.alternate_region.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  base_policy_regions = [%[1]q, %[2]q]
  create_base_policy  = true
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

resource "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.test[*].arn
  vpc_arn         = aws_vpc.test.arn
}

# Alternate region
data "aws_availability_zones" "alternate_region_available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alternate_region" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_subnet" "alternate_region" {
  provider = "awsalternate"

  count = 2

  availability_zone = data.aws_availability_zones.alternate_region_available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.alternate_region.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.alternate_region.id

  tags = {
    Name = "tf-acc-test-networkmanager-core-network-policy-attachment"
  }
}

resource "aws_networkmanager_vpc_attachment" "alternate_region" {
  provider = "awsalternate"

  core_network_id = aws_networkmanager_core_network.test.id
  subnet_arns     = aws_subnet.alternate_region[*].arn
  vpc_arn         = aws_vpc.alternate_region.arn
}
`, acctest.Region(), acctest.AlternateRegion()))
}

func testAccCoreNetworkPolicyAttachmentConfig_expectPolicyErrorInvalidASNRange() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534123"] # not a valid range

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "test"
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPolicies(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"version":"2025.11"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policies":`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"test-policy"`)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
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

func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentRoutingPolicyRules(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentRoutingPolicyRules(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"version":"2025.11"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"attachment-routing-policy-rules":`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-label"`)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
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

func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorRoutingPoliciesWrongVersion(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesWrongVersion(),
				ExpectError: regexache.MustCompile(`routing_policies requires version 2025.11`),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorAttachmentRoutingPolicyRulesWrongVersion(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_attachmentRoutingPolicyRulesWrongVersion(),
				ExpectError: regexache.MustCompile(`attachment_routing_policy_rules requires version 2025.11`),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPolicies() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-policy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }

        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentRoutingPolicyRules() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  attachment_routing_policy_rules {
    rule_number = 1

    conditions {
      type  = "routing-policy-label"
      value = "production"
    }

    action {
      associate_routing_policies = ["policy1", "policy2"]
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesWrongVersion() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2021.12"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-policy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentRoutingPolicyRulesWrongVersion() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2021.12"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  attachment_routing_policy_rules {
    rule_number = 1

    conditions {
      type  = "routing-policy-label"
      value = "production"
    }

    action {
      associate_routing_policies = ["policy1"]
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - All Condition Types
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesAllConditionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllConditionTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prefix-equals"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prefix-in-cidr"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prefix-in-prefix-list"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"asn-in-as-path"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"community-in-list"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"med-equals"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllConditionTypes() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-all-conditions"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 2
      rule_definition {
        match_conditions {
          condition_type = "prefix-in-cidr"
          value          = "192.168.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 3
      rule_definition {
        match_conditions {
          condition_type = "prefix-in-prefix-list"
          value          = "pl-12345678"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 4
      rule_definition {
        match_conditions {
          condition_type = "asn-in-as-path"
          value          = "64512"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 5
      rule_definition {
        match_conditions {
          condition_type = "community-in-list"
          value          = "65000:100"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 6
      rule_definition {
        match_conditions {
          condition_type = "med-equals"
          value          = "50"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - All Action Types
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesAllActionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllActionTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"drop"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"allow"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"summarize"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prepend-asn-list"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"set-med"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"set-local-preference"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"add-community"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllActionTypes() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-all-actions"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "drop"
        }
      }
    }

    routing_policy_rules {
      rule_number = 2
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.1.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 3
      rule_definition {
        match_conditions {
          condition_type = "prefix-in-cidr"
          value          = "10.2.0.0/16"
        }
        actions {
          action_type = "summarize"
        }
      }
    }

    routing_policy_rules {
      rule_number = 4
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.3.0.0/16"
        }
        actions {
          action_type = "prepend-asn-list"
          value       = "65001,65002"
        }
      }
    }

    routing_policy_rules {
      rule_number = 5
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.4.0.0/16"
        }
        actions {
          action_type = "remove-asn-list"
          value       = "65003"
        }
      }
    }

    routing_policy_rules {
      rule_number = 6
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.5.0.0/16"
        }
        actions {
          action_type = "replace-asn-list"
          value       = "65004,65005"
        }
      }
    }

    routing_policy_rules {
      rule_number = 7
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.6.0.0/16"
        }
        actions {
          action_type = "add-community"
          value       = "65000:200"
        }
      }
    }

    routing_policy_rules {
      rule_number = 8
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.7.0.0/16"
        }
        actions {
          action_type = "remove-community"
          value       = "65000:100"
        }
      }
    }

    routing_policy_rules {
      rule_number = 9
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.8.0.0/16"
        }
        actions {
          action_type = "set-med"
          value       = "100"
        }
      }
    }

    routing_policy_rules {
      rule_number = 10
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.9.0.0/16"
        }
        actions {
          action_type = "set-local-preference"
          value       = "200"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Condition Logic AND
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesConditionLogicAnd(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicAnd(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"and"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicAnd() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-and-logic"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        match_conditions {
          condition_type = "asn-in-as-path"
          value          = "64512"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Condition Logic OR
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesConditionLogicOr(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicOr(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"or"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicOr() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-or-logic"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "or"
        match_conditions {
          condition_type = "community-in-list"
          value          = "65000:100"
        }
        match_conditions {
          condition_type = "med-equals"
          value          = "50"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Multiple Policies
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesMultiplePolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesMultiplePolicies(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"inbound-policy"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"outbound-policy"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesMultiplePolicies() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "inbound-policy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }

  routing_policies {
    routing_policy_name      = "outbound-policy"
    routing_policy_direction = "outbound"
    routing_policy_number    = 200

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-in-cidr"
          value          = "192.168.0.0/16"
        }
        actions {
          action_type = "summarize"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Outbound Direction
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesOutbound(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesOutbound(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-direction":"outbound"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesOutbound() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "outbound-test"
    routing_policy_direction = "outbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-in-prefix-list"
          value          = "pl-12345678"
        }
        actions {
          action_type = "summarize"
        }
        actions {
          action_type = "add-community"
          value       = "65000:200"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Multiple Actions
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesMultipleActions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesMultipleActions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"set-local-preference"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prepend-asn-list"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesMultipleActions() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "multiple-actions-test"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
        actions {
          action_type = "set-local-preference"
          value       = "200"
        }
        actions {
          action_type = "prepend-asn-list"
          value       = "65001,65002"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - With Description
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesWithDescription(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesWithDescription(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-description":"Filter and control inbound routes"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesWithDescription() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name        = "test-with-description"
    routing_policy_description = "Filter and control inbound routes"
    routing_policy_direction   = "inbound"
    routing_policy_number      = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Error Duplicate Name
func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorDuplicateRoutingPolicyName(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_duplicateRoutingPolicyName(),
				ExpectError: regexache.MustCompile(`duplicate routing_policy_name`),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_duplicateRoutingPolicyName() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "duplicate-name"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }

  routing_policies {
    routing_policy_name      = "duplicate-name"
    routing_policy_direction = "outbound"
    routing_policy_number    = 200

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "192.168.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Routing Policies - Error Duplicate Rule Number
func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorDuplicateRuleNumber(t *testing.T) {
	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccCoreNetworkPolicyAttachmentConfig_duplicateRuleNumber(),
				ExpectError: regexache.MustCompile(`duplicate rule_number`),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_duplicateRuleNumber() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "test-policy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "10.0.0.0/16"
        }
        actions {
          action_type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          condition_type = "prefix-equals"
          value          = "192.168.0.0/16"
        }
        actions {
          action_type = "drop"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Tag Value Condition
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesTagValue(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesTagValue(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"attachment-policies"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"tag-value"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesTagValue() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  segments {
    name = "development"
  }

  attachment_policies {
    rule_number = 100

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }

    action {
      association_method = "constant"
      segment            = "production"
    }
  }

  attachment_policies {
    rule_number = 200

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "development"
    }

    action {
      association_method = "constant"
      segment            = "development"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - All Condition Types
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesAllConditionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAllConditionTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"account"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"any"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"tag-value"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"tag-name"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"tag-exists"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"resource-id"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"region"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"attachment-type"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAllConditionTypes() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  attachment_policies {
    rule_number = 100
    conditions {
      type     = "account"
      operator = "equals"
      value    = "123456789012"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 200
    conditions {
      type = "any"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 300
    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 400
    conditions {
      type = "tag-name"
      key  = "CostCenter"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 500
    conditions {
      type = "tag-exists"
      key  = "Production"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 600
    conditions {
      type     = "resource-id"
      operator = "equals"
      value    = "vpc-12345678"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 700
    conditions {
      type     = "region"
      operator = "equals"
      value    = %[1]q
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 800
    conditions {
      type     = "attachment-type"
      operator = "equals"
      value    = "vpc"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Condition Logic
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesConditionLogic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesConditionLogic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"and"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"or"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesConditionLogic() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  segments {
    name = "development"
  }

  attachment_policies {
    rule_number     = 100
    condition_logic = "and"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }

    conditions {
      type     = "region"
      operator = "equals"
      value    = %[1]q
    }

    action {
      association_method = "constant"
      segment            = "production"
    }
  }

  attachment_policies {
    rule_number     = 200
    condition_logic = "or"

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "development"
    }

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "staging"
    }

    action {
      association_method = "constant"
      segment            = "development"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - All Operators
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesAllOperators(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAllOperators(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"operator":"equals"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"operator":"not-equals"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"operator":"contains"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"operator":"begins-with"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAllOperators() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "segment"
  }

  attachment_policies {
    rule_number = 100
    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 200
    conditions {
      type     = "tag-value"
      operator = "not-equals"
      key      = "Environment"
      value    = "test"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 300
    conditions {
      type     = "tag-value"
      operator = "contains"
      key      = "Name"
      value    = "prod"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number = 400
    conditions {
      type     = "tag-value"
      operator = "begins-with"
      key      = "Name"
      value    = "vpc-prod"
    }
    action {
      association_method = "constant"
      segment            = "segment"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Association Method Tag
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesAssociationMethodTag(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAssociationMethodTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"association-method":"tag"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"tag-value-of-key":"Segment"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAssociationMethodTag() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  segments {
    name = "development"
  }

  attachment_policies {
    rule_number = 100

    conditions {
      type = "tag-exists"
      key  = "Segment"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "Segment"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Association Method Constant
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesAssociationMethodConstant(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAssociationMethodConstant(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"association-method":"constant"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"segment":"production"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAssociationMethodConstant() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  attachment_policies {
    rule_number = 100

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }

    action {
      association_method = "constant"
      segment            = "production"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Require Acceptance
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesRequireAcceptance(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesRequireAcceptance(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"require-acceptance":true`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesRequireAcceptance() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  attachment_policies {
    rule_number = 100

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Environment"
      value    = "production"
    }

    action {
      association_method = "constant"
      segment            = "production"
      require_acceptance = true
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}

// Attachment Policies - Add To Network Function Group
func TestAccNetworkManagerCoreNetworkPolicyAttachment_attachmentPoliciesAddToNFG(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAddToNFG(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"add-to-network-function-group":"security-group"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"network-function-groups"`)),
				),
			},
		},
	})
}

func testAccCoreNetworkPolicyAttachmentConfig_attachmentPoliciesAddToNFG() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]
    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name = "production"
  }

  network_function_groups {
    name                        = "security-group"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number = 100

    conditions {
      type     = "tag-value"
      operator = "equals"
      key      = "Type"
      value    = "firewall"
    }

    action {
      add_to_network_function_group = "security-group"
    }
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}
`, acctest.Region())
}
