// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworkPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	originalSegmentValue := "segmentValue1"
	expectedJSONOriginal := fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"64512-65534\"],\"dns-support\":true,\"edge-locations\":[{\"location\":\"%s\"}],\"security-group-referencing-support\":false,\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), originalSegmentValue)
	updatedSegmentValue := "segmentValue2"
	expectedJSONUpdated := fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"64512-65534\"],\"dns-support\":true,\"edge-locations\":[{\"location\":\"%s\"}],\"security-group-referencing-support\":false,\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), updatedSegmentValue)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_basic(originalSegmentValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, "aws_networkmanager_core_network.test", names.AttrID),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", expectedJSONOriginal),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", expectedJSONUpdated),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachment(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	segmentValue := "segmentValue"
	expectedJSON := fmt.Sprintf("{\"core-network-configuration\":{\"asn-ranges\":[\"64512-65534\"],\"dns-support\":true,\"edge-locations\":[{\"location\":\"%s\"}],\"security-group-referencing-support\":false,\"vpn-ecmp-support\":true},\"segments\":[{\"isolate-attachments\":false,\"name\":\"%s\",\"require-attachment-acceptance\":true}],\"version\":\"2021.12\"}", acctest.Region(), segmentValue)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// Step 1: Create core network, policy, and VPC attachment (no create-route yet)
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentStep1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.CoreNetworkStateAvailable)),
				),
			},
			{
				// Step 2: Update policy to add create-route with VPC attachment destination
				Config: testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentStep2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"action":"create-route"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"destinations":\["attachment-.+"\]`)),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_id", "aws_networkmanager_core_network.test", names.AttrID),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					acctest.CheckResourceAttrJSONNoDiff(resourceName, "policy_document", expectedJSON),
				),
			},
		},
	})
}

func TestAccNetworkManagerCoreNetworkPolicyAttachment_vpcAttachmentMultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var providers []*schema.Provider
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(fmt.Sprintf(`{"attachment-policies":\[{"action":{"association-method":"constant","segment":"segment"},"condition-logic":"or","conditions":\[{"operator":"equals","type":"resource-id","value":"attachment-.+"}\],"rule-number":1},{"action":{"association-method":"constant","segment":"segment2"},"condition-logic":"or","conditions":\[{"operator":"equals","type":"resource-id","value":"attachment-.+"}\],"rule-number":2}\],"core-network-configuration":{"asn-ranges":\["64512-65534"\],"dns-support":true,"edge-locations":\[{"location":"%s"},{"location":"%s"}\],"security-group-referencing-support":false,"vpn-ecmp-support":true},"segment-actions":\[{"action":"create-route","destination-cidr-blocks":\["10.0.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment"},{"action":"create-route","destination-cidr-blocks":\["10.1.0.0/16"\],"destinations":\["attachment-.+"\],"segment":"segment2"}\],"segments":\[{"isolate-attachments":false,"name":"segment","require-attachment-acceptance":false},{"isolate-attachments":false,"name":"segment2","require-attachment-acceptance":false}\],"version":"2021.12"}`, acctest.Region(), acctest.AlternateRegion()))),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPolicies(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"version":"2025.11"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policies":`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"testpolicy"`)),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_attachmentRoutingPolicyRules(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

// Routing Policies - All Condition Types
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesAllConditionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllConditionTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prefix-equals"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prefix-in-cidr"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"asn-in-as-path"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"community-in-list"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"med-equals"`)),
				),
			},
		},
	})
}

// Routing Policies - All Action Types
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesAllActionTypes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesAllActionTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"drop"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"allow"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"prepend-asn-list"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"set-med"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"set-local-preference"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"add-community"`)),
				),
			},
		},
	})
}

// Routing Policies - Condition Logic AND
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesConditionLogicAnd(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicAnd(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"and"`)),
				),
			},
		},
	})
}

