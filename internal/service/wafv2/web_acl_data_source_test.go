package wafv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/wafv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWAFV2WebACLDataSource_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	datasourceName := "data.aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebACLDataSourceConfig_nonExistent(name),
				ExpectError: regexp.MustCompile(`WAFv2 WebACL not found`),
			},
			{
				Config: testAccWebACLDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "wafv2", regexp.MustCompile(fmt.Sprintf("regional/webacl/%v/.+$", name))),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "scope", resourceName, "scope"),
				),
			},
		},
	})
}

func testAccWebACLDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = "%s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = aws_wafv2_web_acl.test.name
  scope = "REGIONAL"
}
`, name)
}

func testAccWebACLDataSourceConfig_nonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = "%s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-rule-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_wafv2_web_acl" "test" {
  name  = "tf-acc-test-does-not-exist"
  scope = "REGIONAL"
}
`, name)
}
