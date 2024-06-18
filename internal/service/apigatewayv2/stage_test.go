// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2Stage_basicWebSocket(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_basicHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_defaultHTTPStage(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_autoDeployHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_autoDeployHTTP(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_autoDeployHTTP(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					// The stage's DeploymentId attribute will be set asynchronously as deployments are done automatically.
					// resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       testAccStageImportStateIdFunc(resourceName),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deployment_id"},
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigatewayv2.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_accessLogSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	cloudWatchResourceName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettings(rName, "$context.identity.sourceIp $context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.identity.sourceIp $context.requestId"),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_accessLogSettings(rName, "$context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.requestId"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_clientCertificateIdAndDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	certificateResourceName := "aws_api_gateway_client_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_clientCertificateIdAndDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "client_certificate_id", certificateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test stage"),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_clientCertificateIdAndDescriptionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test stage updated"),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_defaultRouteSettingsWebSocket(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultRouteSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_defaultRouteSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"), // No drift detection if not configured
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_defaultRouteSettingsHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultRouteSettingsHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_defaultRouteSettingsHTTPUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_deployment(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	deploymentResourceName := "aws_apigatewayv2_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_deployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_id", deploymentResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_routeSettingsWebSocket(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "$default",
						"throttling_burst_limit":   acctest.Ct0,
						"throttling_rate_limit":    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtTrue,
						"detailed_metrics_enabled": acctest.CtTrue,
						"logging_level":            "ERROR",
						"route_key":                "$connect",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"logging_level":            "INFO",
						"route_key":                "$connect",
						"throttling_burst_limit":   acctest.Ct0,
						"throttling_rate_limit":    acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "$disconnect",
						"throttling_burst_limit":   acctest.Ct0,
						"throttling_rate_limit":    acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_routeSettingsHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtTrue,
						"route_key":                "$default",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsHTTPUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_RouteSettingsHTTP_withRoute(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsHTTPRoute(rName, "GET /first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "GET /first",
						"throttling_burst_limit":   acctest.Ct1,
						"throttling_rate_limit":    acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsHTTPRoute(rName, "POST /second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       acctest.CtFalse,
						"detailed_metrics_enabled": acctest.CtFalse,
						"route_key":                "POST /second",
						"throttling_burst_limit":   acctest.Ct1,
						"throttling_rate_limit":    acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_stageVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var api apigatewayv2.GetApiOutput
	var v apigatewayv2.GetStageOutput
	apiResourceName := "aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_variables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(ctx, apiResourceName, &api),
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					testAccCheckStageARN(resourceName, names.AttrARN, &api, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &api, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexache.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckStageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_stage" {
				continue
			}

			_, err := tfapigatewayv2.FindStageByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Stage %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStageExists(ctx context.Context, n string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindStageByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStageARN(resourceName, attributeName string, api *apigatewayv2.GetApiOutput, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, "apigateway", fmt.Sprintf("/apis/%s/stages/%s", aws.ToString(api.ApiId), aws.ToString(v.StageName)))(s)
	}
}

func testAccCheckStageExecutionARN(resourceName, attributeName string, api *apigatewayv2.GetApiOutput, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARN(resourceName, attributeName, "execute-api", fmt.Sprintf("%s/%s", aws.ToString(api.ApiId), aws.ToString(v.StageName)))(s)
	}
}

func testAccStageImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["api_id"], rs.Primary.ID), nil
	}
}

func testAccStageConfig_apiWebSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"
}
`, rName)
}

func testAccStageConfig_apiHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName)
}

func testAccStageConfig_basicWebSocket(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q
}
`, rName))
}

func testAccStageConfig_basicHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q
}
`, rName))
}

func testAccStageConfig_defaultHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = "$default"
}
`)
}

func testAccStageConfig_autoDeployHTTP(rName string, autoDeploy bool) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_httpProxy(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"
  target    = "integrations/${aws_apigatewayv2_integration.test.id}"
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  auto_deploy = %[2]t
}
`, rName, autoDeploy))
}

func testAccStageConfig_accessLogSettings(rName, format string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.test.arn
    format          = %[2]q
  }
}
`, rName, format))
}

func testAccStageConfig_clientCertificateIdAndDescription(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  client_certificate_id = aws_api_gateway_client_certificate.test.id
  description           = "Test stage"
}
`, rName))
}

func testAccStageConfig_clientCertificateIdAndDescriptionUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_api_gateway_client_certificate" "test" {
  description = %[1]q
}

resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  description = "Test stage updated"
}
`, rName))
}

func testAccStageConfig_defaultRouteSettingsWebSocket(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "ERROR"
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccStageConfig_defaultRouteSettingsWebSocketUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    data_trace_enabled       = false
    detailed_metrics_enabled = true
    logging_level            = "INFO"
    throttling_burst_limit   = 1111
    throttling_rate_limit    = 9999
  }
}
`, rName))
}

func testAccStageConfig_defaultRouteSettingsHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    detailed_metrics_enabled = true
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccStageConfig_defaultRouteSettingsHTTPUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  default_route_settings {
    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }
}
`, rName))
}

func testAccStageConfig_deployment(rName string) string {
	return acctest.ConfigCompose(
		testAccDeploymentConfig_basic(rName, rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  deployment_id = aws_apigatewayv2_deployment.test.id
}
`, rName))
}

func testAccStageConfig_routeSettingsWebSocket(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"
  }

  route_settings {
    route_key = "$connect"

    data_trace_enabled       = true
    detailed_metrics_enabled = true
    logging_level            = "ERROR"
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccStageConfig_routeSettingsWebSocketUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }

  route_settings {
    route_key = "$connect"

    logging_level = "INFO"
  }

  route_settings {
    route_key = "$disconnect"
  }
}
`, rName))
}

func testAccStageConfig_routeSettingsHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    detailed_metrics_enabled = true
    throttling_burst_limit   = 2222
    throttling_rate_limit    = 8888
  }
}
`, rName))
}

func testAccStageConfig_routeSettingsHTTPUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = "$default"

    throttling_burst_limit = 1111
    throttling_rate_limit  = 9999
  }
}
`, rName))
}

func testAccStageConfig_routeSettingsHTTPRoute(rName, routeKey string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiHTTP(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  route_settings {
    route_key = aws_apigatewayv2_route.test.route_key

    throttling_burst_limit = 1
  }
}

resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = %[2]q
  target    = "integrations/${aws_apigatewayv2_integration.test.id}"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id             = aws_apigatewayv2_api.test.id
  integration_type   = "HTTP_PROXY"
  integration_method = "GET"
  integration_uri    = "https://example.com/"
}
`, rName, routeKey))
}

func testAccStageConfig_variables(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  stage_variables = {
    Var1 = "Value1"
    Var2 = "Value2"
  }
}
`, rName))
}

// testAccPreCheckAPIGatewayAccountCloudWatchRoleARN checks whether a CloudWatch role ARN has been configured in the current AWS region.
func testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

	output, err := conn.GetAccount(ctx, &apigateway.GetAccountInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error reading API Gateway Account: %s", err)
	}

	if output == nil || aws.ToString(output.CloudwatchRoleArn) == "" {
		t.Skip("skipping tests; no API Gateway CloudWatch role ARN has been configured in this region")
	}
}
