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

func TestAccWAFRegionalIPSetDataSource_basic(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafregional_ipset.ipset"
	datasourceName := "data.aws_wafregional_ipset.ipset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIPSetDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`WAF Regional IP Set not found`),
			},
			{
				Config: testAccIPSetDataSourceConfig_name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccIPSetDataSourceConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q
}

data "aws_wafregional_ipset" "ipset" {
  name = aws_wafregional_ipset.ipset.name
}
`, name)
}

const testAccIPSetDataSourceConfig_nonExistent = `
data "aws_wafregional_ipset" "ipset" {
  name = "tf-acc-test-does-not-exist"
}
`
