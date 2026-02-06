// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	rName3 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appsync", regexache.MustCompile("apis/.+/functions/.+")),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "max_batch_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sync_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "data_source", "aws_appsync_datasource.test", names.AttrName),
				),
			},
			{
				Config: testAccFunctionConfig_basic(rName1, rName3, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName3),
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
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_code(rName1, rName2, "test-fixtures/test-code.js"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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
	rName := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_sync(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
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
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_description(rName1, rName2, acctest.Region(), "test description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description 1"),
				),
			},
			{
				Config: testAccFunctionConfig_description(rName1, rName2, acctest.Region(), "test description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description 2"),
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
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_responseMappingTemplate(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
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
	rName1 := fmt.Sprintf("tfacctest%d", acctest.RandInt(t))
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config awstypes.FunctionConfiguration

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappsync.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFunctionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_function" {
				continue
			}

			_, err := tfappsync.FindFunctionByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["function_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Function %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFunctionExists(ctx context.Context, t *testing.T, n string, v *awstypes.FunctionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppSyncClient(ctx)

		output, err := tfappsync.FindFunctionByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes["function_id"])

		if err != nil {
			return err
		}

		*v = *output

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
