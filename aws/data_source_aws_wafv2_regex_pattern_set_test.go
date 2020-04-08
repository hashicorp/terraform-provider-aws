package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsWafv2RegexPatternSet_Basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_regex_pattern_set.test"
	datasourceName := "data.aws_wafv2_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafv2RegexPatternSet_NonExistent(name),
				ExpectError: regexp.MustCompile(`WAFV2 RegexPatternSet not found`),
			},
			{
				Config: testAccDataSourceAwsWafv2RegexPatternSet_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "scope", resourceName, "scope"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafv2RegexPatternSet_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%s"
  scope = "REGIONAL"

  regular_expression_list {
    regex_string = "one"
  }
}

data "aws_wafv2_regex_pattern_set" "test" {
  name  = aws_wafv2_regex_pattern_set.test.name
  scope = "REGIONAL"
}
`, name)
}

func testAccDataSourceAwsWafv2RegexPatternSet_NonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%s"
  scope = "REGIONAL"

  regular_expression_list {
    regex_string = "one"
  }
}

data "aws_wafv2_regex_pattern_set" "test" {
  name  = "tf-acc-test-does-not-exist"
  scope = "REGIONAL"
}
`, name)
}
