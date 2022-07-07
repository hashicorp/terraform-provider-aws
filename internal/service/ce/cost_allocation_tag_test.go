package ce_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostAllocationTag_basic(t *testing.T) {
	var output costexplorer.CostAllocationTag
	resourceName := "aws_ce_cost_allocation_tag.test"
	rName := "Tag01"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostAllocationTagConfig_basic(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "type", "UserDefined"),
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
					testAccCheckCostAllocationTagExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
					resource.TestCheckResourceAttr(resourceName, "type", "UserDefined"),
				),
			}, {
				Config: testAccCostAllocationTagConfig_basic(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostAllocationTagExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "tag_key", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "type", "UserDefined"),
				),
			},
		},
	})
}

func testAccCheckCostAllocationTagExists(resourceName string, output *costexplorer.CostAllocationTag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return names.Error(names.CE, names.ErrActionCheckingExistence, tfce.ResCostAllocationTag, resourceName, errors.New("not found in state"))
		}

		ctx := context.TODO()
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn
		costAllocTag, err := tfce.FindCostAllocationTagByKey(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*output = *costAllocTag

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
