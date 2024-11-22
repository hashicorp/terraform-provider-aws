// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayBasePathMapping_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetBasePathMappingOutput
	resourceName := "aws_api_gateway_base_path_mapping.test"
	name := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBasePathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBasePathMappingConfig_basic(name, key, certificate, acctest.ResourcePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/9212
func TestAccAPIGatewayBasePathMapping_BasePath_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetBasePathMappingOutput
	resourceName := "aws_api_gateway_base_path_mapping.test"
	name := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBasePathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBasePathMappingConfig_basic(name, key, certificate, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayBasePathMapping_updates(t *testing.T) {
	ctx := acctest.Context(t)
	var confFirst, conf apigateway.GetBasePathMappingOutput
	resourceName := "aws_api_gateway_base_path_mapping.test"
	name := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBasePathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBasePathMappingConfig_basic(name, key, certificate, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &confFirst),
					testAccCheckBasePathStageAttribute(&confFirst, "test"),
				),
			},
			{
				Config: testAccBasePathMappingConfig_altStageAndAPI(name, key, certificate, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &conf),
					testAccCheckBasePathBasePathAttribute(&conf, "(none)"),
					testAccCheckBasePathStageAttribute(&conf, "test2"),
					testAccCheckRestAPIIDAttributeHasChanged(&conf, &confFirst),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "test2"),
				),
			},
			{
				Config: testAccBasePathMappingConfig_altStageAndAPI(name, key, certificate, "thing"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &conf),
					testAccCheckBasePathBasePathAttribute(&conf, "thing"),
					testAccCheckBasePathStageAttribute(&conf, "test2"),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "base_path", "thing"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayBasePathMapping_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetBasePathMappingOutput
	name := acctest.RandomSubdomain()
	resourceName := "aws_api_gateway_base_path_mapping.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBasePathDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBasePathMappingConfig_basic(name, key, certificate, acctest.ResourcePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBasePathExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceBasePathMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBasePathExists(ctx context.Context, n string, v *apigateway.GetBasePathMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		basePath := rs.Primary.Attributes["base_path"]
		if basePath == "" {
			basePath = "(none)"
		}
		output, err := tfapigateway.FindBasePathMappingByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], basePath)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBasePathDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_base_path_mapping" {
				continue
			}

			basePath := rs.Primary.Attributes["base_path"]
			if basePath == "" {
				basePath = "(none)"
			}
			_, err := tfapigateway.FindBasePathMappingByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], basePath)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Base Path Mapping %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBasePathStageAttribute(conf *apigateway.GetBasePathMappingOutput, basePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Stage == nil {
			return fmt.Errorf("attribute Stage should not be nil")
		}
		if *conf.Stage != basePath {
			return fmt.Errorf("unexpected value Stage: %s", *conf.Stage)
		}

		return nil
	}
}

func testAccCheckRestAPIIDAttributeHasChanged(conf *apigateway.GetBasePathMappingOutput, previousConf *apigateway.GetBasePathMappingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.RestApiId == nil {
			return fmt.Errorf("attribute RestApiId should not be nil")
		}
		if *conf.RestApiId == *previousConf.RestApiId {
			return fmt.Errorf("expected RestApiId to have changed")
		}

		return nil
	}
}

func testAccCheckBasePathBasePathAttribute(conf *apigateway.GetBasePathMappingOutput, basePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.Stage == nil {
			return fmt.Errorf("attribute Stage should not be nil")
		}
		if *conf.BasePath != basePath {
			return fmt.Errorf("unexpected value Stage: %s", *conf.BasePath)
		}

		return nil
	}
}

func testAccBasePathConfig_base(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

resource "aws_api_gateway_rest_api" "test" {
  name        = "tf-acc-apigateway-base-path-mapping"
  description = "Terraform Acceptance Tests"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

# API gateway won't let us deploy an empty API
resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "tf-acc"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "test"
  depends_on  = [aws_api_gateway_integration.test]
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccBasePathMappingConfig_basic(domainName, key, certificate, basePath string) string {
	return acctest.ConfigCompose(testAccBasePathConfig_base(domainName, key, certificate), fmt.Sprintf(`
resource "aws_api_gateway_base_path_mapping" "test" {
  api_id      = aws_api_gateway_rest_api.test.id
  base_path   = %[1]q
  stage_name  = aws_api_gateway_deployment.test.stage_name
  domain_name = aws_api_gateway_domain_name.test.domain_name
}
`, basePath))
}

func testAccBasePathMappingConfig_altStageAndAPI(domainName, key, certificate, basePath string) string {
	return acctest.ConfigCompose(testAccBasePathConfig_base(domainName, key, certificate), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test2" {
  name        = "tf-acc-apigateway-base-path-mapping-alt"
  description = "Terraform Acceptance Tests"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}


resource "aws_api_gateway_stage" "test2" {

  depends_on = [
    aws_api_gateway_deployment.test
  ]

  stage_name    = "test2"
  rest_api_id   = aws_api_gateway_rest_api.test2.id
  deployment_id = aws_api_gateway_deployment.test2.id
}

resource "aws_api_gateway_resource" "test2" {
  rest_api_id = aws_api_gateway_rest_api.test2.id
  parent_id   = aws_api_gateway_rest_api.test2.root_resource_id
  path_part   = "tf-acc"
}

resource "aws_api_gateway_method" "test2" {
  rest_api_id   = aws_api_gateway_rest_api.test2.id
  resource_id   = aws_api_gateway_resource.test2.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "test2" {
  rest_api_id = aws_api_gateway_rest_api.test2.id
  resource_id = aws_api_gateway_resource.test2.id
  http_method = aws_api_gateway_method.test2.http_method
  type        = "MOCK"
}


resource "aws_api_gateway_deployment" "test2" {
  rest_api_id = aws_api_gateway_rest_api.test2.id
  stage_name  = "test"
  depends_on  = [aws_api_gateway_integration.test2]
}

resource "aws_api_gateway_base_path_mapping" "test" {
  api_id      = aws_api_gateway_rest_api.test2.id
  base_path   = %[1]q
  stage_name  = aws_api_gateway_stage.test2.stage_name
  domain_name = aws_api_gateway_domain_name.test.domain_name
}
`, basePath))
}
