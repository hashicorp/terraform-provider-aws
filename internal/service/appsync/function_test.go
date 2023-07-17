// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	rName3 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile("apis/.+/functions/.+")),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "max_batch_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sync_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "data_source", "aws_appsync_datasource.test", "name"),
				),
			},
			{
				Config: testAccFunctionConfig_basic(rName1, rName3, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
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

func testAccFunction_code(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_code(rName1, rName2, "test-fixtures/test-code.js"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttr(resourceName, "runtime.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime.0.name", "APPSYNC_JS"),
					resource.TestCheckResourceAttr(resourceName, "runtime.0.runtime_version", "1.0.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFunctionConfig_code(rName1, rName2, "test-fixtures/test-code-updated.js"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttrSet(resourceName, "code"),
					resource.TestCheckResourceAttr(resourceName, "runtime.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime.0.name", "APPSYNC_JS"),
					resource.TestCheckResourceAttr(resourceName, "runtime.0.runtime_version", "1.0.0"),
				),
			},
		},
	})
}

func testAccFunction_syncConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_sync(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "sync_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sync_config.0.conflict_detection", "VERSION"),
					resource.TestCheckResourceAttr(resourceName, "sync_config.0.conflict_handler", "OPTIMISTIC_CONCURRENCY"),
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

func testAccFunction_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_description(rName1, rName2, acctest.Region(), "test description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "description", "test description 1"),
				),
			},
			{
				Config: testAccFunctionConfig_description(rName1, rName2, acctest.Region(), "test description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "description", "test description 2"),
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

func testAccFunction_responseMappingTemplate(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_responseMappingTemplate(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
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

func testAccFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appsync.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appsync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFunctionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_function" {
				continue
			}

			apiID, functionID, err := tfappsync.DecodeFunctionID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &appsync.GetFunctionInput{
				ApiId:      aws.String(apiID),
				FunctionId: aws.String(functionID),
			}

			_, err = conn.GetFunctionWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
					return nil
				}
				return err
			}
		}
		return nil
	}
}

func testAccCheckFunctionExists(ctx context.Context, name string, config *appsync.FunctionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn(ctx)

		apiID, functionID, err := tfappsync.DecodeFunctionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &appsync.GetFunctionInput{
			ApiId:      aws.String(apiID),
			FunctionId: aws.String(functionID),
		}

		output, err := conn.GetFunctionWithContext(ctx, input)

		if err != nil {
			return err
		}

		*config = *output.FunctionConfiguration

		return nil
	}
}

func testAccFunctionConfig_basic(r1, r2, region string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_dynamoDBRegion(r1, region), fmt.Sprintf(`
resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = %[1]q
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, r2))
}

func testAccFunctionConfig_code(r1, r2, code string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_typeHTTP(r1), fmt.Sprintf(`
resource "aws_appsync_function" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  data_source = aws_appsync_datasource.test.name
  name        = %[1]q
  code        = file("%[2]s")

  runtime {
    name            = "APPSYNC_JS"
    runtime_version = "1.0.0"
  }
}
`, r2, code))
}

func testAccFunctionConfig_sync(rName, region string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_baseDynamoDB(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    region     = %[2]q
    table_name = aws_dynamodb_table.test.name
    versioned  = true

    delta_sync_config {
      base_table_ttl        = 60
      delta_sync_table_name = aws_dynamodb_table.test.name
      delta_sync_table_ttl  = 60
    }
  }
}

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = %[1]q
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF

  sync_config {
    conflict_detection = "VERSION"
    conflict_handler   = "OPTIMISTIC_CONCURRENCY"
  }
}
`, rName, region))
}

func testAccFunctionConfig_description(r1, r2, region, description string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_dynamoDBRegion(r1, region), fmt.Sprintf(`
resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = %[1]q
  description              = %[2]q
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, r2, description))
}

func testAccFunctionConfig_responseMappingTemplate(r1, r2, region string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_dynamoDBRegion(r1, region), fmt.Sprintf(`
resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = %[1]q
  request_mapping_template = <<EOF
{
	"version": "2018-05-29",
	"method": "GET",
	"resourcePath": "/",
	"params":{
		"headers": $utils.http.copyheaders($ctx.request.headers)
	}
}
EOF

  response_mapping_template = <<EOF
#if($ctx.result.statusCode == 200)
	$ctx.result.body
#else
	$utils.appendError($ctx.result.body, $ctx.result.statusCode)
#end
EOF
}
`, r2))
}
