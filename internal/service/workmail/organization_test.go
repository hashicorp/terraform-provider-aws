// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
)

func TestAccWorkMailOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organization workmail.DescribeOrganizationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, t, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "organization_alias", rName),
					resource.TestCheckResourceAttrSet(resourceName, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceName, "default_mail_domain"),
					resource.TestCheckResourceAttrSet(resourceName, "directory_id"),
					resource.TestCheckResourceAttrSet(resourceName, "directory_type"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "Active"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "workmail", regexache.MustCompile(`organization/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "organization_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organization_id",
				ImportStateVerifyIgnore:              []string{"delete_directory"},
			},
		},
	})
}

func TestAccWorkMailOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organization workmail.DescribeOrganizationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, t, resourceName, &organization),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfworkmail.ResourceOrganization, resourceName),
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

func TestAccWorkMailOrganization_deleteDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var before, after workmail.DescribeOrganizationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Step 1: create without delete_directory
				Config: testAccOrganizationConfig_noDeleteDirectory(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, t, resourceName, &before),
					resource.TestCheckNoResourceAttr(resourceName, "delete_directory"),
				),
			},
			{
				// Step 2: add delete_directory = true — must NOT recreate
				Config: testAccOrganizationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, t, resourceName, &after),
					testAccCheckOrganizationNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "delete_directory", acctest.CtTrue),
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

func TestAccWorkMailOrganization_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organization workmail.DescribeOrganizationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WorkMail)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_kmsKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, t, resourceName, &organization),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "Active"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "workmail", regexache.MustCompile(`organization/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "organization_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organization_id",
				ImportStateVerifyIgnore:              []string{"delete_directory", names.AttrKMSKeyARN},
			},
		},
	})
}

func testAccCheckOrganizationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_organization" {
				continue
			}

			out, err := tfworkmail.FindOrganizationByID(ctx, conn, rs.Primary.Attributes["organization_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Workmail returns the organization with State "Deleted" instead of 404
			if aws.ToString(out.State) == "Deleted" {
				continue
			}

			return fmt.Errorf("Workmail Organization %s still exists", rs.Primary.Attributes["organization_id"])
		}

		return nil
	}
}

func testAccCheckOrganizationExists(ctx context.Context, t *testing.T, name string, organization *workmail.DescribeOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not found"))
		}

		id := rs.Primary.Attributes["organization_id"]

		if id == "" {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).WorkMailClient(ctx)

		resp, err := tfworkmail.FindOrganizationByID(ctx, conn, id)
		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, rs.Primary.ID, err)
		}

		*organization = *resp

		return nil
	}
}

func testAccCheckOrganizationNotRecreated(before, after *workmail.DescribeOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.OrganizationId) != aws.ToString(after.OrganizationId) {
			return create.Error(names.WorkMail, create.ErrActionCheckingNotRecreated, tfworkmail.ResNameOrganization, aws.ToString(before.OrganizationId), errors.New("recreated"))
		}
		return nil
	}
}

func testAccOrganizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  delete_directory   = true
}
`, rName)
}

func testAccOrganizationConfig_noDeleteDirectory(rName string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
}
`, rName)
}

func testAccOrganizationConfig_kmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_workmail_organization" "test" {
  organization_alias = %[1]q
  kms_key_arn        = aws_kms_key.test.arn
  delete_directory   = true
}
`, rName)
}
