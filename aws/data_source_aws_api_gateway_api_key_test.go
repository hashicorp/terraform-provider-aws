package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsApiGatewayApiKey(t *testing.T) {
	rName := acctest.RandString(8)
	resourceName1 := "aws_api_gateway_api_key.example_key"
	dataSourceName1 := "data.aws_api_gateway_api_key.test_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayApiKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "id", dataSourceName1, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "name", dataSourceName1, "name"),
					resource.TestCheckResourceAttrPair(resourceName1, "value", dataSourceName1, "value"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayApiKeyConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "example" {
    name = "example"
}

resource "aws_api_gateway_resource" "example_v1" {
    rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    parent_id   = "${aws_api_gateway_rest_api.example.root_resource_id}"
    path_part   = "v1"
}

resource "aws_api_gateway_method" "example_v1_method" {
    rest_api_id   = "${aws_api_gateway_rest_api.example.id}"
    resource_id   = "${aws_api_gateway_resource.example_v1.id}"
    http_method   = "GET"
    authorization = "NONE"
}

resource "aws_api_gateway_integration" "example_v1_integration" {
    rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    resource_id = "${aws_api_gateway_resource.example_v1.id}"
    http_method = "${aws_api_gateway_method.example_v1_method.http_method}"
    type        = "MOCK"
}

resource "aws_api_gateway_deployment" "example_deployment" {
    rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    stage_name  = "example"
    depends_on  = ["aws_api_gateway_resource.example_v1", "aws_api_gateway_method.example_v1_method", "aws_api_gateway_integration.example_v1_integration"]
}

resource "aws_api_gateway_api_key" "example_key" {
    name = "%s"
}

resource "aws_api_gateway_usage_plan" "example_plan" {
    name = "example"

    api_stages {
        api_id = "${aws_api_gateway_rest_api.example.id}"
        stage  = "${aws_api_gateway_deployment.example_deployment.stage_name}"
    }
}

resource "aws_api_gateway_usage_plan_key" "plan_key" {
    key_id = "${aws_api_gateway_api_key.example_key.id}"
    key_type = "API_KEY"
    usage_plan_id = "${aws_api_gateway_usage_plan.example_plan.id}"
}

data "aws_api_gateway_api_key" "test_key" {
    id = "${aws_api_gateway_api_key.example_key.id}"
}
`, r)
}
