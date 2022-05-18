package apigateway_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayExportDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_export.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExportDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aws_api_gateway_stage.test", "rest_api_id", dataSourceName, "rest_api_id"),
					resource.TestCheckResourceAttrPair("aws_api_gateway_stage.test", "stage_name", dataSourceName, "stage_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "body"),
					resource.TestCheckResourceAttrSet(dataSourceName, "content_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, "content_disposition"),
				),
			},
		},
	})
}

func testAccExportDataSourceConfig(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id
}

data "aws_api_gateway_export" "test" {
  rest_api_id = aws_api_gateway_stage.test.rest_api_id
  stage_name  = aws_api_gateway_stage.test.stage_name
  export_type = "oas30"
}
`
}
