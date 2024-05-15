// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostCategory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "effective_start"),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ce", regexache.MustCompile(`costcategory/.+$`)),
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

func TestAccCECostCategory_effectiveStart(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	firstDayOfMonth := func(v time.Time) string {
		return time.Date(v.Year(), v.Month(), 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	}
	now := time.Now()
	firstOfThisMonth := firstDayOfMonth(now)
	firstOfLastMonth := firstDayOfMonth(now.AddDate(0, -1, 0))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_effectiveStart(rName, firstOfThisMonth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "effective_start", firstOfThisMonth),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCostCategoryConfig_effectiveStart(rName, firstOfLastMonth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "effective_start", firstOfLastMonth),
				),
			},
		},
	})
}

func TestAccCECostCategory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfce.ResourceCostCategory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCECostCategory_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccCostCategoryConfig_operandAnd(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccCECostCategory_notWithAnd(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_operandNot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_splitCharges(rName, "PROPORTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccCostCategoryConfig_splitCharges(rName, "EVEN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccCECostCategory_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCostCategoryDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCostCategoryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCostCategoryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, resourceName, &output),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCostCategoryExists(ctx context.Context, n string, v *awstypes.CostCategory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		output, err := tfce.FindCostCategoryByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCostCategoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ce_cost_category" {
				continue
			}

			_, err := tfce.FindCostCategoryByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CE Cost Category %s still exists", rs.Primary.ID)
		}

		return nil
	}
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

func testAccCostCategoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCostCategoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCostCategoryConfig_effectiveStart(rName, date string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test" {
  name            = %[1]q
  rule_version    = "CostCategoryExpression.v1"
  effective_start = %[2]q
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
`, rName, date)
}

func testAccCostCategoryConfig_operandNot(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_cost_category" "test" {
  name         = %[1]q
  rule_version = "CostCategoryExpression.v1"
  rule {
    value = "notax"
    rule {
      and {
        dimension {
          key           = "LINKED_ACCOUNT_NAME"
          values        = ["-prod"]
          match_options = ["ENDS_WITH"]
        }
      }
      and {
        not {
          dimension {
            key           = "RECORD_TYPE"
            values        = ["Tax"]
            match_options = ["EQUALS"]
          }
        }
      }
    }
    type = "REGULAR"
  }
}
`, rName)
}
