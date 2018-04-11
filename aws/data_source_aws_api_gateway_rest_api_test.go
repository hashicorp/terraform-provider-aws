package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsApiGatewayRestApi(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsApiGatewayRestApiConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsApiGatewayRestApiCheck("data.aws_api_gateway_rest_api.by_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayRestApiCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		apiGatewayRestApiResources, ok := s.RootModule().Resources["aws_api_gateway_rest_api.tf_test"]
		if !ok {
			return fmt.Errorf("can't find aws_api_gateway_rest_api.tf_test in state")
		}

		attr := resources.Primary.Attributes

		if attr["name"] != apiGatewayRestApiResources.Primary.Attributes["name"] {
			return fmt.Errorf(
				"name is %s; want %s",
				attr["name"],
				apiGatewayRestApiResources.Primary.Attributes["name"],
			)
		}

		if attr["root_resource_id"] != apiGatewayRestApiResources.Primary.Attributes["root_resource_id"] {
			return fmt.Errorf(
				"root_resource_id is %s; want %s",
				attr["root_resource_id"],
				apiGatewayRestApiResources.Primary.Attributes["root_resource_id"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsApiGatewayRestApiConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_api_gateway_rest_api" "tf_wrong1" {
  name        = "wrong1"
}

resource "aws_api_gateway_rest_api" "tf_test" {
  name        = "tf_test"
}

resource "aws_api_gateway_rest_api" "tf_wrong2" {
  name        = "wrong2"
}

data "aws_api_gateway_rest_api" "by_name" {
  name = "${aws_api_gateway_rest_api.tf_test.name}"
}
`
