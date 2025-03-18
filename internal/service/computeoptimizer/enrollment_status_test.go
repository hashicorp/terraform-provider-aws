// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomputeoptimizer "github.com/hashicorp/terraform-provider-aws/internal/service/computeoptimizer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnrollmentStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v computeoptimizer.GetEnrollmentStatusOutput
	resourceName := "aws_computeoptimizer_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_basic("Active"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "number_of_member_accounts_opted_in", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnrollmentStatusConfig_basic("Inactive"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "number_of_member_accounts_opted_in", "0"),
				),
			},
		},
	})
}

func testAccEnrollmentStatus_includeMemberAccounts(t *testing.T) {
	ctx := acctest.Context(t)
	var v computeoptimizer.GetEnrollmentStatusOutput
	resourceName := "aws_computeoptimizer_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_includeMemberAccounts("Active", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", acctest.CtTrue),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(resourceName, "number_of_member_accounts_opted_in", 0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnrollmentStatusConfig_includeMemberAccounts("Inactive", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", acctest.CtFalse),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(resourceName, "number_of_member_accounts_opted_in", 0),
				),
			},
		},
	})
}

func testAccCheckEnrollmentStatusExists(ctx context.Context, n string, v *computeoptimizer.GetEnrollmentStatusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

		output, err := tfcomputeoptimizer.FindEnrollmentStatus(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEnrollmentStatusConfig_basic(status string) string {
	return fmt.Sprintf(`
resource "aws_computeoptimizer_enrollment_status" "test" {
  status = %[1]q
}
`, status)
}

func testAccEnrollmentStatusConfig_includeMemberAccounts(status string, includeMemberAccounts bool) string {
	return fmt.Sprintf(`
resource "aws_computeoptimizer_enrollment_status" "test" {
  status = %[1]q

  include_member_accounts = %[2]t
}
`, status, includeMemberAccounts)
}
