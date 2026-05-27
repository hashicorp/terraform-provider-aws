// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalXSSMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.XssMatchSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
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

func TestAccWAFRegionalXSSMatchSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.XssMatchSet
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccXSSMatchSetConfig_changeName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.XssMatchSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfwafregional.ResourceXSSMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_changeTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.XssMatchSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccXSSMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"text_transformation":   "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "BODY",
						"text_transformation":   "CMD_LINE",
					}),
				),
			},
		},
	})
}

func TestAccWAFRegionalXSSMatchSet_noTuples(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.XssMatchSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafregional_xss_match_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckXSSMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccXSSMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, t, resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "0"),
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

func testAccCheckXSSMatchSetExists(ctx context.Context, t *testing.T, n string, v *awstypes.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No regional WAF XSS Match Set ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindXSSMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckXSSMatchSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_xss_match_set" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindXSSMatchSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional XSS Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccXSSMatchSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_changeName(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_changeTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuple {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, rName)
}

func testAccXSSMatchSetConfig_noTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q
}
`, rName)
}
