package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsApiGatewayRestApi(t *testing.T) {
	rName := acctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_rest_api.test"
	resourceName := "aws_api_gateway_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayRestApiConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_resource_id", resourceName, "root_resource_id"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

//func testAccDataSourceAwsApiGatewayRestApiCheck(name string) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		resources, ok := s.RootModule().Resources[name]
//		if !ok {
//			return fmt.Errorf("root module has no resource called %s", name)
//		}
//
//		apiGatewayRestApiResources, ok := s.RootModule().Resources["aws_api_gateway_rest_api.tf_test"]
//		if !ok {
//			return fmt.Errorf("can't find aws_api_gateway_rest_api.tf_test in state")
//		}
//
//		attr := resources.Primary.Attributes
//
//		if attr["name"] != apiGatewayRestApiResources.Primary.Attributes["name"] {
//			return fmt.Errorf(
//				"name is %s; want %s",
//				attr["name"],
//				apiGatewayRestApiResources.Primary.Attributes["name"],
//			)
//		}
//
//		if attr["root_resource_id"] != apiGatewayRestApiResources.Primary.Attributes["root_resource_id"] {
//			return fmt.Errorf(
//				"root_resource_id is %s; want %s",
//				attr["root_resource_id"],
//				apiGatewayRestApiResources.Primary.Attributes["root_resource_id"],
//			)
//		}
//
//		return nil
//	}
//}

func testAccDataSourceAwsApiGatewayRestApiConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%[1]s"
}

data "aws_api_gateway_rest_api" "test" {
  name = "${aws_api_gateway_rest_api.test.name}"
}
`, r)
}
