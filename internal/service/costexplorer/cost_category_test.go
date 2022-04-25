package costexplorer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcostexplorer "github.com/hashicorp/terraform-provider-aws/internal/service/costexplorer"
)

func TestAccCostExplorerCostCategory_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostExplorerCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerCostCategoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCostExplorerCostCategory_complete(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostExplorerCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerCostCategoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccCostExplorerCostCategoryOperandAndConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCostExplorerCostCategory_splitCharge(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostExplorerCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerCostCategorySplitChargesConfig(rName, "PROPORTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccCostExplorerCostCategorySplitChargesConfig(rName, "EVEN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccCostExplorerCostCategory_disappears(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_costexplorer_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostExplorerCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostExplorerCostCategoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostExplorerCostCategoryExists(resourceName, &output),
					acctest.CheckResourceDisappears(acctest.Provider, tfcostexplorer.ResourceCostExplorerCostCategory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCostExplorerCostCategoryExists(resourceName string, output *costexplorer.CostCategory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn
		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if err != nil {
			return fmt.Errorf("problem checking for CE Cost Category  existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("CE Cost Category  %q does not exist", rs.Primary.ID)
		}

		*output = *resp.CostCategory

		return nil
	}
}

func testAccCheckCostExplorerCostCategoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_costexplorer_cost_category" {
			continue
		}

		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking CE Cost Category  was destroyed: %w", err)
		}

		if resp != nil && resp.CostCategory != nil {
			return fmt.Errorf("CE Cost Category  %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCostExplorerCostCategoryConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_costexplorer_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "staging"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-stg"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "testing"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-dev"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
}
`, name)
}

func testAccCostExplorerCostCategoryOperandAndConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_costexplorer_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-prod"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-stg"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-dev"]
          match_options = ["ENDS_WITH"]
        }
      }
    }
    type = "REGULAR"
  }
}
`, name)
}

func testAccCostExplorerCostCategorySplitChargesConfig(name, method string) string {
	return fmt.Sprintf(`
resource "aws_costexplorer_cost_category" "test1" {
  name         = "%[1]s-1"
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "staging"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-stg"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  rule {
    value = "testing"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-dev"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
}

resource "aws_costexplorer_cost_category" "test2" {
  name         = "%[1]s-2"
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-prod"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-stg"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-dev"]
          match_options = ["ENDS_WITH"]
        }
      }
    }
    type = "REGULAR"
  }
}

resource "aws_costexplorer_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "production"
    rule {
      dimension {
        key           = "LINKED_ACCOUNT_NAME"
        values        = ["-prod"]
        match_options = ["ENDS_WITH"]
      }
    }
    type = "REGULAR"
  }
  split_charge_rule {
    method  = %[2]q
    source  = aws_costexplorer_cost_category.test1.id
    targets = [aws_costexplorer_cost_category.test2.id]
  }
}
`, name, method)
}
