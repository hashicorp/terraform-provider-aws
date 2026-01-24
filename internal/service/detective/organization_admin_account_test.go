// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_detective_organization_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
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
	resourceName := "aws_detective_organization_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdetective.ResourceOrganizationAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationAdminAccount_MultiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_detective_organization_admin_account.test"
	altResourceName := "aws_detective_organization_admin_account.alternate"
	thirdResourceName := "aws_detective_organization_admin_account.third"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DetectiveServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_multiRegion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, t, resourceName),
					testAccCheckOrganizationAdminAccountExists(ctx, t, altResourceName),
					testAccCheckOrganizationAdminAccountExists(ctx, t, thirdResourceName),
				),
			},
		},
	})
}

func testAccCheckOrganizationAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_organization_admin_account" {
				continue
			}

			_, err := tfdetective.FindOrganizationAdminAccountByAccountID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Detective Organization Admin Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationAdminAccountExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DetectiveClient(ctx)

		_, err := tfdetective.FindOrganizationAdminAccountByAccountID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccOrganizationAdminAccountConfig_self() string {
	return `
resource "aws_detective_organization_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}
`
}

func testAccOrganizationAdminAccountConfig_multiRegion() string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(3), `
resource "aws_detective_organization_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_detective_organization_admin_account" "alternate" {
  provider = awsalternate

  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_detective_organization_admin_account" "third" {
  provider = awsthird

  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "current" {}
`)
}
