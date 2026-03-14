// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
)

func testAccOrganizationAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_macie2_organization_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx, t),
		ErrorCheck:               testAccErrorCheckSkipOrganizationAdminAccount(t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "admin_account_id"),
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
	resourceName := "aws_macie2_organization_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx, t),
		ErrorCheck:               testAccErrorCheckSkipOrganizationAdminAccount(t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccErrorCheckSkipOrganizationAdminAccount(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"AccessDeniedException: The request failed because you must be a user of the management account for your AWS organization to perform this operation",
	)
}

func testAccCheckOrganizationAdminAccountExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).Macie2Client(ctx)

		adminAccount, err := tfmacie2.GetOrganizationAdminAccount(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("macie OrganizationAdminAccount (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_organization_admin_account" {
				continue
			}

			adminAccount, err := tfmacie2.GetOrganizationAdminAccount(ctx, conn, rs.Primary.ID)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) ||
				errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Macie is not enabled") {
				continue
			}

			if adminAccount == nil {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("macie OrganizationAdminAccount %q still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationAdminAccountConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["macie.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_macie2_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id
  depends_on       = [aws_macie2_account.test, aws_organizations_organization.test]
}
`
}
