// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appsync", regexache.MustCompile(fmt.Sprintf("apis/.+/datasources/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "event_bridge_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NONE"),
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

func testAccDataSource_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccDataSourceConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
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

func testAccDataSource_DynamoDB_region(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_dynamoDBRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.region", acctest.Region()),
				),
			},
			{
				Config: testAccDataSourceConfig_typeDynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.region", acctest.Region()),
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

func testAccDataSource_DynamoDB_useCallerCredentials(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.use_caller_credentials", acctest.CtTrue),
				),
			},
			{
				Config: testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.use_caller_credentials", acctest.CtFalse),
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

func TestAccAppSyncDataSource_Elasticsearch_region(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-Elasticsearch tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_elasticSearchRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.0.region", acctest.Region()),
				),
			},
			{
				Config: testAccDataSourceConfig_typeElasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.0.region", acctest.Region()),
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

func TestAccAppSyncDataSource_OpenSearchService_region(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-OpenSearchService tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_openSearchServiceRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.0.region", acctest.Region()),
				),
			},
			{
				Config: testAccDataSourceConfig_typeOpenSearchService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.0.region", acctest.Region()),
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

func testAccDataSource_HTTP_endpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_httpEndpoint(rName, "http://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
				),
			},
			{
				Config: testAccDataSourceConfig_httpEndpoint(rName, "http://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.org"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
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

func testAccDataSource_type(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NONE"),
				),
			},
			{
				Config: testAccDataSourceConfig_typeHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
				),
			},
		},
	})
}

func testAccDataSource_Type_dynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	dynamodbTableResourceName := "aws_dynamodb_table.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeDynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "dynamodb_config.0.table_name", dynamodbTableResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AMAZON_DYNAMODB"),
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

func TestAccAppSyncDataSource_Type_elasticSearch(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-Elasticsearch tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeElasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "elasticsearch_config.0.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AMAZON_ELASTICSEARCH"),
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

func TestAccAppSyncDataSource_Type_openSearchService(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-OpenSearchService tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeOpenSearchService(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "opensearchservice_config.0.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "opensearchservice_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AMAZON_OPENSEARCH_SERVICE"),
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

func testAccDataSource_Type_http(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
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

func testAccDataSource_Type_httpAuth(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeHTTPAuth(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", fmt.Sprintf("https://appsync.%s.amazonaws.com/", acctest.Region())),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.authorization_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.0.signing_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.0.signing_service_name", "appsync"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HTTP"),
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

func testAccDataSource_Type_relationalDatabase(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeRelationalDatabase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.db_cluster_identifier", "aws_rds_cluster.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.aws_secret_store_arn", "aws_secretsmanager_secret.test", names.AttrARN),
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

func testAccDataSource_Type_relationalDatabaseWithOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeRelationalDatabaseOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.schema", "mydb"),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.db_cluster_identifier", "aws_rds_cluster.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.database_name", "aws_rds_cluster.test", names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.aws_secret_store_arn", "aws_secretsmanager_secret.test", names.AttrARN),
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

func testAccDataSource_Type_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.function_arn", lambdaFunctionResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AWS_LAMBDA"),
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

func testAccDataSource_Type_eventBridge(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	eventBusResourceName := "aws_cloudwatch_event_bus.test"
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeEventBridge(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_bridge_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "event_bridge_config.0.event_bus_arn", eventBusResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AMAZON_EVENTBRIDGE"),
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

func testAccDataSource_Type_none(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataSourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NONE"),
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

func testAccCheckDataSourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_datasource" {
				continue
			}

			_, err := tfappsync.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Data Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckExistsDataSource(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		_, err := tfappsync.FindDataSourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_id"], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccDatasourceConfig_baseDynamoDB(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "UserId"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "UserId"
    type = "S"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_dynamodb_table.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDataSourceConfig_baseElasticsearch(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "es:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_elasticsearch_domain.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDataSourceConfig_baseOpenSearchService(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearch_domain" "test" {
  domain_name = %[1]q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "es:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_opensearch_domain.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDatasourceConfig_baseLambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.test"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs16.x"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_lambda_function.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDatasourceConfig_baseEventBridge(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "events:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_cloudwatch_event_bus.test.arn}"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccDataSourceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  description = %[2]q
  name        = %[1]q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}
`, rName, description)
}

func testAccDataSourceConfig_dynamoDBRegion(rName, region string) string {
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
  }
}
`, rName, region))
}

func testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName string, useCallerCredentials bool) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_baseDynamoDB(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_IAM"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    table_name             = aws_dynamodb_table.test.name
    use_caller_credentials = %[2]t
  }
}
`, rName, useCallerCredentials))
}

func testAccDataSourceConfig_elasticSearchRegion(rName, region string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseElasticsearch(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_ELASTICSEARCH"

  elasticsearch_config {
    endpoint = "https://${aws_elasticsearch_domain.test.endpoint}"
    region   = %[2]q
  }
}
`, rName, region))
}

