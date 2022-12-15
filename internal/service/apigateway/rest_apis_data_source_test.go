package apigateway_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayRestAPIsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_api_gateway_rest_apis.test"
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testRestAPIsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testCheckItemInList(dataSourceName, "names", rName),
				),
			},
		},
	})
}

func testCheckItemInList(resource, attr, value string) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		attrValue, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("Attribute %s not found", attr)
		}

		if !strings.Contains(attrValue, value) {
			return fmt.Errorf("Attribute %s does not contain %s", attr, value)
		}

		return nil
	}
}

func testRestAPIsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

data "aws_api_gateway_rest_apis" "test" {}
`, rName)
}
