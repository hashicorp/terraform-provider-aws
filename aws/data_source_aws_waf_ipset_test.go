package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsWafIPSet_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_ipset.ipset"
	datasourceName := "data.aws_waf_ipset.ipset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(waf.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafIPSet_NonExistent,
				ExpectError: regexp.MustCompile(`WAF IP Set not found`),
			},
			{
				Config: testAccDataSourceAwsWafIPSet_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafIPSet_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q
}

data "aws_waf_ipset" "ipset" {
  name = aws_waf_ipset.ipset.name
}
`, name)
}

const testAccDataSourceAwsWafIPSet_NonExistent = `
data "aws_waf_ipset" "ipset" {
  name = "tf-acc-test-does-not-exist"
}
`
