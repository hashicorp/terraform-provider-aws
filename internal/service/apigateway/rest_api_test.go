package apigateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
)

func TestAccAPIGatewayRestAPI_basic(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPINameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "body"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexp.MustCompile(`[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "-1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "root_resource_id", regexp.MustCompile(`[a-z0-9]+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAPIGatewayRestAPI_tags(t *testing.T) {
	var conf apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPITags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				Config: testAccRestAPITags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},

			{
				Config: testAccRestAPITags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_disappears(t *testing.T) {
	var restApi apigateway.RestApi
	resourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPINameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					acctest.CheckResourceDisappears(acctest.Provider, tfapigateway.ResourceRestAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_endpoint(t *testing.T) {
	var restApi apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_EndpointConfiguration(rName, "REGIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "REGIONAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// For backwards compatibility, test removing endpoint_configuration, which should do nothing
			{
				Config: testAccRestAPIConfig_Name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "REGIONAL"),
				),
			},
			// Test updating endpoint type
			{
				PreConfig: func() {
					// Ensure region supports EDGE endpoint
					// This can eventually be moved to a PreCheck function
					// If the region does not support EDGE endpoint type, this test will either show
					// SKIP (if REGIONAL passed) or FAIL (if REGIONAL failed)
					conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn
					output, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
						Name: aws.String(sdkacctest.RandomWithPrefix("tf-acc-test-edge-endpoint-precheck")),
						EndpointConfiguration: &apigateway.EndpointConfiguration{
							Types: []*string{aws.String("EDGE")},
						},
					})
					if err != nil {
						if tfawserr.ErrMessageContains(err, apigateway.ErrCodeBadRequestException, "Endpoint Configuration type EDGE is not supported in this region") {
							t.Skip("Region does not support EDGE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccRestAPIConfig_EndpointConfiguration(rName, "EDGE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "EDGE"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Endpoint_private(t *testing.T) {
	var restApi apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Ensure region supports PRIVATE endpoint
					// This can eventually be moved to a PreCheck function
					conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn
					output, err := conn.CreateRestApi(&apigateway.CreateRestApiInput{
						Name: aws.String(sdkacctest.RandomWithPrefix("tf-acc-test-private-endpoint-precheck")),
						EndpointConfiguration: &apigateway.EndpointConfiguration{
							Types: []*string{aws.String("PRIVATE")},
						},
					})
					if err != nil {
						if tfawserr.ErrMessageContains(err, apigateway.ErrCodeBadRequestException, "Endpoint Configuration type PRIVATE is not supported in this region") {
							t.Skip("Region does not support PRIVATE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(&apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccRestAPIConfig_EndpointConfiguration(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
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

func TestAccAPIGatewayRestAPI_apiKeySource(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIAPIKeySourceConfig(rName, "AUTHORIZER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRestAPIAPIKeySourceConfig(rName, "HEADER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			{
				Config: testAccRestAPINameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_APIKeySource_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIAPIKeySourceOverrideBodyConfig(rName, "AUTHORIZER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated API key source still overrides
			{
				Config: testAccRestAPIAPIKeySourceOverrideBodyConfig(rName, "HEADER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			// Verify updated body API key source is still overridden
			{
				Config: testAccRestAPIAPIKeySourceOverrideBodyConfig(rName, "HEADER", "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_APIKeySource_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIAPIKeySourceSetByBodyConfig(rName, "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
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

func TestAccAPIGatewayRestAPI_binaryMediaTypes(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIBinaryMediaTypes1Config(rName, "application/octet-stream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccRestAPIBinaryMediaTypes1Config(rName, "application/octet"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_BinaryMediaTypes_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIBinaryMediaTypes1OverrideBodyConfig(rName, "application/octet-stream", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccRestAPIBinaryMediaTypes1OverrideBodyConfig(rName, "application/octet", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccRestAPIBinaryMediaTypes1OverrideBodyConfig(rName, "application/octet", "image/png"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_BinaryMediaTypes_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIBinaryMediaTypes1SetByBodyConfig(rName, "application/octet-stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
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

func TestAccAPIGatewayRestAPI_body(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			// The body is expected to only set a title (name) and one route
			{
				Config: testAccRestAPIBodyConfig(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/test"}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "binary_media_types"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccRestAPIBodyConfig(rName, "/update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/update"}),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_description(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDescriptionConfig(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccRestAPIDescriptionConfig(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Description_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDescriptionOverrideBodyConfig(rName, "tfdescription1", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated description still overrides
			{
				Config: testAccRestAPIDescriptionOverrideBodyConfig(rName, "tfdescription2", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription2"),
				),
			},
			// Verify updated body description is still overridden
			{
				Config: testAccRestAPIDescriptionOverrideBodyConfig(rName, "tfdescription2", "oasdescription2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "tfdescription2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Description_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDescriptionSetByBodyConfig(rName, "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "oasdescription1"),
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

func TestAccAPIGatewayRestAPI_disableExecuteAPIEndpoint(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `false`),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `true`),
				),
			},
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", `false`),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_DisableExecuteAPIEndpoint_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointOverrideBodyConfig(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify override can be unset (only for body set to false)
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointOverrideBodyConfig(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "false"),
				),
			},
			// Verify override can be reset
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointOverrideBodyConfig(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_DisableExecuteAPIEndpoint_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDisableExecuteAPIEndpointSetByBodyConfig(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", "true"),
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

func TestAccAPIGatewayRestAPI_Endpoint_vpcEndpointIDs(t *testing.T) {
	var restApi apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIds1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIds2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName2, "id"),
				),
			},
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIds1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test.0"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test.1"
	vpcEndpointResourceName3 := "aws_vpc_endpoint.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIdsOverrideBodyConfig(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated configuration value still overrides
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIdsOverrideBodyConfig(rName, vpcEndpointResourceName3, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, "id"),
				),
			},
			// Verify updated body value is still overridden
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIdsOverrideBodyConfig(rName, vpcEndpointResourceName3, vpcEndpointResourceName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, "id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIEndpointConfigurationVPCEndpointIdsSetByBodyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName, "id"),
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

func TestAccAPIGatewayRestAPI_minimumCompressionSize(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIMinimumCompressionSizeConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			{
				Config: testAccRestAPIMinimumCompressionSizeConfig(rName, -1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "-1"),
				),
			},
			{
				Config: testAccRestAPIMinimumCompressionSizeConfig(rName, 5242880),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "5242880"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_MinimumCompressionSize_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIMinimumCompressionSizeOverrideBodyConfig(rName, 1, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccRestAPIMinimumCompressionSizeOverrideBodyConfig(rName, 2, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "2"),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccRestAPIMinimumCompressionSizeOverrideBodyConfig(rName, 2, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_MinimumCompressionSize_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIMinimumCompressionSizeSetByBodyConfig(rName, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "1048576"),
				),
				// TODO: The attribute type must be changed to NullableTypeInt so it can be Computed properly.
				ExpectNonEmptyPlan: true,
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

func TestAccAPIGatewayRestAPI_Name_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPINameOverrideBodyConfig(rName, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body"},
			},
			// Verify updated name still overrides
			{
				Config: testAccRestAPINameOverrideBodyConfig(rName2, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
			// Verify updated title still overrides
			{
				Config: testAccRestAPINameOverrideBodyConfig(rName2, "title2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_parameters(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIParameters1Config(rName, "basepath", "prepend"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/foo", "/foo/bar", "/foo/bar/baz", "/foo/bar/baz/test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "parameters"},
			},
			{
				Config: testAccRestAPIParameters1Config(rName, "basepath", "ignore"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/test"}),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_basic(t *testing.T) {
	resourceName := "aws_api_gateway_rest_api.test"
	expectedPolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Condition":{"IpAddress":{"aws:SourceIp":"123.123.123.123/32"}},"Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	expectedUpdatePolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Effect":"Deny","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIWithPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", expectedPolicyText),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy"},
			},
			{
				Config: testAccRestAPIUpdatePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", expectedUpdatePolicyText),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_order(t *testing.T) {
	resourceName := "aws_api_gateway_rest_api.test"
	expectedPolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Condition":{"IpAddress":{"aws:SourceIp":["123.123.123.123/32","122.122.122.122/32","169.254.169.253/32"]}},"Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIWithPolicyOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "policy", expectedPolicyText),
				),
			},
			{
				Config:   testAccRestAPIWithPolicyNewOrderConfig(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_overrideBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicyOverrideBodyConfig(rName, "/test", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/test"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "policy"},
			},
			// Verify updated body still has override policy
			{
				Config: testAccRestAPIPolicyOverrideBodyConfig(rName, "/test2", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
				),
			},
			// Verify updated policy still overrides body
			{
				Config: testAccRestAPIPolicyOverrideBodyConfig(rName, "/test2", "Deny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					testAccCheckRestAPIRoutes(&conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Deny"`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_setByBody(t *testing.T) {
	var conf apigateway.RestApi
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRestAPIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIPolicySetByBodyConfig(rName, "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRestAPIExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`"Allow"`)),
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

func testAccCheckRestAPIRoutes(conf *apigateway.RestApi, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

		resp, err := conn.GetResources(&apigateway.GetResourcesInput{
			RestApiId: conf.Id,
		})
		if err != nil {
			return err
		}

		actualRoutePaths := map[string]bool{}
		for _, resource := range resp.Items {
			actualRoutePaths[*resource.Path] = true
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

func testAccCheckRestAPIExists(n string, res *apigateway.RestApi) resource.TestCheckFunc {
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

		if *describe.Id != rs.Primary.ID {
			return fmt.Errorf("APIGateway not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckRestAPIDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_rest_api" {
			continue
		}

		req := &apigateway.GetRestApisInput{}
		describe, err := conn.GetRestApis(req)

		if err == nil {
			if len(describe.Items) != 0 &&
				*describe.Items[0].Id == rs.Primary.ID {
				return fmt.Errorf("API Gateway still exists")
			}
		}

		return err
	}

	return nil
}

func testAccRestAPIConfig_EndpointConfiguration(rName, endpointType string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  endpoint_configuration {
    types = ["%s"]
  }
}
`, rName, endpointType)
}

func testAccRestAPIDisableExecuteAPIEndpointConfig(rName string, disableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  disable_execute_api_endpoint = %[2]t
  name                         = %[1]q
}
`, rName, disableExecuteApiEndpoint)
}

func testAccRestAPIDisableExecuteAPIEndpointOverrideBodyConfig(rName string, configDisableExecuteApiEndpoint bool, bodyDisableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  disable_execute_api_endpoint = %[2]t
  name                         = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      disableExecuteApiEndpoint = %[3]t
    }
  })
}
`, rName, configDisableExecuteApiEndpoint, bodyDisableExecuteApiEndpoint)
}

func testAccRestAPIDisableExecuteAPIEndpointSetByBodyConfig(rName string, bodyDisableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      disableExecuteApiEndpoint = %[2]t
    }
  })
}
`, rName, bodyDisableExecuteApiEndpoint)
}

func testAccRestAPIConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"
}
`, rName)
}

func testAccRestAPIEndpointConfigurationVPCEndpointIds1Config(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id]
  }
}
`, rName))
}

func testAccRestAPIEndpointConfigurationVPCEndpointIds2Config(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_vpc_endpoint" "test2" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [aws_vpc_endpoint.test.id, aws_vpc_endpoint.test2.id]
  }
}
`, rName))
}

func testAccRestAPIEndpointConfigurationVPCEndpointIdsOverrideBodyConfig(rName string, configVpcEndpointResourceName string, bodyVpcEndpointResourceName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  count = 3

  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types            = ["PRIVATE"]
    vpc_endpoint_ids = [%[2]s]
  }

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      vpcEndpointIds = [%[3]s]
    }
  })
}
`, rName, configVpcEndpointResourceName+".id", bodyVpcEndpointResourceName+".id"))
}

func testAccRestAPIEndpointConfigurationVPCEndpointIdsSetByBodyConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  private_dns_enabled = false
  security_group_ids  = [aws_default_security_group.test.id]
  service_name        = "com.amazonaws.${data.aws_region.current.name}.execute-api"
  subnet_ids          = [aws_subnet.test.id]
  vpc_endpoint_type   = "Interface"
  vpc_id              = aws_vpc.test.id
}

resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  endpoint_configuration {
    types = ["PRIVATE"]
  }

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-endpoint-configuration = {
      vpcEndpointIds = [aws_vpc_endpoint.test.id]
    }
  })
}
`, rName))
}

func testAccRestAPITags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRestAPITags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRestAPIWithPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "execute-api:Invoke"
      Condition = {
        IpAddress = {
          "aws:SourceIp" = "123.123.123.123/32"
        }
      }
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccRestAPIUpdatePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "execute-api:Invoke"
      Effect = "Deny"
      Principal = {
        AWS = "*"
      }
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccRestAPIWithPolicyOrderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "execute-api:Invoke"
      Condition = {
        IpAddress = {
          "aws:SourceIp" = [
            "123.123.123.123/32",
            "122.122.122.122/32",
            "169.254.169.253/32",
          ]
        }
      }
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccRestAPIWithPolicyNewOrderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "execute-api:Invoke"
      Condition = {
        IpAddress = {
          "aws:SourceIp" = [
            "122.122.122.122/32",
            "169.254.169.253/32",
            "123.123.123.123/32",
          ]
        }
      }
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Resource = "*"
    }]
  })
}
`, rName)
}

func testAccRestAPIAPIKeySourceConfig(rName string, apiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  api_key_source = %[2]q
  name           = %[1]q
}
`, rName, apiKeySource)
}

func testAccRestAPIAPIKeySourceOverrideBodyConfig(rName string, apiKeySource string, bodyApiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  api_key_source = %[2]q
  name           = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-api-key-source = %[3]q
  })
}
`, rName, apiKeySource, bodyApiKeySource)
}

func testAccRestAPIAPIKeySourceSetByBodyConfig(rName string, bodyApiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-api-key-source = %[2]q
  })
}
`, rName, bodyApiKeySource)
}

