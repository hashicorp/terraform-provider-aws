package apigatewayv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
)

func TestAccAPIGatewayV2Stage_basicWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/", acctest.Region()))),
					resource.TestCheckResourceAttr(resourceName, "name", "$default"),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_autoDeployHTTP(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_autoDeployHTTP(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					// The stage's DeploymentId attribute will be set asynchronously as deployments are done automatically.
					// resource.TestCheckResourceAttrSet(resourceName, "deployment_id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigatewayv2.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_accessLogSettings(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	cloudWatchResourceName := "aws_cloudwatch_log_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettings(rName, "$context.identity.sourceIp $context.requestId"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.identity.sourceIp $context.requestId"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudWatchResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", "$context.requestId"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_clientCertificateIdAndDescription(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	certificateResourceName := "aws_api_gateway_client_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_clientCertificateIdAndDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "client_certificate_id", certificateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test stage"),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", "Test stage updated"),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_defaultRouteSettingsWebSocket(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultRouteSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_defaultRouteSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "INFO"), // No drift detection if not configured
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_defaultRouteSettingsHTTP(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_defaultRouteSettingsHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "2222"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "8888"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_defaultRouteSettingsHTTPUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "1111"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "9999"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_deployment(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	deploymentResourceName := "aws_apigatewayv2_deployment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_deployment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_id", deploymentResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "true",
						"detailed_metrics_enabled": "true",
						"logging_level":            "ERROR",
						"route_key":                "$connect",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsWebSocketUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"logging_level":            "INFO",
						"route_key":                "$connect",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$disconnect",
						"throttling_burst_limit":   "0",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_routeSettingsHTTP(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "true",
						"route_key":                "$default",
						"throttling_burst_limit":   "2222",
						"throttling_rate_limit":    "8888",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsHTTPUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "$default",
						"throttling_burst_limit":   "1111",
						"throttling_rate_limit":    "9999",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_RouteSettingsHTTP_withRoute(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_routeSettingsHTTPRoute(rName, "GET /first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "GET /first",
						"throttling_burst_limit":   "1",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccStageConfig_routeSettingsHTTPRoute(rName, "POST /second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("https://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route_settings.*", map[string]string{
						"data_trace_enabled":       "false",
						"detailed_metrics_enabled": "false",
						"route_key":                "POST /second",
						"throttling_burst_limit":   "1",
						"throttling_rate_limit":    "0",
					}),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_variables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.Var2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2Stage_tags(t *testing.T) {
	var apiId string
	var v apigatewayv2.GetStageOutput
	resourceName := "aws_apigatewayv2_stage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					testAccCheckStageARN(resourceName, "arn", &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					testAccCheckStageExecutionARN(resourceName, "execution_arn", &apiId, &v),
					resource.TestMatchResourceAttr(resourceName, "invoke_url", regexp.MustCompile(fmt.Sprintf("wss://.+\\.execute-api\\.%s.amazonaws\\.com/%s", acctest.Region(), rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
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
					testAccCheckStageExists(resourceName, &apiId, &v),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_deploy", "false"),
					resource.TestCheckResourceAttr(resourceName, "client_certificate_id", ""),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.data_trace_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.detailed_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.logging_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_burst_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_route_settings.0.throttling_rate_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "route_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckStageDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_stage" {
			continue
		}

		_, err := conn.GetStage(&apigatewayv2.GetStageInput{
			ApiId:     aws.String(rs.Primary.Attributes["api_id"]),
			StageName: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 stage %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckStageExists(n string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 stage ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		apiId := aws.String(rs.Primary.Attributes["api_id"])
		resp, err := conn.GetStage(&apigatewayv2.GetStageInput{
			ApiId:     apiId,
			StageName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*vApiId = *apiId
		*v = *resp

		return nil
	}
}

func testAccCheckStageARN(resourceName, attributeName string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, attributeName, "apigateway", fmt.Sprintf("/apis/%s/stages/%s", *vApiId, *v.StageName))(s)
	}
}

func testAccCheckStageExecutionARN(resourceName, attributeName string, vApiId *string, v *apigatewayv2.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARN(resourceName, attributeName, "execute-api", fmt.Sprintf("%s/%s", *vApiId, *v.StageName))(s)
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

func testAccStageConfig_tags(rName string) string {
	return acctest.ConfigCompose(
		testAccStageConfig_apiWebSocket(rName),
		fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id = aws_apigatewayv2_api.test.id
  name   = %[1]q

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName))
}

// testAccPreCheckAPIGatewayAccountCloudWatchRoleARN checks whether a CloudWatch role ARN has been configured in the current AWS region.
func testAccPreCheckAPIGatewayAccountCloudWatchRoleARN(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	output, err := conn.GetAccount(&apigateway.GetAccountInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping tests: %s", err)
	}

	if err != nil {
		t.Fatalf("error reading API Gateway Account: %s", err)
	}

	if output == nil || aws.StringValue(output.CloudwatchRoleArn) == "" {
		t.Skip("skipping tests; no API Gateway CloudWatch role ARN has been configured in this region")
	}
}
