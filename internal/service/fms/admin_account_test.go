// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_fms_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
				),
			},
		},
	})
}

func testAccAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_fms_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					acctest.CheckSDKResourceDisappears(ctx, t, tffms.ResourceAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).FMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fms_admin_account" {
				continue
			}

			_, err := tffms.FindAdminAccount(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FMS Admin Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAdminAccountExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).FMSClient(ctx)

		_, err := tffms.FindAdminAccount(ctx, conn)

		return err
	}
}

const testAccAdminAccountConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`
