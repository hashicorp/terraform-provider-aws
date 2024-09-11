// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2API_basicWebSocket(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeWebsocket)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
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

func TestAccAPIGatewayV2API_basicHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
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

func TestAccAPIGatewayV2API_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2API_allAttributesWebSocket(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeWebsocket)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
				),
			},
			{
				Config: testAccAPIConfig_basicWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeWebsocket)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
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

func TestAccAPIGatewayV2API_allAttributesHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
				),
			},
			{
				Config: testAccAPIConfig_basicHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "v1"),
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

func TestAccAPIGatewayV2API_openAPI(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_open(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_DIFFERENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /test"}),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_updatedOpenYAML(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_DIFFERENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2.0"),
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_openYAMLTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value3"),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_openYAMLTagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "Value4"),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withCors(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_openYAMLCorsConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "delete"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "cors_configuration.0.allow_methods"},
			},
			{
				Config: testAccAPIConfig_openYAMLCorsConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				Config: testAccAPIConfig_openYAMLCorsConfigurationUpdated2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "get"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withMoreFields(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_openYAML(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_DIFFERENT"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /test"}),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_updatedOpen2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description different"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName+"_DIFFERENT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2.0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_failOnWarnings(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			// Invalid body should not be accepted when fail_on_warnings is enabled
			{
				Config:      testAccAPIConfig_failOnWarnings(rName, "fail_on_warnings = true"),
				ExpectError: regexache.MustCompile(`BadRequestException: Warnings found during import`),
			},
			// Warnings do not break the deployment when fail_on_warnings is disabled
			{
				Config: testAccAPIConfig_failOnWarnings(rName, "fail_on_warnings = false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "Title test"),
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "fail_on_warnings", acctest.CtFalse),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
				),
				ExpectNonEmptyPlan: true, // OpenAPI definition overrides HCL configuration.
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "fail_on_warnings"},
			},
			// fail_on_warnings should be optional and false by default
			{
				Config: testAccAPIConfig_failOnWarnings(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					testAccCheckAPIRoutes(ctx, &v, []string{"GET /update"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "fail_on_warnings"},
			},
		},
	})
}

func testAccCheckAPIRoutes(ctx context.Context, v *apigatewayv2.GetApiOutput, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		resp, err := conn.GetRoutes(ctx, &apigatewayv2.GetRoutesInput{
			ApiId: v.ApiId,
		})
		if err != nil {
			return err
		}

		actualRoutePaths := map[string]bool{}
		for _, route := range resp.Items {
			actualRoutePaths[*route.RouteKey] = true
		}

		for _, route := range routes {
			if _, ok := actualRoutePaths[route]; !ok {
				return fmt.Errorf("Expected path %v but did not find it in %v", route, actualRoutePaths)
			}
			delete(actualRoutePaths, route)
		}

		if len(actualRoutePaths) > 0 {
			return fmt.Errorf("Found unexpected paths %v", actualRoutePaths)
		}

		return nil
	}
}

func TestAccAPIGatewayV2API_cors(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_corsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_headers.*", "Authorization"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cors_configuration.0.allow_headers", "cors_configuration.0.allow_methods"},
			},
			{
				Config: testAccAPIConfig_corsConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "HTTP://WWW.EXAMPLE.ORG"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://example.io"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.expose_headers.*", "X-Api-Id"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", "500"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
			{
				Config: testAccAPIConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_quickCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_quickCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, resourceName, &v),
					testAccCheckAPIQuickCreateIntegration(ctx, resourceName, "HTTP_PROXY", "http://www.example.com/"),
					testAccCheckAPIQuickCreateRoute(ctx, resourceName, "GET /pets"),
					testAccCheckAPIQuickCreateStage(ctx, resourceName, "$default"),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", string(awstypes.ProtocolTypeHttp)),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /pets"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrTarget, "http://www.example.com/"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"route_key",
					names.AttrTarget,
				},
			},
		},
	})
}

func testAccCheckAPIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_api" {
				continue
			}

			_, err := tfapigatewayv2.FindAPIByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 API %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIExists(ctx context.Context, n string, v *apigatewayv2.GetApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindAPIByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAPIQuickCreateIntegration(ctx context.Context, n, expectedType, expectedUri string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		resp, err := conn.GetIntegrations(ctx, &apigatewayv2.GetIntegrationsInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of integrations: %d", got)
		}

		if got := string(resp.Items[0].IntegrationType); got != expectedType {
			return fmt.Errorf("Incorrect integration type. Expected: %s, got: %s", expectedType, got)
		}
		if got := aws.ToString(resp.Items[0].IntegrationUri); got != expectedUri {
			return fmt.Errorf("Incorrect integration URI. Expected: %s, got: %s", expectedUri, got)
		}

		return nil
	}
}