func testAccRestAPIBinaryMediaTypes1Config(rName string, binaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  binary_media_types = [%[2]q]
  name               = %[1]q
}
`, rName, binaryMediaTypes1)
}

func testAccRestAPIBinaryMediaTypes1OverrideBodyConfig(rName string, binaryMediaTypes1 string, bodyBinaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  binary_media_types = [%[2]q]
  name               = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-binary-media-types = [%[3]q]
  })
}
`, rName, binaryMediaTypes1, bodyBinaryMediaTypes1)
}

func testAccRestAPIBinaryMediaTypes1SetByBodyConfig(rName string, bodyBinaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-binary-media-types = [%[2]q]
  })
}
`, rName, bodyBinaryMediaTypes1)
}

func testAccRestAPIBodyConfig(rName string, basePath string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      %[2]q = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })
}
`, rName, basePath)
}

func testAccRestAPIDescriptionConfig(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccRestAPIDescriptionOverrideBodyConfig(rName string, description string, bodyDescription string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  description = %[2]q
  name        = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      description = %[3]q
      title       = "test"
      version     = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })
}
`, rName, description, bodyDescription)
}

func testAccRestAPIDescriptionSetByBodyConfig(rName string, bodyDescription string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      description = %[2]q
      title       = "test"
      version     = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })
}
`, rName, bodyDescription)
}

