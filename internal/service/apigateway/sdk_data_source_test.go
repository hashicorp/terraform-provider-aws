package apigateway_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewaySdkDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_sdk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSdkDataSourceConfig_basic(rName),
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

func testAccSdkDataSourceConfig_basic(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id
}

data "aws_api_gateway_sdk" "test" {
  rest_api_id = aws_api_gateway_stage.test.rest_api_id
  stage_name  = aws_api_gateway_stage.test.stage_name
  sdk_type    = "android"

  parameters = {
    groupId         = "test"
    artifactId      = "test"
    artifactVersion = "test"
    invokerPackage  = "test"
  }
}
`
}
