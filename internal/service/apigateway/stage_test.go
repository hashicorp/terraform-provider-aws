// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
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

func testAccStage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Hello world"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "variables.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "variables.one", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "variables.three", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "variables.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccStage_cache(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_cache(rName, "0.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtTrue),
				),
			},

			{
				Config: testAccStageConfig_cache(rName, "1.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "1.6"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/22866
func testAccStage_cacheSizeCacheDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_cache(rName, "0.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_cache(rName, "6.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "6.1"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", ""),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccStageConfig_cacheSizeCacheDisabled(rName, "28.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "28.4"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccStage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stage apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &stage),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccStage_Disappears_restAPI(t *testing.T) {
	ctx := acctest.Context(t)
	var stage apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &stage),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPI(), "aws_api_gateway_rest_api.test"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccStage_accessLogSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_api_gateway_stage.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettings(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccStageConfig_accessLogSettings(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettings(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettings(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccStage_AccessLogSettings_kinesis(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"
	kinesesResourceName := "aws_kinesis_firehose_delivery_stream.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccStage_waf(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_wafACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
				),
			},
			{
				Config: testAccStageConfig_wafACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_arn", "aws_wafregional_web_acl.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStage_canarySettings(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetStageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_canarySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "variables.one", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.percent_traffic", "33.33"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.stage_variable_overrides.one", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.use_stage_cache", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", acctest.Ct0),
				),
			},
			{
				Config: testAccStageConfig_canarySettingsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "variables.one", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.percent_traffic", "66.66"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.stage_variable_overrides.four", "5"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.use_stage_cache", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckStageExists(ctx context.Context, n string, v *apigateway.GetStageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindStageByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_stage" {
				continue
			}

			_, err := tfapigateway.FindStageByTwoPartKey(ctx, conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Stage %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStageImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"]), nil
	}
}

func testAccStageConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccAccountConfig_role0(rName), fmt.Sprintf(`
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
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.co.uk"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "dev"
  description = "This is a dev env"

  variables = {
    "a" = "2"
  }
}
`, rName))
}

func testAccStageConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id
}
`)
}

func testAccStageConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), `
resource "aws_api_gateway_stage" "test" {
  rest_api_id          = aws_api_gateway_rest_api.test.id
  stage_name           = "prod"
  deployment_id        = aws_api_gateway_deployment.test.id
  description          = "Hello world"
  xray_tracing_enabled = true

  variables = {
    one   = "1"
    three = "3"
  }
}
`)
}

func testAccStageConfig_cacheSizeCacheDisabled(rName, size string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id        = aws_api_gateway_rest_api.test.id
  stage_name         = "prod"
  deployment_id      = aws_api_gateway_deployment.test.id
  cache_cluster_size = %[1]q
}
`, size))
}

func testAccStageConfig_cache(rName, size string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.test.id
  cache_cluster_enabled = true
  cache_cluster_size    = %[1]q
}
`, size))
}

func testAccStageConfig_accessLogSettings(rName, format string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.test.arn
    format          = %[2]q
  }
}
`, rName, format))
}

func testAccStageConfig_accessLogSettingsKinesis(rName, format string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test_firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "amazon-apigateway-%[1]s"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test_firehose.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id

  access_log_settings {
    destination_arn = aws_kinesis_firehose_delivery_stream.test.arn
    format          = %[2]q
  }
}
`, rName, format))
}

func testAccStageConfig_wafACL(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_basic(rName), fmt.Sprintf(`
resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = "test"
  default_action {
    type = "ALLOW"
  }
}

resource "aws_wafregional_web_acl_association" "test" {
  resource_arn = aws_api_gateway_stage.test.arn
  web_acl_id   = aws_wafregional_web_acl.test.id
}
`, rName))
}

func testAccStageConfig_canarySettings(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id

  canary_settings {
    percent_traffic = "33.33"
    stage_variable_overrides = {
      one = "3"
    }
    use_stage_cache = "true"
  }
  variables = {
    one = "1"
    two = "2"
  }
}
`)
}

func testAccStageConfig_canarySettingsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id

  canary_settings {
    percent_traffic = "66.66"
    stage_variable_overrides = {
      four = "5"
    }
    use_stage_cache = "false"
  }
  variables = {
    one = "1"
    two = "2"
  }
}
`)
}
