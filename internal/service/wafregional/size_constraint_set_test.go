// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalSizeConstraintSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var constraints awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_size_constraint_set.size_constraint_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &constraints),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sizeConstraintSet),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						names.AttrSize:        "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data":         "",
						names.AttrType: "BODY",
					}),
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

func TestAccWAFRegionalSizeConstraintSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	sizeConstraintSetNewName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_size_constraint_set.size_constraint_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sizeConstraintSet),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "1"),
				),
			},
			{
				Config: testAccSizeConstraintSetConfig_changeName(sizeConstraintSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, sizeConstraintSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "1"),
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

func TestAccWAFRegionalSizeConstraintSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var constraints awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_size_constraint_set.size_constraint_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &constraints),
					acctest.CheckSDKResourceDisappears(ctx, t, tfwafregional.ResourceSizeConstraintSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalSizeConstraintSet_changeConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_size_constraint_set.size_constraint_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						names.AttrSize:        "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data":         "",
						names.AttrType: "BODY",
					}),
				),
			},
			{
				Config: testAccSizeConstraintSetConfig_changes(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "GE",
						"field_to_match.#":    "1",
						names.AttrSize:        "1024",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data":         "",
						names.AttrType: "BODY",
					}),
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

func TestAccWAFRegionalSizeConstraintSet_noConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var constraints awstypes.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_size_constraint_set.size_constraint_set"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_nos(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, t, resourceName, &constraints),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(
						resourceName, "size_constraints.#", "0"),
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

func testAccCheckSizeConstraintSetExists(ctx context.Context, t *testing.T, n string, v *awstypes.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SizeConstraintSet ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindSizeConstraintSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSizeConstraintSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_size_contraint_set" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindSizeConstraintSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Size Constraint Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSizeConstraintSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = %[1]q

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = %[1]q

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetConfig_changes(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = %[1]q

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "GE"
    size                = "1024"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccSizeConstraintSetConfig_nos(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = %[1]q
}
`, name)
}
