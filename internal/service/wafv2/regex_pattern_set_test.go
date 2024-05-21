// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2RegexPatternSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexPatternSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "one",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "two",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccRegexPatternSetConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "one",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "two",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "three",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRegexPatternSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexPatternSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceRegexPatternSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_minimal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexPatternSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNewName := fmt.Sprintf("regex-pattern-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexPatternSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", acctest.Ct2),
				),
			},
			{
				Config: testAccRegexPatternSetConfig_basic(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexPatternSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_oneTag(rName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRegexPatternSetImportStateIdFunc(resourceName),
			},
			{
				Config: testAccRegexPatternSetConfig_twoTags(rName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccRegexPatternSetConfig_oneTag(rName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func testAccCheckRegexPatternSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_regex_pattern_set" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

			_, err := tfwafv2.FindRegexPatternSetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 RegexPatternSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRegexPatternSetExists(ctx context.Context, n string, v *awstypes.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 RegexPatternSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		output, err := tfwafv2.FindRegexPatternSetByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

		if err != nil {
			return err
		}

		*v = *output.RegexPatternSet

		return nil
	}
}

func testAccRegexPatternSetConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }
}
`, name)
}

func testAccRegexPatternSetConfig_update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = %[1]q
  description = "Updated"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  regular_expression {
    regex_string = "three"
  }
}
`, name)
}

func testAccRegexPatternSetConfig_minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"
}
`, name)
}

func testAccRegexPatternSetConfig_oneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey, tagValue)
}

func testAccRegexPatternSetConfig_twoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRegexPatternSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope]), nil
	}
}
