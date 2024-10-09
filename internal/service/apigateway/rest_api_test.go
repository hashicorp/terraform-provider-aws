// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayRestAPI_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/restapis/+.`)),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "body"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "execution_arn", "execute-api", regexache.MustCompile(`[0-9a-z]+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "root_resource_id", regexache.MustCompile(`[0-9a-z]+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var restApi apigateway.GetRestApiOutput
	resourceName := "aws_api_gateway_rest_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceRestAPI(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var restApi apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_endpointConfiguration(rName, "REGIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "REGIONAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"put_rest_api_mode"},
			},
			// For backwards compatibility, test removing endpoint_configuration, which should do nothing
			{
				Config: testAccRestAPIConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
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
					conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)
					output, err := conn.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
						Name: aws.String(sdkacctest.RandomWithPrefix("tf-acc-test-edge-endpoint-precheck")),
						EndpointConfiguration: &types.EndpointConfiguration{
							Types: []types.EndpointType{types.EndpointTypeEdge},
						},
					})
					if err != nil {
						if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Endpoint Configuration type EDGE is not supported in this region") {
							t.Skip("Region does not support EDGE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccRestAPIConfig_endpointConfiguration(rName, "EDGE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "EDGE"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Endpoint_private(t *testing.T) {
	ctx := acctest.Context(t)
	var restApi apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Ensure region supports PRIVATE endpoint
					// This can eventually be moved to a PreCheck function
					conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)
					output, err := conn.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
						Name: aws.String(sdkacctest.RandomWithPrefix("tf-acc-test-private-endpoint-precheck")),
						EndpointConfiguration: &types.EndpointConfiguration{
							Types: []types.EndpointType{types.EndpointTypePrivate},
						},
					})
					if err != nil {
						if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Endpoint Configuration type PRIVATE is not supported in this region") {
							t.Skip("Region does not support PRIVATE endpoint type")
						}
						t.Fatal(err)
					}

					// Be kind and rewind. :)
					_, err = conn.DeleteRestApi(ctx, &apigateway.DeleteRestApiInput{
						RestApiId: output.Id,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccRestAPIConfig_endpointConfiguration(rName, "PRIVATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_apiKeySource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_keySource(rName, "AUTHORIZER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_keySource(rName, "HEADER"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			{
				Config: testAccRestAPIConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_APIKeySource_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_keySourceOverrideBody(rName, "AUTHORIZER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated API key source still overrides
			{
				Config: testAccRestAPIConfig_keySourceOverrideBody(rName, "HEADER", "HEADER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
			// Verify updated body API key source is still overridden
			{
				Config: testAccRestAPIConfig_keySourceOverrideBody(rName, "HEADER", "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "HEADER"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_APIKeySource_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_keySourceSetByBody(rName, "AUTHORIZER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "api_key_source", "AUTHORIZER"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_binaryMediaTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1(rName, "application/octet-stream"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1(rName, "application/octet"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_BinaryMediaTypes_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1OverrideBody(rName, "application/octet-stream", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1OverrideBody(rName, "application/octet", "image/jpeg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1OverrideBody(rName, "application/octet", "image/png"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_BinaryMediaTypes_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_binaryMediaTypes1SetByBody(rName, "application/octet-stream"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "binary_media_types.0", "application/octet-stream"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_body(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			// The body is expected to only set a title (name) and one route
			{
				Config: testAccRestAPIConfig_body(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/test"}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_body(rName, "/update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/update"}),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, "execution_arn"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_description(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Description_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_descriptionOverrideBody(rName, "tfdescription1", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tfdescription1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated description still overrides
			{
				Config: testAccRestAPIConfig_descriptionOverrideBody(rName, "tfdescription2", "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tfdescription2"),
				),
			},
			// Verify updated body description is still overridden
			{
				Config: testAccRestAPIConfig_descriptionOverrideBody(rName, "tfdescription2", "oasdescription2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tfdescription2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Description_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_descriptionSetByBody(rName, "oasdescription1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "oasdescription1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_disableExecuteAPIEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_disableExecuteEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_disableExecuteEndpoint(rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
				),
			},
			{
				Config: testAccRestAPIConfig_disableExecuteEndpoint(rName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_DisableExecuteAPIEndpoint_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_disableExecuteEndpointOverrideBody(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify override can be unset (only for body set to false)
			{
				Config: testAccRestAPIConfig_disableExecuteEndpointOverrideBody(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtFalse),
				),
			},
			// Verify override can be reset
			{
				Config: testAccRestAPIConfig_disableExecuteEndpointOverrideBody(rName, true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_DisableExecuteAPIEndpoint_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_disableExecuteEndpointSetByBody(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "disable_execute_api_endpoint", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Endpoint_vpcEndpointIDs(t *testing.T) {
	ctx := acctest.Context(t)
	var restApi apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_vpcEndpointIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIds2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName2, names.AttrID),
				),
			},
			{
				Config: testAccRestAPIConfig_vpcEndpointIDs1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &restApi),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.types.0", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test.0"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test.1"
	vpcEndpointResourceName3 := "aws_vpc_endpoint.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsOverrideBody(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated configuration value still overrides
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsOverrideBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			// Verify updated body value is still overridden
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsOverrideBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_mergeBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test.0"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test.1"
	vpcEndpointResourceName3 := "aws_vpc_endpoint.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsMergeBody(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},

			// Verify updated endpoint configuration, and endpoint from OAS is discarded.
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsMergeBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			// Verify updated endpoint configuration, and endpoint from OAS is discarded.
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsMergeBody(rName, vpcEndpointResourceName3, vpcEndpointResourceName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName3, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_overrideToMergeBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName1 := "aws_vpc_endpoint.test.0"
	vpcEndpointResourceName2 := "aws_vpc_endpoint.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsOverrideBody(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},

			// Add the new attribute and verify works as desired.
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsMergeBody(rName, vpcEndpointResourceName1, vpcEndpointResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName1, names.AttrID),
					testAccCheckRestAPIEndpointsCount(ctx, &conf, 1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_EndpointVPCEndpointIDs_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsSetByBody(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "endpoint_configuration.0.vpc_endpoint_ids.*", vpcEndpointResourceName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_minimumCompressionSize(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_minimumCompressionSize(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_minimumCompressionSize(rName, "-1"), // -1 removes existing values
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", ""),
				),
			},
			{
				Config: testAccRestAPIConfig_minimumCompressionSize(rName, "5242880"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "5242880"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_MinimumCompressionSize_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_minimumCompressionSizeOverrideBody(rName, acctest.Ct1, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated minimum compression size still overrides
			{
				Config: testAccRestAPIConfig_minimumCompressionSizeOverrideBody(rName, acctest.Ct2, 5242880),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", acctest.Ct2),
				),
			},
			// Verify updated body minimum compression size is still overridden
			{
				Config: testAccRestAPIConfig_minimumCompressionSizeOverrideBody(rName, acctest.Ct2, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_MinimumCompressionSize_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_minimumCompressionSizeSetByBody(rName, 1048576),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "minimum_compression_size", "1048576"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Name_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_nameOverrideBody(rName, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify updated name still overrides
			{
				Config: testAccRestAPIConfig_nameOverrideBody(rName2, "title1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
			// Verify updated title still overrides
			{
				Config: testAccRestAPIConfig_nameOverrideBody(rName2, "title2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_FailOnWarnings(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			// Verify invalid body fails creation, when fail_on_warnings is true
			{
				Config:      testAccRestAPIConfig_failOnWarnings(rName, "original", "fail_on_warnings = true"),
				ExpectError: regexache.MustCompile(`BadRequestException: Warnings found during import`),
			},
			// Verify invalid body succeeds creation, when fail_on_warnings is not set
			{
				Config: testAccRestAPIConfig_failOnWarnings(rName, "original", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/users"}),
					resource.TestMatchResourceAttr(resourceName, names.AttrDescription, regexache.MustCompile(`original`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
			// Verify invalid body fails update, when fail_on_warnings is true
			{
				Config:      testAccRestAPIConfig_failOnWarnings(rName, "update", "fail_on_warnings = true"),
				ExpectError: regexache.MustCompile(`BadRequestException: Warnings found during import`),
			},
			// Verify invalid body succeeds update, when fail_on_warnings is not set
			{
				Config: testAccRestAPIConfig_failOnWarnings(rName, "update", ""),
				Check:  resource.TestMatchResourceAttr(resourceName, names.AttrDescription, regexache.MustCompile(`update`)),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_parameters1(rName, "basepath", "prepend"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/foo", "/foo/bar", "/foo/bar/baz", "/foo/bar/baz/test"}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", names.AttrParameters, "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_parameters1(rName, "basepath", "ignore"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/test"}),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_rest_api.test"
	expectedPolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Condition":{"IpAddress":{"aws:SourceIp":"123.123.123.123/32"}},"Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	expectedUpdatePolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Effect":"Deny","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, expectedPolicyText),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy, "put_rest_api_mode"},
			},
			{
				Config: testAccRestAPIConfig_updatePolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, expectedUpdatePolicyText),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_order(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_api_gateway_rest_api.test"
	expectedPolicyText := `{"Statement":[{"Action":"execute-api:Invoke","Condition":{"IpAddress":{"aws:SourceIp":["123.123.123.123/32","122.122.122.122/32","169.254.169.253/32"]}},"Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*"}],"Version":"2012-10-17"}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_policyOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, expectedPolicyText),
				),
			},
			{
				Config:   testAccRestAPIConfig_policyNewOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_overrideBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_policyOverrideBody(rName, "/test", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/test"}),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Allow"`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", names.AttrPolicy, "put_rest_api_mode"},
			},
			// Verify updated body still has override policy
			{
				Config: testAccRestAPIConfig_policyOverrideBody(rName, "/test2", "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Allow"`)),
				),
			},
			// Verify updated policy still overrides body
			{
				Config: testAccRestAPIConfig_policyOverrideBody(rName, "/test2", "Deny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					testAccCheckRestAPIRoutes(ctx, &conf, []string{"/", "/test2"}),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Deny"`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPI_Policy_setByBody(t *testing.T) {
	ctx := acctest.Context(t)
	var conf apigateway.GetRestApiOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_rest_api.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRESTAPIDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIConfig_policySetByBody(rName, "Allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRESTAPIExists(ctx, resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Allow"`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"body", "put_rest_api_mode"},
			},
		},
	})
}

func testAccCheckRestAPIRoutes(ctx context.Context, conf *apigateway.GetRestApiOutput, routes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		resp, err := conn.GetResources(ctx, &apigateway.GetResourcesInput{
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

func testAccCheckRestAPIEndpointsCount(ctx context.Context, conf *apigateway.GetRestApiOutput, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		resp, err := conn.GetRestApi(ctx, &apigateway.GetRestApiInput{
			RestApiId: conf.Id,
		})
		if err != nil {
			return err
		}

		actualEndpoints := map[string]bool{}
		for _, endpoint := range resp.EndpointConfiguration.VpcEndpointIds {
			actualEndpoints[endpoint] = true
		}

		if len(resp.EndpointConfiguration.VpcEndpointIds) != count {
			return fmt.Errorf("Found unexpected endpoint in endpoints %v", actualEndpoints)
		}

		return nil
	}
}

func testAccCheckRESTAPIExists(ctx context.Context, n string, v *apigateway.GetRestApiOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindRestAPIByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRESTAPIDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_rest_api" {
				continue
			}

			_, err := tfapigateway.FindRestAPIByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway REST API %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRestAPIConfig_endpointConfiguration(rName, endpointType string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "%s"

  endpoint_configuration {
    types = ["%s"]
  }
}
`, rName, endpointType)
}

func testAccRestAPIConfig_disableExecuteEndpoint(rName string, disableExecuteApiEndpoint bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  disable_execute_api_endpoint = %[2]t
  name                         = %[1]q
}
`, rName, disableExecuteApiEndpoint)
}

func testAccRestAPIConfig_disableExecuteEndpointOverrideBody(rName string, configDisableExecuteApiEndpoint bool, bodyDisableExecuteApiEndpoint bool) string {
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

func testAccRestAPIConfig_disableExecuteEndpointSetByBody(rName string, bodyDisableExecuteApiEndpoint bool) string {
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

func testAccRestAPIConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRestAPIConfig_vpcEndpointIDs1(rName string) string {
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

func testAccRestAPIConfig_endpointConfigurationVPCEndpointIds2(rName string) string {
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

func testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsOverrideBody(rName string, configVpcEndpointResourceName string, bodyVpcEndpointResourceName string) string {
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

func testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsMergeBody(rName string, configVpcEndpointResourceName string, bodyVpcEndpointResourceName string) string {
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
  name              = %[1]q
  put_rest_api_mode = "merge"

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

func testAccRestAPIConfig_endpointConfigurationVPCEndpointIdsSetByBody(rName string) string {
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

func testAccRestAPIConfig_policy(rName string) string {
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

func testAccRestAPIConfig_updatePolicy(rName string) string {
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

func testAccRestAPIConfig_policyOrder(rName string) string {
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

func testAccRestAPIConfig_policyNewOrder(rName string) string {
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

func testAccRestAPIConfig_keySource(rName string, apiKeySource string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  api_key_source = %[2]q
  name           = %[1]q
}
`, rName, apiKeySource)
}

func testAccRestAPIConfig_keySourceOverrideBody(rName string, apiKeySource string, bodyApiKeySource string) string {
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

func testAccRestAPIConfig_keySourceSetByBody(rName string, bodyApiKeySource string) string {
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

func testAccRestAPIConfig_binaryMediaTypes1(rName string, binaryMediaTypes1 string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  binary_media_types = [%[2]q]
  name               = %[1]q
}
`, rName, binaryMediaTypes1)
}

func testAccRestAPIConfig_binaryMediaTypes1OverrideBody(rName string, binaryMediaTypes1 string, bodyBinaryMediaTypes1 string) string {
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

func testAccRestAPIConfig_binaryMediaTypes1SetByBody(rName string, bodyBinaryMediaTypes1 string) string {
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

func testAccRestAPIConfig_body(rName string, basePath string) string {
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

func testAccRestAPIConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccRestAPIConfig_descriptionOverrideBody(rName string, description string, bodyDescription string) string {
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

func testAccRestAPIConfig_descriptionSetByBody(rName string, bodyDescription string) string {
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

func testAccRestAPIConfig_minimumCompressionSize(rName string, minimumCompressionSize string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = %[1]q
  minimum_compression_size = %[2]q
}
`, rName, minimumCompressionSize)
}

func testAccRestAPIConfig_minimumCompressionSizeOverrideBody(rName string, minimumCompressionSize string, bodyMinimumCompressionSize int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name                     = %[1]q
  minimum_compression_size = %[2]q

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

func testAccRestAPIConfig_minimumCompressionSizeSetByBody(rName string, bodyMinimumCompressionSize int) string {
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

func testAccRestAPIConfig_nameOverrideBody(rName string, title string) string {
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

func testAccRestAPIConfig_parameters1(rName string, parameterKey1 string, parameterValue1 string) string {
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

func testAccRestAPIConfig_policyOverrideBody(rName string, bodyPath string, policyEffect string) string {
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

func testAccRestAPIConfig_policySetByBody(rName string, bodyPolicyEffect string) string {
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

func testAccRestAPIConfig_failOnWarnings(rName string, title string, failOnWarnings string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name        = %[1]q
  description = %[2]q
  %[3]s
  body = jsonencode({
    openapi = "3.0.0"
    info = {
      title       = "Sample API"
      description = %[2]q
      version     = "0.1.9"
    }
    servers = [{
      url = "http://api.example.com/v1"
    }]
    components = {
      invalid_key_will_warn = "a_value"
    }
    paths = {
      "/users" = {
        get = {
          summary     = "Returns a list of users."
          description = "Optional extended description in CommonMark or HTML."
          responses = {
            "200" = {
              description = "A JSON array of user names"
              content = {
                "application/json" = {
                  schema = {
                    type = "array"
                    items = {
                      type = "string"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  })
}
`, rName, title, failOnWarnings)
}
