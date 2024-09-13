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

func testAccOrganizationAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "admin_account_id"),
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

func testAccOrganizationAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceOrganizationAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationAdminAccount_MultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_organization_admin_account.test"
	altResourceName := "aws_securityhub_organization_admin_account.alternate"
	thirdResourceName := "aws_securityhub_organization_admin_account.third"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_multiRegion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, resourceName),
					testAccCheckOrganizationAdminAccountExists(ctx, altResourceName),
					testAccCheckOrganizationAdminAccountExists(ctx, thirdResourceName),
				),
			},
		},
	})
}

func testAccCheckOrganizationAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_organization_admin_account" {
				continue
			}

			_, err := tfsecurityhub.FindAdminAccountByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Organization Admin Account (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationAdminAccountExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubClient(ctx)

		_, err := tfsecurityhub.FindAdminAccountByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccOrganizationAdminAccountConfig_self() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`
}

func testAccOrganizationAdminAccountConfig_multiRegion() string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["securityhub.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_securityhub_organization_admin_account" "alternate" {
  provider = awsalternate

  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_securityhub_organization_admin_account" "third" {
  provider = awsthird

  depends_on = [aws_organizations_organization.test]

  admin_account_id = data.aws_caller_identity.current.account_id
}
`)
}
