package apigatewayv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
)

func TestAccAPIGatewayV2API_basicWebSocket(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
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
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
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
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigatewayv2.ResourceAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2API_allAttributesWebSocket(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAPIConfig_basicWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesWebSocket(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$context.authorizer.usageIdentifierKey"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.service"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
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
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAPIConfig_basicHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
				),
			},
			{
				Config: testAccAPIConfig_allAttributesHTTP(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "v1"),
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
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_OpenAPI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_UpdatedOpenAPIYAML(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withTags(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_OpenAPIYAML_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /test"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_OpenAPIYAML_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1U"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2U"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withCors(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_OpenAPIYAML_corsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "delete"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_OpenAPIYAML_corsConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
				),
			},
			{
				Config: testAccAPIConfig_OpenAPIYAML_corsConfigurationUpdated2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "get"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.google.de"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_OpenAPI_withMoreFields(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_OpenAPIYAML(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccAPIConfig_UpdatedOpenAPI2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "description test"),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "version", "2017-04-21T04:08:08Z"),
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
				),
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
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			// Invalid body should not be accepted when fail_on_warnings is enabled
			{
				Config:      testAccAPIConfig_FailOnWarnings(rName, "fail_on_warnings = true"),
				ExpectError: regexp.MustCompile(`BadRequestException: Warnings found during import`),
			},
			// Warnings do not break the deployment when fail_on_warnings is disabled
			{
				Config: testAccAPIConfig_FailOnWarnings(rName, "fail_on_warnings = false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "fail_on_warnings", "false"),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "fail_on_warnings"},
			},
			// fail_on_warnings should be optional and false by default
			{
				Config: testAccAPIConfig_FailOnWarnings(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					testAccCheckAPIRoutes(&v, []string{"GET /update"}),
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

func testAccCheckAPIRoutes(v *apigatewayv2.GetApiOutput, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetRoutes(&apigatewayv2.GetRoutesInput{
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

func TestAccAPIGatewayV2API_tags(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIConfig_basicWebSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeWebsocket),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.body.action"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_cors(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_corsConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", "false"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_headers.*", "Authorization"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "put"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIConfig_corsConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_headers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_methods.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_methods.*", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.allow_origins.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "HTTP://WWW.EXAMPLE.ORG"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.allow_origins.*", "https://example.io"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.expose_headers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_configuration.0.expose_headers.*", "X-Api-Id"),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.0.max_age", "500"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				Config: testAccAPIConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2API_quickCreate(t *testing.T) {
	var v apigatewayv2.GetApiOutput
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIConfig_quickCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIExists(resourceName, &v),
					testAccCheckAPIQuickCreateIntegration(resourceName, "HTTP_PROXY", "http://www.example.com/"),
					testAccCheckAPIQuickCreateRoute(resourceName, "GET /pets"),
					testAccCheckAPIQuickCreateStage(resourceName, "$default"),
					resource.TestCheckResourceAttrSet(resourceName, "api_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "api_key_selection_expression", "$request.header.x-api-key"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/apis/.+`)),
					resource.TestCheckResourceAttr(resourceName, "cors_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocol_type", apigatewayv2.ProtocolTypeHttp),
					resource.TestCheckResourceAttr(resourceName, "route_key", "GET /pets"),
					resource.TestCheckResourceAttr(resourceName, "route_selection_expression", "$request.method $request.path"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target", "http://www.example.com/"),
					resource.TestCheckResourceAttr(resourceName, "version", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"route_key",
					"target",
				},
			},
		},
	})
}

func testAccCheckAPIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_api" {
			continue
		}

		_, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 API %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAPIExists(n string, v *apigatewayv2.GetApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccCheckAPIQuickCreateIntegration(n, expectedType, expectedUri string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetIntegrations(&apigatewayv2.GetIntegrationsInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of integrations: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].IntegrationType); got != expectedType {
			return fmt.Errorf("Incorrect integration type. Expected: %s, got: %s", expectedType, got)
		}
		if got := aws.StringValue(resp.Items[0].IntegrationUri); got != expectedUri {
			return fmt.Errorf("Incorrect integration URI. Expected: %s, got: %s", expectedUri, got)
		}

		return nil
	}
}

func testAccCheckAPIQuickCreateRoute(n, expectedRouteKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetRoutes(&apigatewayv2.GetRoutesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of routes: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].RouteKey); got != expectedRouteKey {
			return fmt.Errorf("Incorrect route key. Expected: %s, got: %s", expectedRouteKey, got)
		}

		return nil
	}
}

func testAccCheckAPIQuickCreateStage(n, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 API ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayV2Conn

		resp, err := conn.GetStages(&apigatewayv2.GetStagesInput{
			ApiId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if got := len(resp.Items); got != 1 {
			return fmt.Errorf("Incorrect number of stages: %d", got)
		}

		if got := aws.StringValue(resp.Items[0].StageName); got != expectedName {
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

func testAccAPIConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name                       = %[1]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
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

func testAccAPIConfig_OpenAPI(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%s_DIFFERENT",
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML_corsConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["delete"]
    allow_origins = ["https://www.google.de"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML_corsConfigurationUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML_corsConfigurationUpdated2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  cors_configuration {
    allow_methods = ["put", "get"]
    allow_origins = ["https://www.google.de", "https://www.example.com"]
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_OpenAPIYAML_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  tags = {
    Key1 = "Value1U"
    Key2 = "Value2U"
  }
  body = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
  version: 2.0
tags:
  - name: Key3
    x-amazon-apigateway-tag-value: Value3
  - name: Key4
    x-amazon-apigateway-tag-value: Value3
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
`, rName, rName)
}

func testAccAPIConfig_UpdatedOpenAPIYAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  body          = <<EOF
---
openapi: 3.0.1
info:
  title: %s_DIFFERENT
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
`, rName, rName)
}

func testAccAPIConfig_UpdatedOpenAPI2(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = "%s"
  protocol_type = "HTTP"
  version       = "2017-04-21T04:08:08Z"
  description   = "description test"
  body          = <<EOF
{
  "openapi": "3.0.1",
  "info": {
    "title": "%s_DIFFERENT",
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
`, rName, rName)
}

func testAccAPIConfig_FailOnWarnings(rName string, failOnWarnings string) string {
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