// Routing Policies - Condition Logic OR
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesConditionLogicOr(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesConditionLogicOr(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"condition-logic":"or"`)),
				),
			},
		},
	})
}

// Routing Policies - Multiple Policies
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesMultiplePolicies(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesMultiplePolicies(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"inboundpolicy"`)),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-name":"outboundpolicy"`)),
				),
			},
		},
	})
}

// Routing Policies - Outbound Direction
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesOutbound(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesOutbound(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-direction":"outbound"`)),
				),
			},
		},
	})
}

// Routing Policies - With Description
func TestAccNetworkManagerCoreNetworkPolicyAttachment_routingPoliciesWithDescription(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networkmanager_core_network_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkPolicyAttachmentConfig_routingPoliciesWithDescription(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCoreNetworkPolicyAttachmentExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexache.MustCompile(`"routing-policy-description":"Filter and control inbound routes"`)),
				),
			},
		},
	})
}

// Routing Policies - Error Duplicate Name
func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorDuplicateRoutingPolicyName(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

// Routing Policies - Error Duplicate Rule Number
func TestAccNetworkManagerCoreNetworkPolicyAttachment_expectErrorDuplicateRuleNumber(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

func testAccCheckCoreNetworkPolicyAttachmentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkManagerClient(ctx)

		// pass in latestPolicyVersionID to get the latest version id by default
		const latestPolicyVersionID = -1
		_, err := tfnetworkmanager.FindCoreNetworkPolicyByTwoPartKey(ctx, conn, rs.Primary.ID, latestPolicyVersionID)

		return err
	}
}

