package waf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWAFRuleDataSource_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_rule.wafrule"
	datasourceName := "data.aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(waf.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRuleDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`WAF Rules not found`),
			},
			{
				Config: testAccRuleDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccRuleDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "wafrule" {
  name        = %[1]q
  metric_name = "WafruleTest"
}

data "aws_waf_rule" "wafrule" {
  name = aws_waf_rule.wafrule.name
}
`, name)
}

const testAccRuleDataSourceConfig_nonExistent = `
data "aws_waf_rule" "wafrule" {
  name = "tf-acc-test-does-not-exist"
}
`
