package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAPIGatewayStage_basic(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, "description", "Hello world"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "variables.one", "1"),
					resource.TestCheckResourceAttr(resourceName, "variables.three", "3"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "true"),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "stage_name", "prod"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "invoke_url"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "xray_tracing_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAPIGatewayStage_cache(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfigCacheConfig(rName, "0.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "true"),
				),
			},

			{
				Config: testAccStageConfigCacheConfig(rName, "1.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "1.6"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "true"),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "false"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/22866
func TestAccAPIGatewayStage_cache_size_cache_disabled(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfigCacheSizeCacheDisabled(rName, "0.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "false"),
				),
			},
			{
				Config: testAccStageConfigCacheConfig(rName, "0.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_size", "0.5"),
					resource.TestCheckResourceAttr(resourceName, "cache_cluster_enabled", "true"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12756
func TestAccAPIGatewayStage_Disappears_referencingDeployment(t *testing.T) {
	var stage apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageReferencingDeploymentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage),
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

func TestAccAPIGatewayStage_tags(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccStageImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccStageConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStageConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayStage_disappears(t *testing.T) {
	var stage apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageReferencingDeploymentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceStage(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayStage_disappears_restApi(t *testing.T) {
	var stage apigateway.Stage
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageReferencingDeploymentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &stage),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRestAPI(), "aws_api_gateway_rest_api.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceStage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayStage_accessLogSettings(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_api_gateway_stage.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettings(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccStageConfig_accessLogSettings(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettings(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettings(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayStage_AccessLogSettings_kinesis(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"
	kinesesResourceName := "aws_kinesis_firehose_delivery_stream.test"
	clf := `$context.identity.sourceIp $context.identity.caller $context.identity.user [$context.requestTime] "$context.httpMethod $context.resourcePath $context.protocol" $context.status $context.responseLength $context.requestId`
	json := `{ "requestId":"$context.requestId", "ip": "$context.identity.sourceIp", "caller":"$context.identity.caller", "user":"$context.identity.user", "requestTime":"$context.requestTime", "httpMethod":"$context.httpMethod", "resourcePath":"$context.resourcePath", "status":"$context.status", "protocol":"$context.protocol", "responseLength":"$context.responseLength" }`
	xml := `<request id="$context.requestId"> <ip>$context.identity.sourceIp</ip> <caller>$context.identity.caller</caller> <user>$context.identity.user</user> <requestTime>$context.requestTime</requestTime> <httpMethod>$context.httpMethod</httpMethod> <resourcePath>$context.resourcePath</resourcePath> <status>$context.status</status> <protocol>$context.protocol</protocol> <responseLength>$context.responseLength</responseLength> </request>`
	csv := `$context.identity.sourceIp,$context.identity.caller,$context.identity.user,$context.requestTime,$context.httpMethod,$context.resourcePath,$context.protocol,$context.status,$context.responseLength,$context.requestId`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, clf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", clf),
				),
			},

			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, json),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", json),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, xml),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", xml),
				),
			},
			{
				Config: testAccStageConfig_accessLogSettingsKinesis(rName, csv),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "access_log_settings.0.destination_arn", kinesesResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.0.format", csv),
				),
			},
			{
				Config: testAccStageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/.+/stages/prod`)),
					resource.TestCheckResourceAttr(resourceName, "access_log_settings.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayStage_waf(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_wafACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
				),
			},
			{
				Config: testAccStageConfig_wafACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_arn", "aws_wafregional_web_acl.test", "arn"),
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

func TestAccAPIGatewayStage_canarySettings(t *testing.T) {
	var conf apigateway.Stage
	rName := sdkacctest.RandString(5)
	resourceName := "aws_api_gateway_stage.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStageConfig_canarySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "variables.one", "1"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.percent_traffic", "33.33"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.stage_variable_overrides.one", "3"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.use_stage_cache", "true"),
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
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.#", "0"),
				),
			},
			{
				Config: testAccStageConfig_canarySettingsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStageExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "variables.one", "1"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.percent_traffic", "66.66"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.stage_variable_overrides.four", "5"),
					resource.TestCheckResourceAttr(resourceName, "canary_settings.0.use_stage_cache", "false"),
				),
			},
		},
	})
}

func testAccCheckStageExists(n string, res *apigateway.Stage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway Stage ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn
		out, err := tfapigateway.FindStageByName(conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"])
		if err != nil {
			return err
		}

		if out == nil {
			return fmt.Errorf("API Gateway Stage not found")
		}

		*res = *out

		return nil
	}
}

func testAccCheckStageDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_stage" {
			continue
		}

		_, err := tfapigateway.FindStageByName(conn, rs.Primary.Attributes["rest_api_id"], rs.Primary.Attributes["stage_name"])
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
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "tf-acc-test-%s"
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

resource "aws_api_gateway_method_response" "error" {
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
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "dev" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "dev"
  description = "This is a dev env"

  variables = {
    "a" = "2"
  }
}
`, rName)
}

