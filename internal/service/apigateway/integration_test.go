// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccIntegrationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'updated'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-FooBar", "'Baz'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", "{'foobar': 'bar}"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.text/html", "<html>Foo</html>"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "2000"),
				),
			},
			{
				Config: testAccIntegrationConfig_updateURI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de/updated"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "2000"),
				),
			},
			{
				Config: testAccIntegrationConfig_updateNoTemplates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "2000"),
				),
			},
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "29000"),
				),
			},
		},
	})
}

func TestAccAPIGatewayIntegration_contentHandling(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},

			{
				Config: testAccIntegrationConfig_updateContentHandling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_BINARY"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				Config: testAccIntegrationConfig_removeContentHandling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", ""),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayIntegration_Parameters_cacheKey(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_Parameters_cacheKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.param", "method.request.path.param"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.param"),
					resource.TestCheckResourceAttr(resourceName, "cache_namespace", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayIntegration_Parameters_cacheKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_Parameters_cacheKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.param", "method.request.path.param"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.param"),
					resource.TestCheckResourceAttr(resourceName, "cache_namespace", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_cacheKeyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "passthrough_behavior", "WHEN_NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "content_handling", "CONVERT_TO_TEXT"),
					resource.TestCheckResourceAttr(resourceName, "credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Authorization", "'static'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.X-Foo", "'Bar'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.param", "method.request.path.param"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.param"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.test1"),
					resource.TestCheckResourceAttr(resourceName, "cache_namespace", "foobar"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/json", ""),
					resource.TestCheckResourceAttr(resourceName, "request_templates.application/xml", "#set($inputRoot = $input.path('$'))\n{ }"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayIntegration_Parameters_requestUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_Parameters_requestUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-2", "X-Some-Header-3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-2", "method.request.querystring.X-Some-Header-2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-3", "method.request.querystring.X-Some-Header-3"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_requestUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-3", "method.request.querystring.X-Some-Header-3"),
				),
			},
		},
	})
}

