package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestDataSourceAwsApiGatewayResource(t *testing.T) {
	rName := acctest.RandString(8)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsApiGatewayResourceCheck("aws_api_gateway_resource.example_v1"),
					testAccDataSourceAwsApiGatewayResourceCheck("aws_api_gateway_resource.example_v1_endpoint"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayResourceCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		created, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		datasource, ok := s.RootModule().Resources[fmt.Sprintf("data.%s", name)]
		if !ok {
			return fmt.Errorf("root modules has no datasource called data.%s", name)
		}

		rattr := created.Primary.Attributes
		dattr := datasource.Primary.Attributes

		if got, want := rattr["id"], dattr["id"]; got != want {
			return fmt.Errorf(
				"id is %s; want %s",
				got,
				want,
			)
		}

		if got, want := rattr["path_part"], dattr["path_part"]; got != want {
			return fmt.Errorf(
				"path_part is %s; want %s",
				got,
				want,
			)
		}

		if got, want := rattr["parent_id"], dattr["parent_id"]; got != want {
			return fmt.Errorf(
				"parent_id is %s; want %s",
				got,
				want,
			)
		}

		return nil
	}
}

func testAccDataSourceAwsApiGatewayResourceConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "example" {
	name = "%s_example"
}

resource "aws_api_gateway_resource" "example_v1" {
	rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    parent_id   = "${aws_api_gateway_rest_api.example.root_resource_id}"
	path_part   = "v1"
}

resource "aws_api_gateway_resource" "example_v1_endpoint" {
	rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    parent_id   = "${aws_api_gateway_resource.example_v1.id}"
	path_part   = "endpoint"
}

data "aws_api_gateway_resource" "example_v1" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  path        = "/v1"
}

data "aws_api_gateway_resource" "example_v1_endpoint" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  path        = "/v1/endpoint"
}
`, r)
}