func testAccStageBaseDeploymentStageNameConfig(rName string, stageName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "test" {
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
  rest_api_id = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method" "test" {
  authorization = "NONE"
  http_method   = "GET"
  resource_id   = aws_api_gateway_resource.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
}

resource "aws_api_gateway_method_response" "error" {
  http_method = aws_api_gateway_method.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "GET"
  resource_id             = aws_api_gateway_resource.test.id
  rest_api_id             = aws_api_gateway_rest_api.test.id
  type                    = "HTTP"
  uri                     = "https://www.google.co.uk"
}

resource "aws_api_gateway_integration_response" "test" {
  http_method = aws_api_gateway_integration.test.http_method
  resource_id = aws_api_gateway_resource.test.id
  rest_api_id = aws_api_gateway_rest_api.test.id
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = %[2]q
}
`, rName, stageName)
}

func testAccStageReferencingDeploymentConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccStageBaseDeploymentStageNameConfig(rName, ""),
		`
# Due to oddities with API Gateway, certain environments utilize
# a double deployment configuration. This can cause destroy errors.
resource "aws_api_gateway_stage" "test" {
  deployment_id = aws_api_gateway_deployment.test.id
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "test"

  lifecycle {
    ignore_changes = [deployment_id]
  }
}

resource "aws_api_gateway_deployment" "test2" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = aws_api_gateway_stage.test.stage_name
}
`)
}

func testAccStageConfig_basic(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id
}
`
}

func testAccStageConfigUpdated(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id          = aws_api_gateway_rest_api.test.id
  stage_name           = "prod"
  deployment_id        = aws_api_gateway_deployment.dev.id
  description          = "Hello world"
  xray_tracing_enabled = true

  variables = {
    one   = "1"
    three = "3"
  }
}
`
}

func testAccStageConfigCacheSizeCacheDisabled(rName, size string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id        = aws_api_gateway_rest_api.test.id
  stage_name         = "prod"
  deployment_id      = aws_api_gateway_deployment.dev.id
  cache_cluster_size = %[1]q
}
`, size)
}

func testAccStageConfigCacheConfig(rName, size string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id           = aws_api_gateway_rest_api.test.id
  stage_name            = "prod"
  deployment_id         = aws_api_gateway_deployment.dev.id
  cache_cluster_enabled = true
  cache_cluster_size    = %[1]q
}
`, size)
}

func testAccStageConfig_accessLogSettings(rName string, format string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%s"
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.test.arn
    format          = %q
  }
}
`, rName, format)
}

func testAccStageConfig_accessLogSettingsKinesis(rName string, format string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = "%[1]s"

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
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}

resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

  access_log_settings {
    destination_arn = aws_kinesis_firehose_delivery_stream.test.arn
    format          = %q
  }
}
`, rName, format)
}

func testAccStageConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccStageConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccStageConfig_base(rName) + fmt.Sprintf(`
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccStageConfig_wafACL(rName string) string {
	return testAccStageConfig_basic(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccStageConfig_canarySettings(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

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
`
}

func testAccStageConfig_canarySettingsUpdated(rName string) string {
	return testAccStageConfig_base(rName) + `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.dev.id

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
`
}
