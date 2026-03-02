// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccWAFRegionalRegexMatchSet_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccRegexMatchSet_basic,
		"changePatterns":     testAccRegexMatchSet_changePatterns,
		"noPatterns":         testAccRegexMatchSet_noPatterns,
		acctest.CtDisappears: testAccRegexMatchSet_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccRegexMatchSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var matchSet awstypes.RegexMatchSet
	var patternSet awstypes.RegexPatternSet
	var idx int

	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	fieldToMatch := awstypes.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: awstypes.MatchFieldType("HEADER"),
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig_basic(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexMatchSetExists(ctx, t, resourceName, &matchSet),
					testAccCheckRegexPatternSetExists(ctx, t, "aws_wafregional_regex_pattern_set.test", &patternSet),
					computeRegexMatchSetTuple(&patternSet, &fieldToMatch, "NONE", &idx),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
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

func testAccRegexMatchSet_changePatterns(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RegexMatchSet
	var patternSet awstypes.RegexPatternSet
	var idx1, idx2 int

	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig_basic(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegexMatchSetExists(ctx, t, resourceName, &before),
					testAccCheckRegexPatternSetExists(ctx, t, "aws_wafregional_regex_pattern_set.test", &patternSet),
					computeRegexMatchSetTuple(&patternSet, &awstypes.FieldToMatch{Data: aws.String("User-Agent"), Type: awstypes.MatchFieldType("HEADER")}, "NONE", &idx1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccRegexMatchSetConfig_changePatterns(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegexMatchSetExists(ctx, t, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),

					computeRegexMatchSetTuple(&patternSet, &awstypes.FieldToMatch{Data: aws.String("Referer"), Type: awstypes.MatchFieldType("HEADER")}, "COMPRESS_WHITE_SPACE", &idx2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "COMPRESS_WHITE_SPACE",
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

func testAccRegexMatchSet_noPatterns(t *testing.T) {
	ctx := acctest.Context(t)
	var matchSet awstypes.RegexMatchSet
	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegexMatchSetExists(ctx, t, resourceName, &matchSet),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "0"),
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

func testAccRegexMatchSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var matchSet awstypes.RegexMatchSet
	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegexMatchSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig_basic(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexMatchSetExists(ctx, t, resourceName, &matchSet),
					acctest.CheckSDKResourceDisappears(ctx, t, tfwafregional.ResourceRegexMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRegexMatchSetExists(ctx context.Context, t *testing.T, n string, v *awstypes.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindRegexMatchSetByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRegexMatchSetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_regex_match_set" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindRegexMatchSetByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Regex Match Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRegexMatchSetConfig_basic(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = %[1]q

  regex_match_tuple {
    field_to_match {
      data = "User-Agent"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_wafregional_regex_pattern_set.test.id
    text_transformation  = "NONE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = %[2]q
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccRegexMatchSetConfig_changePatterns(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = %[1]q

  regex_match_tuple {
    field_to_match {
      data = "Referer"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_wafregional_regex_pattern_set.test.id
    text_transformation  = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = %[2]q
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = %[1]q
}
`, matchSetName)
}

func computeRegexMatchSetTuple(patternSet *awstypes.RegexPatternSet, fieldToMatch *awstypes.FieldToMatch, textTransformation string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := map[string]any{
			"field_to_match":       tfwafregional.FlattenFieldToMatch(fieldToMatch),
			"regex_pattern_set_id": *patternSet.RegexPatternSetId,
			"text_transformation":  textTransformation,
		}

		*idx = tfwafregional.RegexMatchSetTupleHash(m)

		return nil
	}
}
