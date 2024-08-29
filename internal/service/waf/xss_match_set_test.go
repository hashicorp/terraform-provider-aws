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

func TestAccWAFXSSMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`xssmatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
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

func TestAccWAFXSSMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	xssMatchSetNewName := fmt.Sprintf("xssMatchSetNewName-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct2),
				),
			},
			{
				Config: testAccXSSMatchSetConfig_changeName(xssMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, xssMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct2),
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

func TestAccWAFXSSMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.XssMatchSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceXSSMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFXSSMatchSet_changeTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.XssMatchSet
	setName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccXSSMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"text_transformation":   "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      acctest.Ct1,
						"field_to_match.0.data": "",
						"field_to_match.0.type": "BODY",
						"text_transformation":   "CMD_LINE",
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

func TestAccWAFXSSMatchSet_noTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.XssMatchSet
	setName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", acctest.Ct0),
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

func testAccCheckXSSMatchSetExists(ctx context.Context, n string, v *awstypes.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindXSSMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckXSSMatchSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_xss_match_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindXSSMatchSetByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF XSS Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccXSSMatchSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccXSSMatchSetConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccXSSMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuples {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccXSSMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q
}
`, name)
}
