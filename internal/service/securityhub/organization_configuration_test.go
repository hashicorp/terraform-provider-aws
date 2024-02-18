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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
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
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
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
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableStandards("DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
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
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "0"),
				),
			},
		},
	})
}

func testAccOrganizationConfiguration_centralConfiguration(t *testing.T) {
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
		CheckDestroy:             testAccCheckOrganizationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_centralConfiguration(false, "NONE", "CENTRAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "organization_configuration.#", "1"),
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

		_, err := conn.DescribeOrganizationConfiguration(ctx, &securityhub.DescribeOrganizationConfigurationInput{})
		return err
	}
}

func testAccCheckOrganizationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_organization_configuration" {
				continue
			}

			out, err := conn.DescribeOrganizationConfiguration(ctx, &securityhub.DescribeOrganizationConfigurationInput{})
			if err != nil {
				return err
			}

			if out != nil && out.OrganizationConfiguration != nil && out.OrganizationConfiguration.ConfigurationType == types.OrganizationConfigurationConfigurationTypeCentral {
				return fmt.Errorf("Security Hub Organization Configuration (%s) still exists", rs.Primary.ID)
			}
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
