package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/wafv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccDataSourceAwsWafv2IPSet_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_ip_set.test"
	datasourceName := "data.aws_wafv2_ip_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		ErrorCheck: acctest.ErrorCheck(t, wafv2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafv2IPSet_NonExistent(name),
				ExpectError: regexp.MustCompile(`WAFv2 IPSet not found`),
			},
			{
				Config: testAccDataSourceAwsWafv2IPSet_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "addresses", resourceName, "addresses"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(datasourceName, "arn", "wafv2", regexp.MustCompile(fmt.Sprintf("regional/ipset/%v/.+$", name))),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_address_version", resourceName, "ip_address_version"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "scope", resourceName, "scope"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafv2IPSet_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
}

data "aws_wafv2_ip_set" "test" {
  name  = aws_wafv2_ip_set.test.name
  scope = "REGIONAL"
}
`, name)
}

func testAccDataSourceAwsWafv2IPSet_NonExistent(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
}

data "aws_wafv2_ip_set" "test" {
  name  = "tf-acc-test-does-not-exist"
  scope = "REGIONAL"
}
`, name)
}
