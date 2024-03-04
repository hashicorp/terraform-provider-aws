// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfigurationPolicyAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_configuration_policy_association.test"
	accountTarget := "data.aws_caller_identity.member.account_id"
	ouTarget := "aws_organizations_organizational_unit.test.id"
	rootTarget := "data.aws_organizations_organization.test.roots[0].id"
	policy1 := "aws_securityhub_configuration_policy.test_1.id"
	policy2 := "aws_securityhub_configuration_policy.test_2.id"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckAlternateRegionIs(t, acctest.Region())
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationPolicyAssociationConfig_base(ouTarget, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_base(ouTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
				),
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_base(rootTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "parent_id"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
				),
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_base(accountTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "data.aws_caller_identity.member", "account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
				),
			},
		},
	})
}

func testAccCheckConfigurationPolicyAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)
		out, err := conn.GetConfigurationPolicyAssociation(ctx, &securityhub.GetConfigurationPolicyAssociationInput{
			Target: tfsecurityhub.GetTarget(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if out.AssociationStatus == types.ConfigurationPolicyAssociationStatusFailed {
			return fmt.Errorf("unexpected association status: %s %s", out.AssociationStatus, *out.AssociationStatusMessage)
		}

		return nil
	}
}

const testAccOrganizationalUnitConfig_base = `
data "aws_organizations_organization" "test" {
	provider = awsalternate
}

resource "aws_organizations_organizational_unit" "test" {
	provider  = awsalternate

	name      = "testAccConfigurationPolicyOrgUnitConfig_base"
	parent_id = data.aws_organizations_organization.test.roots[0].id
}
`

const testAccConfigurationPoliciesConfig_base = `
resource "aws_securityhub_configuration_policy" "test_1" {
	name = "test1"
	security_hub_policy {
		service_enabled       = true
		enabled_standard_arns = ["arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"]
		security_controls_configuration {
			disabled_control_identifiers = []
		}
	}
	
	depends_on = [aws_securityhub_organization_configuration.test]
}

resource "aws_securityhub_configuration_policy" "test_2" {
	name = "test2"
	security_hub_policy {
		service_enabled       = true
		enabled_standard_arns = ["arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"]
		security_controls_configuration {
			enabled_control_identifiers = ["CloudTrail.1"]
		}
	}
	
	depends_on = [aws_securityhub_organization_configuration.test]
}
`

func testAccConfigurationPolicyAssociationConfig_base(targetID, policyID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		testAccOrganizationalUnitConfig_base,
		testAccCentralConfigurationEnabledConfig_base,
		testAccConfigurationPoliciesConfig_base,
		fmt.Sprintf(`
			resource "aws_securityhub_configuration_policy_association" "test" {
				target_id = %[1]s
				policy_id = %[2]s
			}
			`, targetID, policyID))
}
