// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostAllocationTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostAllocationTag
	resourceName := "aws_ce_cost_allocation_tag.test"
	rName := "Tag01"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostAllocationTagDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostAllocationTagConfig_basic(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UserDefined"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCostAllocationTagConfig_basic(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UserDefined"),
				),
			}, {
				Config: testAccCostAllocationTagConfig_basic(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UserDefined"),
				),
			},
		},
	})
}

func TestAccCECostAllocationTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostAllocationTag
	resourceName := "aws_ce_cost_allocation_tag.test"
	rName := "Tag02"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostAllocationTagDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostAllocationTagConfig_basic(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(ctx, resourceName, &output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfce.ResourceCostAllocationTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCostAllocationTagExists(ctx context.Context, n string, v *awstypes.CostAllocationTag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		output, err := tfce.FindCostAllocationTagByTagKey(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCostAllocationTagDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ce_cost_allocation_tag" {
				continue
			}

			output, err := tfce.FindCostAllocationTagByTagKey(ctx, conn, rs.Primary.ID)

			if err != nil {
				return err
			}

			if output.Status == awstypes.CostAllocationTagStatusInactive {
				continue
			}

			return fmt.Errorf("Cost Explorer Anomaly Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCostAllocationTagConfig_basic(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_allocation_tag" "test" {
  tag_key = %[1]q
  status  = %[2]q
}
`, rName, status)
}
