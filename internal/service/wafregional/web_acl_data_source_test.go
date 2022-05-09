package wafregional_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/wafregional"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccWAFRegionalWebACLDataSource_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_web_acl.web_acl"
	datasourceName := "data.aws_wafregional_web_acl.web_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccWebACLDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`web ACLs not found`),
			},
			{
				Config: testAccWebACLDataSourceConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccWebACLDataSourceConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "web_acl" {
  name        = %[1]q
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }
}

data "aws_wafregional_web_acl" "web_acl" {
  name = aws_wafregional_web_acl.web_acl.name
}
`, name)
}

const testAccWebACLDataSourceConfig_NonExistent = `
data "aws_wafregional_web_acl" "web_acl" {
  name = "tf-acc-test-does-not-exist"
}
`
