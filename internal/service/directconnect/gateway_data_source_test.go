package directconnect_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDirectConnectGatewayDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dx_gateway.test"
	datasourceName := "data.aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccGatewayDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Direct Connect Gateway not found`),
			},
			{
				Config: testAccGatewayDataSourceConfig_name(rName, sdkacctest.RandIntRange(64512, 65534)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "amazon_side_asn", resourceName, "amazon_side_asn"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_account_id", resourceName, "owner_account_id"),
				),
			},
		},
	})
}

func testAccGatewayDataSourceConfig_name(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "wrong" {
  amazon_side_asn = "%d"
  name            = "%s-wrong"
}

resource "aws_dx_gateway" "test" {
  amazon_side_asn = "%d"
  name            = "%s"
}

data "aws_dx_gateway" "test" {
  name = aws_dx_gateway.test.name
}
`, rBgpAsn+1, rName, rBgpAsn, rName)
}

const testAccGatewayDataSourceConfig_nonExistent = `
data "aws_dx_gateway" "test" {
  name = "tf-acc-test-does-not-exist"
}
`
