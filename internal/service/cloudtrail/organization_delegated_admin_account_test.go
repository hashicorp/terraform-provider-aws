// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Prerequisites:
// * Organizations management account
// * Organization member account
func testAccOrganizationDelegatedAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudtrail_organization_delegated_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDelegatedAdminAccountConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDelegatedAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN), // nosemgrep:ci.semgrep.acctest.checks.arn-resourceattrset // TODO: need environment where this test can run
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEmail),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "service_principal", tfcloudtrail.ServicePrincipal),
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

func testAccOrganizationDelegatedAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudtrail_organization_delegated_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDelegatedAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationDelegatedAdminAccountExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudtrail.ResourceOrganizationDelegatedAdminAccount, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudtrail_organization_delegated_admin_account" {
				continue
			}

			_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.ID, tfcloudtrail.ServicePrincipal)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudTrail Organization Delegated Admin Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationDelegatedAdminAccountExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.ID, tfcloudtrail.ServicePrincipal)

		return err
	}
}

const testAccOrganizationDelegatedAdminAccountConfig_basic = `
data "aws_organizations_organization" "test" {}

resource "aws_cloudtrail_organization_delegated_admin_account" "test" {
  account_id = data.aws_organizations_organization.test.non_master_accounts[0].id
}
`
