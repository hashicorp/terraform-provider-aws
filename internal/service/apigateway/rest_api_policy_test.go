package apigateway_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayRestAPIPolicy_basic(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Action":"execute-api:Invoke".+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy"},
			},
			{
				Config: testAccRestAPIPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"aws:SourceIp":"123.123.123.123/32".+`))),
			},
		},
	})
}

func TestAccAPIGatewayRestAPIPolicy_disappears(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRestAPIPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPIPolicy_Disappears_restAPI(t *testing.T) {
	var v apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRestAPI(), "aws_api_gateway_rest_api.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRestAPIPolicyExists(n string, res *apigateway.RestApi) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		req := &apigateway.GetRestApiInput{
			RestApiId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetRestApi(req)
		if err != nil {
			return err
		}

		normalizedPolicy, err := structure.NormalizeJsonString(`"` + aws.StringValue(describe.Policy) + `"`)
		if err != nil {
			return fmt.Errorf("error normalizing API Gateway REST API policy JSON: %w", err)
		}
		policy, err := strconv.Unquote(normalizedPolicy)
		if err != nil {
			return fmt.Errorf("error unescaping API Gateway REST API policy: %w", err)
		}

		if aws.StringValue(describe.Id) != rs.Primary.ID &&
			policy != rs.Primary.Attributes["policy"] {
			return fmt.Errorf("API Gateway REST API Policy not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckRestAPIPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_rest_api_policy" {
			continue
		}

		req := &apigateway.GetRestApisInput{}
		describe, err := conn.GetRestApis(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				aws.StringValue(describe.Items[0].Id) == rs.Primary.ID &&
				aws.StringValue(describe.Items[0].Policy) == "" {
				return fmt.Errorf("API Gateway REST API Policy still exists")
			}
		}

		return err
	}

	return nil
}

func testAccRestAPIPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api_policy" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Deny"
      Principal = {
        AWS = "*"
      }
      Action   = "execute-api:Invoke"
      Resource = aws_api_gateway_rest_api.test.arn
    }]
  })
}
`, rName)
}

func testAccRestAPIPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api_policy" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "execute-api:Invoke"
      Resource = aws_api_gateway_rest_api.test.arn
      Condition = {
        IpAddress = {
          "aws:SourceIp" = "123.123.123.123/32"
        }
      }
    }]
  })
}
`, rName)
}
