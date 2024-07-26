// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_fms_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
				),
			},
		},
	})
}

func testAccAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_fms_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffms.ResourceAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fms_admin_account" {
				continue
			}

			_, err := tffms.FindAdminAccount(ctx, conn)

			if tfresource.NotFound(err) {
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

func testAccAdminAccountExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FMSClient(ctx)

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
