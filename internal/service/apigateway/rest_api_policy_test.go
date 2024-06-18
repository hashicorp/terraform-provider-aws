// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayRestAPIPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigateway.GetRestApiOutput
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestAPIPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Action":"execute-api:Invoke".+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
			{
				Config: testAccRestAPIPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"aws:SourceIp":"123.123.123.123/32".+`))),
			},
		},
	})
}

func TestAccAPIGatewayRestAPIPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigateway.GetRestApiOutput
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestAPIPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPIPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPIPolicy_Disappears_restAPI(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigateway.GetRestApiOutput
	resourceName := "aws_api_gateway_rest_api_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRestAPIPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPI(), "aws_api_gateway_rest_api.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRestAPIPolicyExists(ctx context.Context, n string, res *apigateway.GetRestApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		req := &apigateway.GetRestApiInput{
			RestApiId: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetRestApi(ctx, req)
		if err != nil {
			return err
		}

		normalizedPolicy, err := structure.NormalizeJsonString(`"` + aws.ToString(describe.Policy) + `"`)
		if err != nil {
			return fmt.Errorf("error normalizing API Gateway REST API policy JSON: %w", err)
		}
		policy, err := strconv.Unquote(normalizedPolicy)
		if err != nil {
			return fmt.Errorf("error unescaping API Gateway REST API policy: %w", err)
		}

		if aws.ToString(describe.Id) != rs.Primary.ID &&
			policy != rs.Primary.Attributes[names.AttrPolicy] {
			return fmt.Errorf("API Gateway REST API Policy not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckRestAPIPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_rest_api_policy" {
				continue
			}

			req := &apigateway.GetRestApisInput{}
			describe, err := conn.GetRestApis(ctx, req)
			if err == nil {
				if len(describe.Items) != 0 &&
					aws.ToString(describe.Items[0].Id) == rs.Primary.ID &&
					aws.ToString(describe.Items[0].Policy) == "" {
					return fmt.Errorf("API Gateway REST API Policy still exists")
				}
			}

			return err
		}

		return nil
	}
}

func testAccRestAPIPolicyConfig_basic(rName string) string {
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

func testAccRestAPIPolicyConfig_updated(rName string) string {
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
