// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
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
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "false"),
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
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnableStandards("DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
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
					testAccOrganizationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_standards", "NONE"),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn(ctx)

		_, err := conn.DescribeOrganizationConfigurationWithContext(ctx, &securityhub.DescribeOrganizationConfigurationInput{})

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
