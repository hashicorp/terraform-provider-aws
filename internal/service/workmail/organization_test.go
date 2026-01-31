// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workmail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfworkmail "github.com/hashicorp/terraform-provider-aws/internal/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWorkMailOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organization workmail.DescribeOrganizationOutput
	rAlias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(names.WorkMailServiceID)) // service is lower case
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rAlias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rAlias),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "workmail", regexache.MustCompile(`organization/.`)),
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

func TestAccWorkMailOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var organization workmail.DescribeOrganizationOutput
	rAlias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workmail_organization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(names.WorkMailServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkMailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic(rAlias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkmail.ResourceOrganization, resourceName),
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

func testAccCheckOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workmail_organization" {
				continue
			}

			out, err := tfworkmail.FindOrganizationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.WorkMail, create.ErrActionCheckingDestroyed, tfworkmail.ResNameOrganization, rs.Primary.ID, err)
			}

			if aws.ToString(out.State) == tfworkmail.StatusDeleted {
				return nil
			}

			return create.Error(names.WorkMail, create.ErrActionCheckingDestroyed, tfworkmail.ResNameOrganization, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOrganizationExists(ctx context.Context, name string, organization *workmail.DescribeOrganizationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient(ctx)

		resp, err := tfworkmail.FindOrganizationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.WorkMail, create.ErrActionCheckingExistence, tfworkmail.ResNameOrganization, rs.Primary.ID, err)
		}

		*organization = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkMailClient(ctx)

	input := workmail.ListOrganizationsInput{}
	_, err := conn.ListOrganizations(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrganizationConfig_basic(rAlias string) string {
	return fmt.Sprintf(`
resource "aws_workmail_organization" "test" {
  alias = %[1]q
}
`, rAlias)
}
