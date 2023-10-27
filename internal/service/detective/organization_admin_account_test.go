// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdetective "github.com/hashicorp/terraform-provider-aws/internal/service/detective"
)

func testAccOrganizationAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_detective_organization_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, detective.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, detective.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountConfig_self(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdetective.ResourceOrganizationAdminAccount(), resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, detective.EndpointsID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_detective_organization_admin_account" {
				continue
			}

			adminAccount, err := tfdetective.FindAdminAccount(ctx, conn, rs.Primary.ID)

			// Because of this resource's dependency, the Organizations organization
			// will be deleted first, resulting in the following valid error
			if tfawserr.ErrMessageContains(err, detective.ErrCodeValidationException, "account is not a member of an organization") {
				continue
			}

			if err != nil {
				return err
			}

			if adminAccount == nil {
				continue
			}

			return fmt.Errorf("expected Detective Organization Admin Account (%s) to be removed", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).DetectiveConn(ctx)

		adminAccount, err := tfdetective.FindAdminAccount(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if adminAccount == nil {
			return fmt.Errorf("Detective Organization Admin Account (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationAdminAccountConfig_self() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["detective.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  account_id = data.aws_caller_identity.current.account_id
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
  aws_service_access_principals = ["detective.${data.aws_partition.current.dns_suffix}"]
  feature_set                   = "ALL"
}

resource "aws_detective_graph" "test" {}

resource "aws_detective_organization_admin_account" "test" {
  depends_on = [aws_organizations_organization.test]

  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_detective_organization_admin_account" "alternate" {
  provider = awsalternate

  depends_on = [aws_organizations_organization.test]

  account_id = data.aws_caller_identity.current.account_id
}

resource "aws_detective_organization_admin_account" "third" {
  provider = awsthird

  depends_on = [aws_organizations_organization.test]

  account_id = data.aws_caller_identity.current.account_id
}
`)
}
