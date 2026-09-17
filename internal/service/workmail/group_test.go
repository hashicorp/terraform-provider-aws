// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkMailGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var group workmail.DescribeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := fmt.Sprintf("group%s", acctest.RandStringFromCharSet(t, 8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEmail),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "hidden_from_global_address_list", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGroupImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organization_id",
			},
		},
	})
}

func TestAccWorkMailGroup_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := fmt.Sprintf("group%s", acctest.RandStringFromCharSet(t, 8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "hidden_from_global_address_list", acctest.CtFalse),
				),
			},
			{
				Config: testAccGroupConfig_updated(rName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "hidden_from_global_address_list", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccWorkMailGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var group workmail.DescribeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	groupName := fmt.Sprintf("group%s", acctest.RandStringFromCharSet(t, 8, "abcdefghijklmnopqrstuvwxyz0123456789"))
	resourceName := "aws_workmail_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupExists(ctx, t, resourceName, &group),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkmail.ResourceGroup, resourceName),
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

func testAccGroupImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["group_id"]), nil
	}
}

func testAccCheckGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_group" {
				continue
			}

			_, err := tfworkmail.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["group_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Workmail group %s still exists", rs.Primary.Attributes["group_id"])
		}

		return nil
	}
}

func testAccCheckGroupExists(ctx context.Context, t *testing.T, name string, groups ...*workmail.DescribeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		out, err := tfworkmail.FindGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["organization_id"], rs.Primary.Attributes["group_id"])
		if err != nil {
			return err
		}

		if len(groups) > 0 && groups[0] != nil {
			*groups[0] = *out
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_group" "test" {
  organization_id = aws_workmail_organization.test.organization_id
  email           = "%[2]s@${aws_workmail_organization.test.default_mail_domain}"
  name            = %[2]q
}
`, rName, groupName)
}

func testAccGroupConfig_updated(rName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}

resource "aws_workmail_group" "test" {
  organization_id                 = aws_workmail_organization.test.organization_id
  email                           = "%[2]s@${aws_workmail_organization.test.default_mail_domain}"
  hidden_from_global_address_list = true
  name                            = %[2]q
}
`, rName, groupName)
}
