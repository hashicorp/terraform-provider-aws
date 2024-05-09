// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfinspector2 "github.com/hashicorp/terraform-provider-aws/internal/service/inspector2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMemberAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_member_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, "data.aws_caller_identity.member", names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "delegated_admin_account_id", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", string(types.RelationshipStatusEnabled)),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccMemberAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_inspector2_member_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Inspector2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckMemberAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMemberAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMemberAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfinspector2.ResourceMemberAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMemberAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Inspector2 Member Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		_, err := tfinspector2.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckMemberAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Inspector2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector2_member_association" {
				continue
			}

			_, err := tfinspector2.FindMemberByAccountID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector2 Member Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccMemberAssociationConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "current" {}

resource "aws_inspector2_delegated_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
}

data "aws_caller_identity" "member" {
  provider = "awsalternate"
}

resource "aws_inspector2_member_association" "test" {
  account_id = data.aws_caller_identity.member.account_id

  depends_on = [aws_inspector2_delegated_admin_account.test]
}
`)
}
