// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

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

func TestAccWAFByteMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSet),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer2",
						"text_transformation":   "NONE",
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

func TestAccWAFByteMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	byteMatchSetNewName := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSet),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct2),
				),
			},
			{
				Config: testAccByteMatchSetConfig_changeName(byteMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct2),
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

func TestAccWAFByteMatchSet_changeTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer2",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccByteMatchSetConfig_changeTuples(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"positional_constraint": "CONTAINS_WORD",
						"target_string":         "blah",
						"text_transformation":   "URL_DECODE",
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

func TestAccWAFByteMatchSet_noTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var byteSet awstypes.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_noTuples(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &byteSet),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", acctest.Ct0),
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

func TestAccWAFByteMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceByteMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckByteMatchSetExists(ctx context.Context, n string, v *awstypes.ByteMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindByteMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckByteMatchSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_byte_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindByteMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF ByteMatchSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccByteMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = %[1]q

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer2"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}
`, name)
}

func testAccByteMatchSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = %[1]q

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer2"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}
`, name)
}

func testAccByteMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = %[1]q

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "URL_DECODE"
    target_string         = "blah"
    positional_constraint = "CONTAINS_WORD"

    field_to_match {
      type = "METHOD"
      # data field omitted as the type is neither "HEADER" nor "SINGLE_QUERY_ARG"
    }
  }
}
`, name)
}

func testAccByteMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = %[1]q
}
`, name)
}