func testAccCheckAPIQuickCreateRoute(ctx context.Context, n, expectedRouteKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		resp, err := conn.GetRoutes(ctx, &apigatewayv2.GetRoutesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of routes: %d", got)
		}

		if got := aws.ToString(resp.Items[0].RouteKey); got != expectedRouteKey {
			return fmt.Errorf("Incorrect route key. Expected: %s, got: %s", expectedRouteKey, got)
		}

		return nil
	}
}

func testAccCheckAPIQuickCreateStage(ctx context.Context, n, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		resp, err := conn.GetStages(ctx, &apigatewayv2.GetStagesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of stages: %d", got)
		}

		if got := aws.ToString(resp.Items[0].StageName); got != expectedName {
			return fmt.Errorf("Incorrect stage name. Expected: %s, got: %s", expectedName, got)
		}

		return nil
	}
}

func testAccAPIConfig_basicWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccAPIConfig_basicHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccAPIConfig_allAttributesWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  api_key_selection_expression = "$context.authorizer.usageIdentifierKey"
  description                  = "test description"
  disable_execute_api_endpoint = true
  name                         = %[1]q
  protocol_type                = "WEBSOCKET"
  route_selection_expression   = "$request.body.service"
  version                      = "v1"
}
`, rName)
}

func testAccAPIConfig_allAttributesHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  description                  = "test description"
  disable_execute_api_endpoint = true
  name                         = %[1]q
  protocol_type                = "HTTP"
  version                      = "v1"
}
`, rName)
}

func testAccAPIConfig_corsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"

  cors_configuration {
    allow_headers = ["Authorization"]
    allow_methods = ["GET", "put"]
    allow_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccAPIConfig_corsConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"

  cors_configuration {
    allow_credentials = true
    allow_methods     = ["*"]
    allow_origins     = ["HTTP://WWW.EXAMPLE.ORG", "https://example.io"]
    expose_headers    = ["X-Api-Id"]
    max_age           = 500
  }
}
`, rName)
}

func testAccAPIConfig_quickCreate(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  target        = "http://www.example.com/"
  route_key     = "GET /pets"
}
`, rName)
}

func testAccAPIConfig_open(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%[1]s_DIFFERENT",
    "version": "1.0"
  },
  "paths": {
    "/test": {
      "get": {
        "x-amazon-apigateway-integration": {
          "type": "HTTP_PROXY",
          "httpMethod": "GET",
          "payloadFormatVersion": "1.0",
          "uri": "https://www.google.de"
        }
      }
    }
  }
}
EOF
}
`, rName)
}

func testAccAPIConfig_openYAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 1.0
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_openYAMLCorsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["delete"]
    allow_origins = ["https://www.google.de"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_openYAMLCorsConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_openYAMLCorsConfigurationUpdated2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["put", "get"]
    allow_origins = ["https://www.google.de", "https://www.example.com"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
x-amazon-apigateway-cors:
  allow_methods:
    - delete
  allow_origins:
    - https://www.google.de
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_openYAMLTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
tags:
  - name: Key1
    x-amazon-apigateway-tag-value: Value3
paths:
  "/test":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: '1.0'
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_openYAMLTagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1U"
    Key2 = "Value2U"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
tags:
  - name: Key3
    x-amazon-apigateway-tag-value: Value3
  - name: Key4
    x-amazon-apigateway-tag-value: Value4
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_updatedOpenYAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %[1]s_DIFFERENT
  version: 2.0
paths:
  "/update":
    get:
      x-amazon-apigateway-integration:
        type: HTTP_PROXY
        httpMethod: GET
        payloadFormatVersion: 1.0
        uri: https://www.google.de
EOF
}
`, rName)
}

func testAccAPIConfig_updatedOpen2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  version       = "2017-04-21T04:08:08Z"
  description   = "description test"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%[1]s_DIFFERENT",
    "version": "2.0",
    "description": "description different"
  },
  "paths": {
    "/update": {
      "get": {
        "x-amazon-apigateway-integration": {
          "type": "HTTP_PROXY",
          "httpMethod": "GET",
          "payloadFormatVersion": "1.0",
          "uri": "https://www.google.de"
        }
      }
    }
  }
}
EOF
}
`, rName)
}

func testAccAPIConfig_failOnWarnings(rName string, failOnWarnings string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "Title test",
    "version": "2.0",
    "description": "Description test"
  },
  "paths": {
    "/update": {
      "get": {
        "x-amazon-apigateway-integration": {
          "type": "HTTP_PROXY",
          "httpMethod": "GET",
          "payloadFormatVersion": "1.0",
          "uri": "https://www.google.de"
        },
        "responses": {
          "200": {
            "description": "Response description",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ModelThatDoesNotExist"
                }
              }
            }
          }
        }
      }
    }
  }
}
EOF
  %[2]s
}
`, rName, failOnWarnings)
}
