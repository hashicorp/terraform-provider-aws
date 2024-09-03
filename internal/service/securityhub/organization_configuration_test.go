// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_autoEnableStandards(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableStandards("DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableStandards("NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_centralConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccOrganizationConfigurationConfig_centralConfigurationInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(false, "NONE", "CENTRAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.configuration_type", "CENTRAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(true, "DEFAULT", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.configuration_type", "LOCAL"),
				),
			},
		},
	})
}

func testAccCheckOrganizationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindOrganizationConfiguration(ctx, conn)

		return err
	}
}

func testAccCheckOrganizationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_standards_subscription" {
				continue
			}

			output, err := tfsecurityhub.FindOrganizationConfiguration(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// LOCAL ConfigurationType => deleted.
			if output.OrganizationConfiguration.ConfigurationType == types.OrganizationConfigurationConfigurationTypeLocal {
				continue
			}

			return fmt.Errorf("Security Hub Standards Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccOrganizationConfigurationConfig_base = `
resource "aws_securityhub_account" "test" {}

data "aws_caller_identity" "current" {}

resource "aws_securityhub_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id

  depends_on = [aws_securityhub_account.test]
}
`

func testAccOrganizationConfigurationConfig_basic(autoEnable bool) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_securityhub_organization_configuration" "test" {
  auto_enable = %[1]t

  depends_on = [aws_securityhub_organization_admin_account.test]
}
`, autoEnable))
}

func testAccOrganizationConfigurationConfig_autoEnableStandards(autoEnableStandards string) string {
	return acctest.ConfigCompose(testAccOrganizationConfigurationConfig_base, fmt.Sprintf(`
resource "aws_securityhub_organization_configuration" "test" {
  auto_enable           = true
  auto_enable_standards = %[1]q

  depends_on = [aws_securityhub_organization_admin_account.test]
}
`, autoEnableStandards))
}

// Central configuration can only be enabled from a *member* delegated admin account.
// The primary provider is expected to be an organizations member account and the alternate provider is expected to be the organizations management account.
const testAccMemberAccountDelegatedAdminConfig_base = `
data "aws_caller_identity" "member" {}

resource "aws_securityhub_organization_admin_account" "test" {
  provider = awsalternate

  admin_account_id = data.aws_caller_identity.member.account_id
}
`

// Initialize all the providers used by acceptance tests.
var testAccOrganizationConfigurationConfig_centralConfigurationInit = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "member" {}

data "aws_caller_identity" "management" {
  provider = awsalternate
}
`)

func testAccOrganizationConfigurationConfig_centralConfiguration(autoEnable bool, autoEnableStandards, configType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccMemberAccountDelegatedAdminConfig_base,
		fmt.Sprintf(`
resource "aws_securityhub_finding_aggregator" "test" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_organization_admin_account.test]
}

resource "aws_securityhub_organization_configuration" "test" {
  auto_enable           = %[1]t
  auto_enable_standards = %[2]q
  organization_configuration {
    configuration_type = %[3]q
  }

  depends_on = [aws_securityhub_finding_aggregator.test]
}
`, autoEnable, autoEnableStandards, configType))
}
