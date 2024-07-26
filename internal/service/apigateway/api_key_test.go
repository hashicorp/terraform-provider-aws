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
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayAPIKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1, apiKey2 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apikeys/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, "customer_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey2),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/apikeys/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, "customer_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrValue),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_customerID(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1, apiKey2 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_customerID(rName, "cid1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, "customer_id", "cid1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_customerID(rName, "cid2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey2),
					testAccCheckAPIKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, "customer_id", "cid2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_description(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1, apiKey2 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey2),
					testAccCheckAPIKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1, apiKey2 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAPIKeyConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey2),
					testAccCheckAPIKeyNotRecreated(&apiKey1, &apiKey2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKey_value(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_value(rName, `8bjqPK0BkA5N32bwYj4no2aw3eqsSM1o67eXFTNU`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, `8bjqPK0BkA5N32bwYj4no2aw3eqsSM1o67eXFTNU`),
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

func TestAccAPIGatewayAPIKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey1 apigateway.GetApiKeyOutput
	resourceName := "aws_api_gateway_api_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, resourceName, &apiKey1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapigateway.ResourceAPIKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAPIKeyExists(ctx context.Context, n string, v *apigateway.GetApiKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		output, err := tfapigateway.FindAPIKeyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAPIKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_api_key" {
				continue
			}

			_, err := tfapigateway.FindAPIKeyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway API Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAPIKeyNotRecreated(i, j *apigateway.GetApiKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedDate).Equal(aws.ToTime(j.CreatedDate)) {
			return fmt.Errorf("API Gateway API Key recreated")
		}

		return nil
	}
}

func testAccAPIKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAPIKeyConfig_customerID(rName, customerID string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  customer_id = %[2]q
  name        = %[1]q
}
`, rName, customerID)
}

func testAccAPIKeyConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}

func testAccAPIKeyConfig_enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  enabled = %[2]t
  name    = %[1]q
}
`, rName, enabled)
}

func testAccAPIKeyConfig_value(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name  = %[1]q
  value = %[2]q
}
`, rName, value)
}
