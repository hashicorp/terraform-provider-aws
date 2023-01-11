package lambda_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaFunctionsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_functions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "function_arns.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "function_names.#", "0"),
				),
			},
		},
	})
}

func testAccFunctionsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName, rName, rName, rName), `
data "aws_lambda_functions" "test" {
  depends_on = [aws_lambda_function.test]
}
`)
}
