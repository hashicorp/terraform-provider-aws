// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegexPatternSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_regex_pattern_set.test"
	datasourceName := "data.aws_waf_regex_pattern_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRegexPatternSetDataSourceConfig_nonExistent(name),
				ExpectError: regexache.MustCompile(`WAF RegexPatternSet not found`),
			},
			{
				Config: testAccRegexPatternSetDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "regex_pattern_strings.#", resourceName, "regex_pattern_strings.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "regex_pattern_strings.0", resourceName, "regex_pattern_strings.0"),
				),
			},
		},
	})
}

func testAccRegexPatternSetDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"

  regex_pattern_strings = ["one"]
}

data "aws_waf_regex_pattern_set" "test" {
  name = aws_waf_regex_pattern_set.test.name
}
`, name)
}

func testAccRegexPatternSetDataSourceConfig_nonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"

  regex_pattern_strings = ["one"]
}

data "aws_waf_regex_pattern_set" "test" {
  name = "tf-acc-test-does-not-exist"
}
`, name)
}