func testAccDataSourceConfig_openSearchServiceRegion(rName, region string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseOpenSearchService(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_OPENSEARCH_SERVICE"

  opensearchservice_config {
    endpoint = "https://${aws_opensearch_domain.test.endpoint}"
    region   = %[2]q
  }
}
`, rName, region))
}

func testAccDataSourceConfig_httpEndpoint(rName, endpoint string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = %[1]q
  type   = "HTTP"

  http_config {
    endpoint = %[2]q
  }
}
`, rName, endpoint)
}

func testAccDataSourceConfig_typeDynamoDB(rName string) string {
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
    table_name = aws_dynamodb_table.test.name
  }
}
`, rName))
}

func testAccDataSourceConfig_typeElasticsearch(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseElasticsearch(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_ELASTICSEARCH"

  elasticsearch_config {
    endpoint = "https://${aws_elasticsearch_domain.test.endpoint}"
  }
}
`, rName))
}

func testAccDataSourceConfig_typeOpenSearchService(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseOpenSearchService(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_OPENSEARCH_SERVICE"

  opensearchservice_config {
    endpoint = "https://${aws_opensearch_domain.test.endpoint}"
  }
}
`, rName))
}

func testAccDataSourceConfig_typeHTTP(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = %[1]q
  type   = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}
`, rName)
}

func testAccDataSourceConfig_typeHTTPAuth(rName, region string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "appsync.amazonaws.com"
        },
        "Effect" : "Allow"
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : [
          "appsync:ListGraphqlApis"
        ],
        "Effect" : "Allow",
        "Resource" : [
          "*"
        ]
      }
    ]
  })
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  type             = "HTTP"
  service_role_arn = aws_iam_role.test.arn

  http_config {
    endpoint = "https://appsync.%[2]s.amazonaws.com/"

    authorization_config {
      authorization_type = "AWS_IAM"

      aws_iam_config {
        signing_region       = %[2]q
        signing_service_name = "appsync"
      }
    }
  }
}
`, rName, region)
}

func testAccDataSourceConfig_baseRelationalDatabase(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode(
    {
      username = "foo"
      password = "mustbeeightcharaters"
    }
  )
}

resource "aws_rds_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = %[2]q
  engine_mode         = "serverless"
  database_name       = "mydb"
  master_username     = "foo"
  master_password     = "mustbeeightcharaters"
  skip_final_snapshot = true

  scaling_configuration {
    min_capacity = 1
    max_capacity = 2
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "appsync.amazonaws.com"
        },
        "Effect" : "Allow"
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : [
          "rds:*"
        ],
        "Effect" : "Allow",
        "Resource" : [
          aws_rds_cluster.test.arn
        ]
      }
    ]
  })
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}
`, rName, tfrds.ClusterEngineAuroraMySQL)
}

func testAccDataSourceConfig_typeRelationalDatabase(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseRelationalDatabase(rName), fmt.Sprintf(`
resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "RELATIONAL_DATABASE"

  relational_database_config {
    http_endpoint_config {
      db_cluster_identifier = aws_rds_cluster.test.id
      aws_secret_store_arn  = aws_secretsmanager_secret.test.arn
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_typeRelationalDatabaseOptions(rName string) string {
	return acctest.ConfigCompose(testAccDataSourceConfig_baseRelationalDatabase(rName), fmt.Sprintf(`
resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "RELATIONAL_DATABASE"

  relational_database_config {
    http_endpoint_config {
      db_cluster_identifier = aws_rds_cluster.test.id
      database_name         = aws_rds_cluster.test.database_name
      aws_secret_store_arn  = aws_secretsmanager_secret.test.arn
      schema                = "mydb"
    }
  }
}
`, rName))
}

func testAccDataSourceConfig_typeLambda(rName string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_baseLambda(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AWS_LAMBDA"

  lambda_config {
    function_arn = aws_lambda_function.test.arn
  }
}
`, rName))
}

func testAccDataSourceConfig_typeEventBridge(rName string) string {
	return acctest.ConfigCompose(testAccDatasourceConfig_baseEventBridge(rName), fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %[1]q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_EVENTBRIDGE"

  event_bridge_config {
    event_bus_arn = aws_cloudwatch_event_bus.test.arn
  }
}
`, rName))
}

func testAccDataSourceConfig_typeNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[1]q
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = %[1]q
  type   = "NONE"
}
`, rName)
}
