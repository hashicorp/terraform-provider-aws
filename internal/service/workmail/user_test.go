// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workmail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkMailUser_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var user workmail.DescribeUserOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	userName := fmt.Sprintf("user%s", sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, userName, "Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "Test User"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, userName),
					resource.TestCheckResourceAttr(resourceName, "hidden_from_global_address_list", "false"),
					resource.TestCheckResourceAttr(resourceName, "role", "USER"),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organization_id",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

func TestAccWorkMailUser_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	userName := fmt.Sprintf("user%s", sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, userName, "Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "Test User"),
				),
			},
			{
				Config: testAccUserConfig_updated(rName, userName, "Updated User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, "Updated User"),
					resource.TestCheckResourceAttr(resourceName, "city", "Seattle"),
					resource.TestCheckResourceAttr(resourceName, "job_title", "Engineer"),
					resource.TestCheckResourceAttr(resourceName, "telephone", "+1-555-0100"),
				),
			},
		},
	})
}

func TestAccWorkMailUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var user workmail.DescribeUserOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	userName := fmt.Sprintf("user%s", sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, userName, "Test User"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserExists(ctx, t, resourceName, &user),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkmail.ResourceUser, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccUserImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"]), nil
	}
}

func testAccCheckUserDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_user" {
				continue
			}

			_, err := tfworkmail.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.WorkMail, create.ErrActionCheckingDestroyed, tfworkmail.ResNameUser, rs.Primary.ID, err)
			}

			return create.Error(names.WorkMail, create.ErrActionCheckingDestroyed, tfworkmail.ResNameUser, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, t *testing.T, name string, users ...*workmail.DescribeUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameUser, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		out, err := tfworkmail.FindUserByTwoPartKey(ctx, conn, rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["user_id"])
		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameUser, rs.Primary.ID, err)
		}

		if len(users) > 0 && users[0] != nil {
			*users[0] = *out
		}

		return nil
	}
}

func testAccUserConfig_basic(rName, userName, displayName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_user" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  name            = %[2]q
  display_name    = %[3]q
  password        = "TestTest1234!"
}
`, rName, userName, displayName)
}

func testAccUserConfig_updated(rName, userName, displayName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_user" "test" {
  organization_id                 = aws_workmail_organization.test.organization_id
  name                            = %[2]q
  display_name                    = %[3]q
  password                        = "TestTest1234!"
  city                            = "Seattle"
  hidden_from_global_address_list = true
  job_title                       = "Engineer"
  telephone                       = "+1-555-0100"
}
`, rName, userName, displayName)
}
