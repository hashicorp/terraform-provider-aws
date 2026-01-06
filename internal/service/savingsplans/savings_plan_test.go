// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package savingsplans_test

/*
import (
	"context"
	"fmt"
	"testing"

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

func TestAccSavingsPlansPlan_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var savingsPlan awstypes.SavingsPlan
	resourceName := "aws_savingsplans_plan.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSavingsPlanDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.SavingsPlansServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccSavingsPlanConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSavingsPlanExists(ctx, resourceName, &savingsPlan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
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

func testAccCheckSavingsPlanExists(ctx context.Context, n string, v *awstypes.SavingsPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SavingsPlansClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		output, err := tfsavingsplans.FindSavingsPlanByID(ctx, conn, rs.Primary.ID)

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
			if rs.Type != "aws_savingsplans_plan" {
				continue
			}

			output, err := tfsavingsplans.FindSavingsPlanByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Savings Plans in active state cannot be destroyed, only queued ones can
			if output.State != awstypes.SavingsPlanStateQueuedDeleted {
				return fmt.Errorf("Savings Plan %s still exists with state %s", rs.Primary.ID, output.State)
			}
		}

		return nil
	}
}

func testAccSavingsPlanConfig_basic() string {
	return `
# Note: You need to provide a valid savings_plan_offering_id
# Use the aws_savingsplans_offerings data source to find valid offerings
resource "aws_savingsplans_plan" "test" {
  savings_plan_offering_id = "example-offering-id"
  commitment               = "1.0"
}
`
}
*/
