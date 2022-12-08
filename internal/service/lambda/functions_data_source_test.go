package lambda_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaFunctionsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_lambda_functions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLambdaFunctionsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "function_names.#", regexp.MustCompile("[^0].*$")),
					resource.TestMatchResourceAttr(dataSourceName, "function_arns.#", regexp.MustCompile("[^0].*$")),
				),
			},
		},
	})
}

func testAccLambdaFunctionsDataSourceConfig_basic() string {
	return `
data "aws_lambda_functions" "test" {}
`
}
