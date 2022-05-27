package appsync_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
)

func testAccFunction_basic(t *testing.T) {
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	rName3 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile("apis/.+/functions/.+")),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "max_batch_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "sync_config.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "data_source", "aws_appsync_datasource.test", "name"),
				),
			},
			{
				Config: testAccFunctionConfig(rName1, rName3, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
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

func testAccFunction_syncConfig(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionSyncConfig(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
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
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionDescriptionConfig(rName1, rName2, acctest.Region(), "test description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "description", "test description 1"),
				),
			},
			{
				Config: testAccFunctionDescriptionConfig(rName1, rName2, acctest.Region(), "test description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
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
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionResponseMappingTemplateConfig(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
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
	rName1 := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	rName2 := fmt.Sprintf("tfexample%s", sdkacctest.RandString(8))
	resourceName := "aws_appsync_function.test"
	var config appsync.FunctionConfiguration

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig(rName1, rName2, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, &config),
					acctest.CheckResourceDisappears(acctest.Provider, tfappsync.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFunctionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
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

		_, err = conn.GetFunction(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckFunctionExists(name string, config *appsync.FunctionConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		apiID, functionID, err := tfappsync.DecodeFunctionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &appsync.GetFunctionInput{
			ApiId:      aws.String(apiID),
			FunctionId: aws.String(functionID),
		}

		output, err := conn.GetFunction(input)

		if err != nil {
			return err
		}

		*config = *output.FunctionConfiguration

		return nil
	}
}

func testAccFunctionConfig(r1, r2, region string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
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
`, testAccDataSourceConfig_DynamoDBConfig_region(r1, region), r2)
}

func testAccFunctionSyncConfig(rName, region string) string {
	return testAccDatasourceConfig_dynamoDBBase(rName) + fmt.Sprintf(`
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
`, rName, region)
}

func testAccFunctionDescriptionConfig(r1, r2, region, description string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
  description              = "%[3]s"
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
`, testAccDataSourceConfig_DynamoDBConfig_region(r1, region), r2, description)
}

func testAccFunctionResponseMappingTemplateConfig(r1, r2, region string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_appsync_function" "test" {
  api_id                   = aws_appsync_graphql_api.test.id
  data_source              = aws_appsync_datasource.test.name
  name                     = "%[2]s"
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
`, testAccDataSourceConfig_DynamoDBConfig_region(r1, region), r2)
}