func testAccCoreNetworkPolicyAttachmentConfig_basic(segmentValue string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["64512-65534"]

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

// Step 1: Base policy with attachment_policies (no create-route) to create VPC attachment first
func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentStep1() string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets("tf-acc-test-networkmanager-core-network-policy-attachment", 2), fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["64512-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type = "any"
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

resource "aws_networkmanager_vpc_attachment" "test" {
  core_network_id = aws_networkmanager_core_network_policy_attachment.test.core_network_id
  subnet_arns     = aws_subnet.test[*].arn
  vpc_arn         = aws_vpc.test.arn
}
`, acctest.Region()))
}

// Step 2: Update policy to add create-route with VPC attachment destination
func testAccCoreNetworkPolicyAttachmentConfig_vpcAttachmentStep2() string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets("tf-acc-test-networkmanager-core-network-policy-attachment", 2), fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["64512-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type = "any"
    }

    action {
      association_method = "constant"
      segment            = "segment"
    }
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
  global_network_id = aws_networkmanager_global_network.test.id
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
    asn_ranges = ["64512-65534"]

    edge_locations {
      location = %[1]q
    }

    edge_locations {
      location = %[2]q
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = false
  }

  segments {
    name                          = "segment2"
    require_attachment_acceptance = false
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

    conditions {
      type     = "resource-id"
      operator = "equals"
      value    = aws_networkmanager_vpc_attachment.test.id
    }

    action {
      association_method = "constant"
      segment            = "segment"
    }
  }

  attachment_policies {
    rule_number     = 2
    condition_logic = "or"

    conditions {
      type     = "resource-id"
      operator = "equals"
      value    = aws_networkmanager_vpc_attachment.alternate_region.id
    }

    action {
      association_method = "constant"
      segment            = "segment2"
    }
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

func testAccCoreNetworkPolicyAttachmentConfig_routingPolicies() string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
}

data "aws_networkmanager_core_network_policy_document" "test" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = %[1]q
    }
  }

  segments {
    name                          = "segment"
    require_attachment_acceptance = true
  }

  attachment_policies {
    rule_number     = 100
    condition_logic = "or"

    conditions {
      type = "tag-exists"
      key  = "segment"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "segment"
    }
  }

  routing_policies {
    routing_policy_name      = "testpolicy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        condition_logic = "and"

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
    name                          = "segment"
    require_attachment_acceptance = true
  }

  attachment_policies {
    rule_number     = 100
    condition_logic = "or"

    conditions {
      type = "tag-exists"
      key  = "segment"
    }

    action {
      association_method = "tag"
      tag_value_of_key   = "segment"
    }
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

  routing_policies {
    routing_policy_name      = "policy2"
    routing_policy_direction = "outbound"
    routing_policy_number    = 200

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          type  = "prefix-in-cidr"
          value = "192.168.0.0/16"
        }

        action {
          type = "drop"
        }
      }
    }
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
    routing_policy_name      = "testpolicy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        action {
          type = "allow"
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
    routing_policy_name      = "testallconditions"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 2
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-in-cidr"
          value = "192.168.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 3
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "asn-in-as-path"
          value = "64512"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 4
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "community-in-list"
          value = "65000:100"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 5
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "med-equals"
          value = "50"
        }
        action {
          type = "allow"
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
    routing_policy_name      = "testallactions"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "drop"
        }
      }
    }

    routing_policy_rules {
      rule_number = 2
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.1.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 3
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-in-cidr"
          value = "10.2.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 4
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.3.0.0/16"
        }
        action {
          type  = "prepend-asn-list"
          value = "65001,65002"
        }
      }
    }

    routing_policy_rules {
      rule_number = 5
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.4.0.0/16"
        }
        action {
          type  = "remove-asn-list"
          value = "65003"
        }
      }
    }

    routing_policy_rules {
      rule_number = 6
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.5.0.0/16"
        }
        action {
          type  = "replace-asn-list"
          value = "65004,65005"
        }
      }
    }

    routing_policy_rules {
      rule_number = 7
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.6.0.0/16"
        }
        action {
          type  = "add-community"
          value = "65000:200"
        }
      }
    }

    routing_policy_rules {
      rule_number = 8
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.7.0.0/16"
        }
        action {
          type  = "remove-community"
          value = "65000:100"
        }
      }
    }

    routing_policy_rules {
      rule_number = 9
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.8.0.0/16"
        }
        action {
          type  = "set-med"
          value = "100"
        }
      }
    }

    routing_policy_rules {
      rule_number = 10
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.9.0.0/16"
        }
        action {
          type  = "set-local-preference"
          value = "200"
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
    routing_policy_name      = "testandlogic"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        match_conditions {
          type  = "asn-in-as-path"
          value = "64512"
        }
        action {
          type = "allow"
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
    routing_policy_name      = "testorlogic"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "or"
        match_conditions {
          type  = "community-in-list"
          value = "65000:100"
        }
        match_conditions {
          type  = "med-equals"
          value = "50"
        }
        action {
          type = "allow"
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
    routing_policy_name      = "inboundpolicy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }
  }

  routing_policies {
    routing_policy_name      = "outboundpolicy"
    routing_policy_direction = "outbound"
    routing_policy_number    = 200

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        condition_logic = "and"
        match_conditions {
          type  = "prefix-in-cidr"
          value = "192.168.0.0/16"
        }
        action {
          type  = "summarize"
          value = "192.0.0.0/8"
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
    routing_policy_name      = "outboundtest"
    routing_policy_direction = "outbound"
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
    routing_policy_name        = "testwithdescription"
    routing_policy_description = "Filter and control inbound routes"
    routing_policy_direction   = "inbound"
    routing_policy_number      = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "allow"
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
    routing_policy_name      = "duplicatename"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }
  }

  routing_policies {
    routing_policy_name      = "duplicatename"
    routing_policy_direction = "outbound"
    routing_policy_number    = 200

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          type  = "prefix-equals"
          value = "192.168.0.0/16"
        }
        action {
          type = "allow"
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
    routing_policy_name      = "testpolicy"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          type  = "prefix-equals"
          value = "10.0.0.0/16"
        }
        action {
          type = "allow"
        }
      }
    }

    routing_policy_rules {
      rule_number = 1
      rule_definition {
        match_conditions {
          type  = "prefix-equals"
          value = "192.168.0.0/16"
        }
        action {
          type = "drop"
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
