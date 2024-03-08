// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
		CheckDestroy:             testAccCheckConfigurationPolicyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationPolicyAssociationConfig_basic(ouTarget, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_1", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_basic(ouTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "id"),
				),
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_basic(rootTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "aws_organizations_organizational_unit.test", "parent_id"),
				),
			},
			{
				Config: testAccConfigurationPolicyAssociationConfig_basic(accountTarget, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", "aws_securityhub_configuration_policy.test_2", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", "data.aws_caller_identity.member", "account_id"),
				),
			},
		},
	})
}

func testAccConfigurationPolicyAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_configuration_policy_association.test"
	ouTarget := "aws_organizations_organizational_unit.test.id"
	policy1 := "aws_securityhub_configuration_policy.test_1.id"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckAlternateRegionIs(t, acctest.Region())
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckConfigurationPolicyAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationPolicyAssociationConfig_basic(ouTarget, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationPolicyAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceConfigurationPolicyAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		_, err := tfsecurityhub.FindConfigurationPolicyAssociationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckConfigurationPolicyAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_configuration_policy_association" {
				continue
			}

			_, err := tfsecurityhub.FindConfigurationPolicyAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Configuration Policy Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccOrganizationalUnitConfig_base = `
data "aws_organizations_organization" "test" {
  provider = awsalternate
}

resource "aws_organizations_organizational_unit" "test" {
  provider = awsalternate

  name      = "testAccConfigurationPolicyOrgUnitConfig_base"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}
`

// lintignore:AWSAT005
const testAccConfigurationPoliciesConfig_base = `
resource "aws_securityhub_configuration_policy" "test_1" {
  name = "test1"

  configuration_policy {
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

  configuration_policy {
    service_enabled       = true
    enabled_standard_arns = ["arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"]

    security_controls_configuration {
      enabled_control_identifiers = ["CloudTrail.1"]
    }
  }

  depends_on = [aws_securityhub_organization_configuration.test]
}
`

func testAccConfigurationPolicyAssociationConfig_basic(targetID, policyID string) string {
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