func TestAccAPIGatewayIntegration_Parameters_requestCacheKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-2", "X-Some-Header-3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.Host", "'method.request.querystring.name'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.proxy", "method.request.path.proxy"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-2", "method.request.querystring.X-Some-Header-2"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-3", "method.request.querystring.X-Some-Header-3"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "9"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.header.Host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-3"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.Host", "'method.request.querystring.name'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.proxy", "method.request.path.proxy"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-3", "method.request.querystring.X-Some-Header-3"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.header.Host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-3"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-4"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.Host", "'method.request.querystring.name'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.proxy", "method.request.path.proxy"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-4", "method.request.querystring.X-Some-Header-4"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.header.Host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-4"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName, []string{"X-Some-Header-1", "X-Some-Header-4", "X-Some-Header-5", "X-Some-Header-6", "X-Some-Header-7"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "7"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.Host", "'method.request.querystring.name'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.proxy", "method.request.path.proxy"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-1", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-4", "method.request.querystring.X-Some-Header-4"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-5", "method.request.querystring.X-Some-Header-5"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-6", "method.request.querystring.X-Some-Header-6"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-7", "method.request.querystring.X-Some-Header-7"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "13"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.header.Host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-7"),
				),
			},
			{
				Config: testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName, []string{"X-Some-Header-2"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP_PROXY"),
					resource.TestCheckResourceAttr(resourceName, "integration_http_method", "ANY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrURI, "https://www.google.de"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.header.Host", "'method.request.querystring.name'"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.path.proxy", "method.request.path.proxy"),
					resource.TestCheckResourceAttr(resourceName, "request_parameters.integration.request.querystring.X-Some-Header-2", "method.request.querystring.X-Some-Header-2"),
					resource.TestCheckResourceAttr(resourceName, "cache_key_parameters.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.path.proxy"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.header.Host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "method.request.querystring.X-Some-Header-2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cache_key_parameters.*", "integration.request.querystring.X-Some-Header-2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayIntegration_integrationType(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_typeInternet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, ""),
				),
			},
			{
				Config: testAccIntegrationConfig_typeVPCLink(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "VPC_LINK"),
					resource.TestMatchResourceAttr(resourceName, names.AttrConnectionID, regexache.MustCompile("^[0-9a-z]+$")),
				),
			},
			{
				Config: testAccIntegrationConfig_typeInternet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "connection_type", "INTERNET"),
					resource.TestCheckResourceAttr(resourceName, names.AttrConnectionID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayIntegration_TLS_insecureSkipVerification(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_tlsInsecureSkipVerification(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.insecure_skip_verification", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccIntegrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccIntegrationConfig_tlsInsecureSkipVerification(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.0.insecure_skip_verification", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccAPIGatewayIntegration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetIntegrationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_integration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntegrationExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceIntegration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIntegrationExists(ctx context.Context, n string, v *apigateway.GetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindIntegrationByThreePartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIntegrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_integration" {
				continue
			}

			_, err := tfapigateway.FindIntegrationByThreePartKey(ctx, conn, rs.Primary.Attributes["http_method"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["rest_api_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Integration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccIntegrationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["http_method"]), nil
	}
}

func testAccIntegrationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
  }

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
}
`, rName)
}

func testAccIntegrationConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = "{'foobar': 'bar}"
    "text/html"        = "<html>Foo</html>"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'updated'"
    "integration.request.header.X-FooBar"        = "'Baz'"
  }

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_updateURI(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
  }

  type                    = "HTTP"
  uri                     = "https://www.google.de/updated"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_updateContentHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
  }

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_BINARY"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_removeContentHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
  }

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_updateNoTemplates(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_Parameters_cacheKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "{param}"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.path.param" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
    "integration.request.path.param"             = "method.request.path.param"
  }

  cache_key_parameters = ["method.request.path.param"]
  cache_namespace      = "foobar"

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_Parameters_cacheKeyUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "{param}"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.path.param"        = false
    "method.request.querystring.test1" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  request_templates = {
    "application/json" = ""
    "application/xml"  = "#set($inputRoot = $input.path('$'))\n{ }"
  }

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
    "integration.request.header.X-Foo"           = "'Bar'"
    "integration.request.path.param"             = "method.request.path.param"
  }

  cache_key_parameters = ["method.request.path.param", "method.request.querystring.test1"]
  cache_namespace      = "foobar"

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
  timeout_milliseconds    = 2000
}
`, rName)
}

func testAccIntegrationConfig_Parameters_requestUpdate(rName string, params []string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "{param}"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = { for param in var.utm_params : "method.request.querystring.${param}" => false }
}

variable "utm_params" {
  type    = list(string)
  default = ["%[2]s"]
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "ANY"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  timeout_milliseconds    = 2000
  type                    = "HTTP_PROXY"
  uri                     = "https://www.google.de"

  request_parameters = { for param in var.utm_params : "integration.request.querystring.${param}" => "method.request.querystring.${param}" }
}
`, rName, strings.Join(params, `", "`))
}

func testAccIntegrationConfig_Parameters_requestCacheKeyUpdate(rName string, headers []string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "{param}"
}

variable "utm_params" {
  type    = list(string)
  default = ["%[2]s"]
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "ANY"
  authorization = "NONE"

  request_parameters = merge(
    {
      "method.request.path.proxy" = true
    },
    { for param in var.utm_params : "method.request.querystring.${param}" => false }
  )
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "ANY"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  timeout_milliseconds    = 2000
  type                    = "HTTP_PROXY"
  uri                     = "https://www.google.de"

  request_parameters = merge({
    "integration.request.path.proxy"  = "method.request.path.proxy"
    "integration.request.header.Host" = "'method.request.querystring.name'"
    },
    { for param in var.utm_params : "integration.request.querystring.${param}" => "method.request.querystring.${param}" }
  )

  cache_key_parameters = concat(
    ["method.request.path.proxy", "integration.request.path.proxy", "integration.request.header.Host"],
    [for param in var.utm_params : "integration.request.querystring.${param}"],
    [for param in var.utm_params : "method.request.querystring.${param}"]
  )
}
`, rName, strings.Join(headers, `", "`))
}

func testAccIntegrationConfig_IntegrationTypeBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }
}

resource "aws_lb" "test" {
  name               = %[1]q
  internal           = true
  load_balancer_type = "network"
  subnets            = aws_subnet.test[*].id
}

resource "aws_api_gateway_vpc_link" "test" {
  name        = %[1]q
  target_arns = [aws_lb.test.arn]
}
`, rName))
}

func testAccIntegrationConfig_typeVPCLink(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_IntegrationTypeBase(rName), `
resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"

  connection_type = "VPC_LINK"
  connection_id   = aws_api_gateway_vpc_link.test.id
}
`)
}

func testAccIntegrationConfig_typeInternet(rName string) string {
	return acctest.ConfigCompose(testAccIntegrationConfig_IntegrationTypeBase(rName), `
resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"
}
`)
}

func testAccIntegrationConfig_tlsInsecureSkipVerification(rName string, insecureSkipVerification bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"

  request_models = {
    "application/json" = "Error"
  }

  request_parameters = {
    "method.request.path.param" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
  passthrough_behavior    = "WHEN_NO_MATCH"
  content_handling        = "CONVERT_TO_TEXT"

  tls_config {
    insecure_skip_verification = %[2]t
  }
}
`, rName, insecureSkipVerification)
}
