// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminRegion_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSSOAdminRegion_basic,
		acctest.CtDisappears: testAccSSOAdminRegion_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSSOAdminRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var region ssoadmin.DescribeRegionOutput
	resourceName := "aws_ssoadmin_region.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, t, resourceName, &region),
					resource.TestCheckResourceAttr(resourceName, "region_name", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "is_primary_region", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.RegionStatusActive)),
					resource.TestCheckResourceAttrSet(resourceName, "added_date"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
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

func testAccSSOAdminRegion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var region ssoadmin.DescribeRegionOutput
	resourceName := "aws_ssoadmin_region.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckMultipleRegion(t, 2)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, t, resourceName, &region),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceRegion, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckRegionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_region" {
				continue
			}

			_, err := tfssoadmin.FindRegionByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameRegion, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameRegion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRegionExists(ctx context.Context, t *testing.T, name string, v *ssoadmin.DescribeRegionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameRegion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameRegion, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		output, err := tfssoadmin.FindRegionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameRegion, rs.Primary.ID, err)
		}

		*v = *output

		return nil
	}
}

func testAccRegionConfig_basic() string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_region" "test" {
  instance_arn = one(data.aws_ssoadmin_instances.test.arns)
  region_name  = %[1]q
}
`, acctest.AlternateRegion())
}
