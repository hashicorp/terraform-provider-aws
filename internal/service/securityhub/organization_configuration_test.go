// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "0"),
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
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "0"),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableStandards("DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "0"),
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
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "0"),
				),
			},
		},
	})
}

// CENTRAL configuration has unique set of constraints vs other SecurityHub organization_config:
//
// 1. Must be done from a *member* delegated admin account:
// "Central configuration couldn't be enabled because the organization management account is designated as the delegated Security Hub administrator account."
// "Designate a different account as the delegated administrator, and retry."
//
// To allow for this the following is a multi-account test:
// The primary provider is expected to be a member account and alternate provider a management account.
//
// 2. Dependencies on DelegatedAdmin and FindingAggregators invert (!) after central config is created.
// "Finding Aggregator must be created to enable Central Configuration"
// "You must [...] disable central configuration in order to remove or change your aggregation Region."
// "You must [...] disable central configuration in order to remove or change the delegated Security Hub administrator"
//
// Due to this API behaviour, this resource isn't very terraform friendly.
func TestAccOrganizationConfiguration_centralConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckAlternateRegionIs(t, acctest.Region())
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// Start with LOCAL e.g default behaviour
				// This allows us to create an finding_aggregator in test without breaking dependency flow on destroy
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(true, "DEFAULT", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.configuration_type", "LOCAL"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Enable CENTRAL configuration
			{
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(false, "NONE", "CENTRAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.configuration_type", "CENTRAL"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.status", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Go back to LOCAL; this allows us to destroy the delegated admin with the given dependency flow that is necessary for creates.
			{
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(true, "DEFAULT", "LOCAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.configuration_type", "LOCAL"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.0.status", "ENABLED"),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.WaitOrganizationConfigurationEnabled(ctx, conn, 2*time.Minute)

		return err
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

func testAccOrganizationConfigurationConfig_centralConfiguration(autoEnable bool, autoEnableStandards, configType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
data "aws_caller_identity" "member" {}

resource "aws_securityhub_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.member.account_id

  provider = awsalternate
}

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

  depends_on = [aws_securityhub_organization_admin_account.test]
}
`, autoEnable, autoEnableStandards, configType))
}
