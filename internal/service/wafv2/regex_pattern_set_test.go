package wafv2_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_wafv2_regex_pattern_set", &resource.Sweeper{
		Name: "aws_wafv2_regex_pattern_set",
		F:    sweepRegexPatternSets,
		Dependencies: []string{
			"aws_wafv2_rule_group",
			"aws_wafv2_web_acl",
		},
	})
}

func sweepRegexPatternSets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFV2Conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = tfwafv2.ListRegexPatternSetsPages(conn, input, func(page *wafv2.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexPatternSet := range page.RegexPatternSets {
			id := aws.StringValue(regexPatternSet.Id)

			r := tfwafv2.ResourceRegexPatternSet()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", regexPatternSet.LockToken)
			d.Set("name", regexPatternSet.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 Regex Pattern Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 Regex Pattern Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 Regex Pattern Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccWAFV2RegexPatternSet_basic(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "one",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "two",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccRegexPatternSetConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "one",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "two",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regular_expression.*", map[string]string{
						"regex_string": "three",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v wafv2.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_Minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafv2.ResourceRegexPatternSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_minimal(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_Minimal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "0"),
				),
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_changeNameForceNew(t *testing.T) {
	var before, after wafv2.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNewName := fmt.Sprintf("regex-pattern-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
				),
			},
			{
				Config: testAccRegexPatternSetConfig(rNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rNewName),
					resource.TestCheckResourceAttr(resourceName, "description", rNewName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "regular_expression.#", "2"),
				),
			},
		},
	})
}

func TestAccWAFV2RegexPatternSet_tags(t *testing.T) {
	var v wafv2.RegexPatternSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegexPatternSetConfig_OneTag(rName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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
				Config: testAccRegexPatternSetConfig_TwoTags(rName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccRegexPatternSetConfig_OneTag(rName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegexPatternSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/regexpatternset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func testAccCheckRegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_regex_pattern_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetRegexPatternSet(
			&wafv2.GetRegexPatternSetInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp != nil && resp.RegexPatternSet != nil && aws.StringValue(resp.RegexPatternSet.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 RegexPatternSet %s still exists", rs.Primary.ID)
			}
			return nil
		}

		// Return nil if the RegexPatternSet is already destroyed
		if tfawserr.ErrMessageContains(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckRegexPatternSetExists(n string, v *wafv2.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 RegexPatternSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetRegexPatternSet(&wafv2.GetRegexPatternSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.RegexPatternSet == nil {
			return fmt.Errorf("Error getting WAFv2 RegexPatternSet for %s", rs.Primary.ID)
		}

		if aws.StringValue(resp.RegexPatternSet.Id) == rs.Primary.ID {
			*v = *resp.RegexPatternSet
			return nil
		}

		return fmt.Errorf("WAFv2 RegexPatternSet (%s) not found", rs.Primary.ID)
	}
}

func testAccRegexPatternSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }
}
`, name, name)
}

func testAccRegexPatternSetConfig_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
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

func testAccRegexPatternSetConfig_Minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%s"
  scope = "REGIONAL"
}
`, name)
}

func testAccRegexPatternSetConfig_OneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  regular_expression {
    regex_string = "two"
  }

  tags = {
    "%s" = "%s"
  }
}
`, name, name, tagKey, tagValue)
}

func testAccRegexPatternSetConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRegexPatternSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
