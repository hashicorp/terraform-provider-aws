// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/inspector2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegatedAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_delegated_admin_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegatedAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdminAccountExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegatedAdminAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegatedAdminAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegatedAdminAccountExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceDelegatedAdminAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegatedAdminAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_delegated_admin_account" {
				continue
			}

			st, _, err := tfinspector2.FindDelegatedAdminAccountStatusID(ctx, conn, rs.Primary.ID)

			if st == "" && errs.Contains(err, "admin account not found") {
				return nil
			}

			if err != nil {
				return err
			}

			return create.Error(names.Inspector2, create.ErrActionCheckingDestroyed, tfinspector2.ResNameDelegatedAdminAccount, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDelegatedAdminAccountExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameDelegatedAdminAccount, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameDelegatedAdminAccount, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		_, _, err := tfinspector2.FindDelegatedAdminAccountStatusID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.Inspector2, create.ErrActionCheckingExistence, tfinspector2.ResNameDelegatedAdminAccount, rs.Primary.ID, err)
		}

		return nil
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
