// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalByteMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, byteMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", acctest.Ct2),
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

func TestAccWAFRegionalByteMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	byteMatchSetNewName := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, byteMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", acctest.Ct2),
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
				Config: testAccByteMatchSetConfig_changeName(byteMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, byteMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", acctest.Ct2),
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

func TestAccWAFRegionalByteMatchSet_changeByteMatchTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, byteMatchSetName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrName, byteMatchSetName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"positional_constraint": "EXACTLY",
						"target_string":         "GET",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"positional_constraint": "ENDS_WITH",
						"target_string":         "badrefer2+",
						"text_transformation":   "LOWERCASE",
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

func TestAccWAFRegionalByteMatchSet_noByteMatchTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var byteMatchSet awstypes.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_noDescriptors(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &byteMatchSet),
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

func TestAccWAFRegionalByteMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckByteMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccByteMatchSetConfig_basic(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceByteMatchSet(), resourceName),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindByteMatchSetByID(ctx, conn, rs.Primary.ID)

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
			if rs.Type != "aws_wafregional_byte_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindByteMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Byte Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccByteMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
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
resource "aws_wafregional_byte_match_set" "byte_set" {
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

func testAccByteMatchSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_match_set" {
  name = %[1]q
}
`, name)
}

func testAccByteMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = %[1]q

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "GET"
    positional_constraint = "EXACTLY"

    field_to_match {
      type = "METHOD"
    }
  }

  byte_match_tuples {
    text_transformation   = "LOWERCASE"
    target_string         = "badrefer2+"
    positional_constraint = "ENDS_WITH"

    field_to_match {
      type = "URI"
    }
  }
}
`, name)
}
