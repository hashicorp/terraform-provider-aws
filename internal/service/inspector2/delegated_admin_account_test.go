// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegatedAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_delegated_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegatedAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdminAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, "relationship_status"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					idx := -1
					for i, rs := range s {
						if v, ok := rs.Attributes["relationship_status"]; ok && v != "" {
							idx = i
						}
					}

					if idx == -1 {
						return fmt.Errorf("expected aws_inspector2_delegated_admin_account to be in state, not found")
					}

					rs := s[idx]

					if rs.Attributes["relationship_status"] != string(types.RelationshipStatusEnabled) {
						return fmt.Errorf("expected relationship_status attribute to be set and be %s, received: %s", string(types.RelationshipStatusEnabled), rs.Attributes["relationship_status"])
					}

					return nil
				},
			},
		},
	})
}

func testAccDelegatedAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_delegated_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegatedAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdminAccountExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfinspector2.ResourceDelegatedAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegatedAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_delegated_admin_account" {
				continue
			}

			_, err := tfinspector2.FindDelegatedAdminAccountByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector2 Delegated Admin Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDelegatedAdminAccountExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Inspector2Client(ctx)

		_, err := tfinspector2.FindDelegatedAdminAccountByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDelegatedAdminAccountConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}
`
}
