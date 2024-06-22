// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFSizeConstraintSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`sizeconstraintset/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    acctest.Ct1,
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

func TestAccWAFSizeConstraintSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	sizeConstraintSetNewName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct1),
				),
			},
			{
				Config: testAccSizeConstraintSetConfig_changeName(sizeConstraintSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, sizeConstraintSetNewName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct1),
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

func TestAccWAFSizeConstraintSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceSizeConstraintSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFSizeConstraintSet_changeConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    acctest.Ct1,
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
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "GE",
						"field_to_match.#":    acctest.Ct1,
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

func TestAccWAFSizeConstraintSet_noConstraints(t *testing.T) {
	ctx := acctest.Context(t)
	var contraints awstypes.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSizeConstraintSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSizeConstraintSetConfig_nos(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSizeConstraintSetExists(ctx, resourceName, &contraints),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", acctest.Ct0),
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

func testAccCheckSizeConstraintSetExists(ctx context.Context, n string, v *awstypes.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindSizeConstraintSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSizeConstraintSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_size_contraint_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindSizeConstraintSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Size Constraint Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSizeConstraintSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
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
resource "aws_waf_size_constraint_set" "size_constraint_set" {
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
resource "aws_waf_size_constraint_set" "size_constraint_set" {
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
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = %[1]q
}
`, name)
}