func testAccRestAPIMinimumCompressionSizeConfig(rName string, minimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  minimum_compression_size = %[2]d
  name                     = %[1]q
}
`, rName, minimumCompressionSize)
}

func testAccRestAPIMinimumCompressionSizeOverrideBodyConfig(rName string, minimumCompressionSize int, bodyMinimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  minimum_compression_size = %[2]d
  name                     = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-minimum-compression-size = %[3]d
  })
}
`, rName, minimumCompressionSize, bodyMinimumCompressionSize)
}

func testAccRestAPIMinimumCompressionSizeSetByBodyConfig(rName string, bodyMinimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-minimum-compression-size = %[2]d
  })
}
`, rName, bodyMinimumCompressionSize)
}

func testAccRestAPINameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRestAPINameOverrideBodyConfig(rName string, title string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = %[2]q
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })
}
`, rName, title)
}

func testAccRestAPIParameters1Config(rName string, parameterKey1 string, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes  = ["https"]
    basePath = "/foo/bar/baz"
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })

  parameters = {
    %[2]s = %[3]q
  }
}
`, rName, parameterKey1, parameterValue1)
}

func testAccRestAPIPolicyOverrideBodyConfig(rName string, bodyPath string, policyEffect string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      %[2]q = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
  })

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "execute-api:Invoke"
        Condition = {
          IpAddress = {
            "aws:SourceIp" = "123.123.123.123/32"
          }
        }
        Effect = %[3]q
        Principal = {
          AWS = "*"
        }
        Resource = "*"
      }
    ]
  })
}
`, rName, bodyPath, policyEffect)
}

func testAccRestAPIPolicySetByBodyConfig(rName string, bodyPolicyEffect string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q

  body = jsonencode({
    swagger = "2.0"
    info = {
      title   = "test"
      version = "2017-04-20T04:08:08Z"
    }
    schemes = ["https"]
    paths = {
      "/test" = {
        get = {
          responses = {
            "200" = {
              description = "OK"
            }
          }
          x-amazon-apigateway-integration = {
            httpMethod = "GET"
            type       = "HTTP"
            responses = {
              default = {
                statusCode = 200
              }
            }
            uri = "https://api.example.com/"
          }
        }
      }
    }
    x-amazon-apigateway-policy = {
      Version = "2012-10-17"
      Statement = [
        {
          Action = "execute-api:Invoke"
          Condition = {
            IpAddress = {
              "aws:SourceIp" = "123.123.123.123/32"
            }
          }
          Effect = %[2]q
          Principal = {
            AWS = "*"
          }
          Resource = "*"
        }
      ]
    }
  })
}
`, rName, bodyPolicyEffect)
}
