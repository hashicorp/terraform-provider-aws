package apigateway_test

import (
	"strconv"
	"testing"

	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayAuthorizersDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "data.aws_api_gateway_authorizers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizerDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "items.0.identity_source", "method.request.header.Authorization"),
					resource.TestCheckResourceAttr(resourceName, "items.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "items.0.type", "TOKEN"),
					resource.TestCheckResourceAttr(resourceName, "items.0.authorizer_result_ttl_in_seconds", strconv.Itoa(tfapigateway.DefaultAuthorizerTTL)),
				),
			},
		},
	})
}

func testAccAuthorizerDataSource(rName string) string {
	return testAccAuthorizerConfig_lambda(rName) + `
data "aws_api_gateway_authorizers" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  depends_on  = [aws_api_gateway_authorizer.test]
}
`
}
