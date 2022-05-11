package waf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccWAFRegexMatchSet_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccRegexMatchSet_basic,
		"changePatterns": testAccRegexMatchSet_changePatterns,
		"noPatterns":     testAccRegexMatchSet_noPatterns,
		"disappears":     testAccRegexMatchSet_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccRegexMatchSet_basic(t *testing.T) {
	var matchSet waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx int

	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	fieldToMatch := waf.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: aws.String("HEADER"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`regexmatchset/.+`)),
					computeRegexMatchSetTuple(&patternSet, &fieldToMatch, "NONE", &idx),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
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
	var before, after waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx1, idx2 int

	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegexMatchSetExists(resourceName, &before),
					testAccCheckRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
					computeRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("User-Agent"), Type: aws.String("HEADER")}, "NONE", &idx1),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
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
					testAccCheckRegexMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),

					computeRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("Referer"), Type: aws.String("HEADER")}, "COMPRESS_WHITE_SPACE", &idx2),
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
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegexMatchSetExists(resourceName, &matchSet),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
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
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckRegexMatchSetDisappears(&matchSet),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func computeRegexMatchSetTuple(patternSet *waf.RegexPatternSet, fieldToMatch *waf.FieldToMatch, textTransformation string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := map[string]interface{}{
			"field_to_match":       tfwaf.FlattenFieldToMatch(fieldToMatch),
			"regex_pattern_set_id": *patternSet.RegexPatternSetId,
			"text_transformation":  textTransformation,
		}

		*idx = tfwaf.RegexMatchSetTupleHash(m)

		return nil
	}
}

func testAccCheckRegexMatchSetDisappears(set *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}

			for _, tuple := range set.RegexMatchTuples {
				req.Updates = append(req.Updates, &waf.RegexMatchSetUpdate{
					Action:          aws.String("DELETE"),
					RegexMatchTuple: tuple,
				})
			}

			return conn.UpdateRegexMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regex Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}
			return conn.DeleteRegexMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regex Match Set: %s", err)
		}

		return nil
	}
}

func testAccCheckRegexMatchSetExists(n string, v *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regex Match Set ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.RegexMatchSet.RegexMatchSetId == rs.Primary.ID {
			*v = *resp.RegexMatchSet
			return nil
		}

		return fmt.Errorf("WAF Regex Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckRegexMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_regex_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RegexMatchSet.RegexMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regex Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Regex Pattern Set is already destroyed
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccRegexMatchSetConfig(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "User-Agent"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_waf_regex_pattern_set.test.id
    text_transformation  = "NONE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccRegexMatchSetConfig_changePatterns(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "Referer"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_waf_regex_pattern_set.test.id
    text_transformation  = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"
}
`, matchSetName)
}
