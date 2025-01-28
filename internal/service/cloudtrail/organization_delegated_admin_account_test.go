// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Prerequisites:
// * Organizations management account
// * Organization member account
func testAccOrganizationDelegatedAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudtrail_organization_delegated_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDelegatedAdminAccountConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDelegatedAdminAccountExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationDelegatedAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationDelegatedAdminAccountExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudtrail.ResourceOrganizationDelegatedAdminAccount, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationDelegatedAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudtrail_organization_delegated_admin_account" {
				continue
			}

			_, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, rs.Primary.ID, tfcloudtrail.ServicePrincipal)

			if tfresource.NotFound(err) {
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

func testAccCheckOrganizationDelegatedAdminAccountExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

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
