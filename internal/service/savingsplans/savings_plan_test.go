// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/savingsplans/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsavingsplans "github.com/hashicorp/terraform-provider-aws/internal/service/savingsplans"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Note: These tests are commented out because running them would create
// actual Savings Plans with real financial commitments that cannot be cancelled.
// Use these as templates for manual testing only.

func TestAccSavingsPlansSavingsPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var savingsPlan awstypes.SavingsPlan
	resourceName := "aws_savingsplans_savings_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSavingsPlanDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.SavingsPlansServiceID),
		Steps: []resource.TestStep{
			{
				Config:      testAccSavingsPlanConfig_basic(),
				ExpectError: regexache.MustCompile(`Offering ID not found`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSavingsPlanExists(ctx, resourceName, &savingsPlan),
				),
			},
		},
	})
}

func testAccCheckSavingsPlanExists(ctx context.Context, n string, v *awstypes.SavingsPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SavingsPlansClient(ctx)

		output, err := tfsavingsplans.FindSavingsPlanByID(ctx, conn, rs.Primary.Attributes["savings_plan_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSavingsPlanDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SavingsPlansClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_savingsplans_savings_plan" {
				continue
			}

			_, err := tfsavingsplans.FindSavingsPlanByID(ctx, conn, rs.Primary.Attributes["savings_plan_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Savings Plan %s still exists", rs.Primary.Attributes["savings_plan_id"])
		}

		return nil
	}
}

func testAccSavingsPlanConfig_basic() string {
	return `
# Note: You need to provide a valid savings_plan_offering_id
# Use the aws_savingsplans_offerings data source to find valid offerings
resource "aws_savingsplans_savings_plan" "test" {
  savings_plan_offering_id = "00000000-0000-0000-0000-000000000000"
  commitment               = "1.0"
}
`
}
