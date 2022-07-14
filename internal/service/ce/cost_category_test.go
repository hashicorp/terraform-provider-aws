package ce_test

import (
	"errors"
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
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostCategory_basic(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
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

func TestAccCECostCategory_complete(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccCostCategoryConfig_operandAnd(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
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

func TestAccCECostCategory_splitCharge(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_splitCharges(rName, "PROPORTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccCostCategoryConfig_splitCharges(rName, "EVEN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
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

func TestAccCECostCategory_disappears(t *testing.T) {
	var output costexplorer.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCostCategoryDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(resourceName, &output),
					acctest.CheckResourceDisappears(acctest.Provider, tfce.ResourceCostCategory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCostCategoryExists(resourceName string, output *costexplorer.CostCategory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return names.Error(names.CE, names.ErrActionCheckingExistence, tfce.ResCostCategory, resourceName, errors.New("not found in state"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn
		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if err != nil {
			return names.Error(names.CE, names.ErrActionCheckingExistence, tfce.ResCostCategory, rs.Primary.ID, err)
		}

		if resp == nil {
			return names.Error(names.CE, names.ErrActionCheckingExistence, tfce.ResCostCategory, rs.Primary.ID, errors.New("not found"))
		}

		*output = *resp.CostCategory

		return nil
	}
}

func testAccCheckCostCategoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_cost_category" {
			continue
		}

		resp, err := conn.DescribeCostCategoryDefinition(&costexplorer.DescribeCostCategoryDefinitionInput{CostCategoryArn: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, costexplorer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResCostCategory, rs.Primary.ID, err)
		}

		if resp != nil && resp.CostCategory != nil {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResCostCategory, rs.Primary.ID, errors.New("still exists"))
		}
	}

	return nil

}

func testAccCostCategoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test" {
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
`, rName)
}

func testAccCostCategoryConfig_operandAnd(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test" {
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
`, rName)
}

func testAccCostCategoryConfig_splitCharges(rName, method string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test1" {
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

resource "aws_ce_cost_category" "test2" {
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

resource "aws_ce_cost_category" "test" {
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
    source  = aws_ce_cost_category.test1.id
    targets = [aws_ce_cost_category.test2.id]
  }
}
`, rName, method)
}
