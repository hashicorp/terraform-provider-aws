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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminRegion_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:       testAccSSOAdminRegion_basic,
		acctest.CtDisappears:  testAccSSOAdminRegion_disappears,
		"Identity":            testAccSSOAdminRegion_identitySerial,
		"ListBasic":           testAccSSOAdminRegion_listBasic,
		"ListIncludeResource": testAccSSOAdminRegion_listIncludeResource,
		"ListRegionOverride":  testAccSSOAdminRegion_listRegionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSSOAdminRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_region.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_arn"),
					resource.TestCheckResourceAttr(resourceName, "region_name", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.RegionStatusActive)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, intflex.ResourceIdSeparator, "instance_arn", "region_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "instance_arn",
			},
		},
	})
}

func testAccSSOAdminRegion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ssoadmin_region.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionExists(ctx, t, resourceName),
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

			_, err := tfssoadmin.FindRegionByTwoPartKey(ctx, conn, rs.Primary.Attributes["instance_arn"], rs.Primary.Attributes["region_name"])
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameRegion, rs.Primary.Attributes["instance_arn"], err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameRegion, rs.Primary.Attributes["instance_arn"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRegionExists(ctx context.Context, t *testing.T, n string, _ ...*ssoadmin.DescribeRegionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameRegion, n, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindRegionByTwoPartKey(ctx, conn, rs.Primary.Attributes["instance_arn"], rs.Primary.Attributes["region_name"])
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameRegion, rs.Primary.Attributes["instance_arn"], err)
		}

		return nil
	}
}

func testAccRegionConfig_basic() string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_region" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  region_name  = %[1]q
}
`, acctest.AlternateRegion())
}
