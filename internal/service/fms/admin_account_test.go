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

func testAccAdminAccount_adminScope_basic(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.all_regions_enabled", "true"),
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

func testAccAdminAccount_adminScope_update(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.all_regions_enabled", "true"),
				),
			},
			{
				Config: testAccAdminAccountConfig_adminScope_regionScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.all_regions_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.regions.#", "2"),
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

func testAccAdminAccount_adminScope_regionScope(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_regionScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.all_regions_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.region_scope.0.regions.#", "2"),
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

func testAccAdminAccount_adminScope_policyTypeScope(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_policyTypeScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.policy_type_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.policy_type_scope.0.all_policy_types_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.policy_type_scope.0.policy_types.#", "2"),
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

func testAccAdminAccount_adminScope_accountScope(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_accountScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.account_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.account_scope.0.all_accounts_enabled", "true"),
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

func testAccAdminAccount_adminScope_ouScope(t *testing.T) {
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
				Config: testAccAdminAccountConfig_adminScope_ouScope,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.organizational_unit_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_scope.0.organizational_unit_scope.0.all_organizational_units_enabled", "true"),
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

const testAccAdminAccountConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`

const testAccAdminAccountConfig_adminScope_basic = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id

  admin_scope {
    region_scope {
      all_regions_enabled = true
    }
  }
}
`

const testAccAdminAccountConfig_adminScope_regionScope = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id

  admin_scope {
    region_scope {
      regions             = ["us-east-1", "us-west-2"]
      all_regions_enabled = false
    }
  }
}
`

const testAccAdminAccountConfig_adminScope_policyTypeScope = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id

  admin_scope {
    policy_type_scope {
      policy_types              = ["WAF", "WAFV2"]
      all_policy_types_enabled  = false
    }
  }
}
`

const testAccAdminAccountConfig_adminScope_accountScope = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id

  admin_scope {
    account_scope {
      all_accounts_enabled = true
    }
  }
}
`

const testAccAdminAccountConfig_adminScope_ouScope = `
data "aws_caller_identity" "current" {}

resource "aws_fms_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id

  admin_scope {
    organizational_unit_scope {
      all_organizational_units_enabled = true
    }
  }
}
`
