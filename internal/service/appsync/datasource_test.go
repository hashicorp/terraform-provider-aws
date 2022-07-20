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

func testAccDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appsync", regexp.MustCompile(fmt.Sprintf("apis/.+/datasources/%s", rName))),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccDataSourceConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_dynamoDBRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.region", acctest.Region()),
				),
			},
			{
				Config: testAccDataSourceConfig_typeDynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "1"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.use_caller_credentials", "true"),
				),
			},
			{
				Config: testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.use_caller_credentials", "false"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-Elasticsearch tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_elasticSearchRegion(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.0.region", acctest.Region()),
				),
			},
			{
				Config: testAccDataSourceConfig_typeElasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", "1"),
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

func testAccDataSource_HTTP_endpoint(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_httpEndpoint(rName, "http://example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "type", "HTTP"),
				),
			},
			{
				Config: testAccDataSourceConfig_httpEndpoint(rName, "http://example.org"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.org"),
					resource.TestCheckResourceAttr(resourceName, "type", "HTTP"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
				),
			},
			{
				Config: testAccDataSourceConfig_typeHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "HTTP"),
				),
			},
		},
	})
}

func testAccDataSource_Type_dynamoDB(t *testing.T) {
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	dynamodbTableResourceName := "aws_dynamodb_table.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeDynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dynamodb_config.0.table_name", dynamodbTableResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "dynamodb_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "AMAZON_DYNAMODB"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_appsync_datasource.test"

	// Keep this test Parallel as it takes considerably longer to run than any non-Elasticsearch tests.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeElasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "elasticsearch_config.0.endpoint"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "AMAZON_ELASTICSEARCH"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", "http://example.com"),
					resource.TestCheckResourceAttr(resourceName, "type", "HTTP"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeHTTPAuth(rName, acctest.Region()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "http_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.endpoint", fmt.Sprintf("https://appsync.%s.amazonaws.com/", acctest.Region())),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.authorization_type", "AWS_IAM"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.0.signing_region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "http_config.0.authorization_config.0.aws_iam_config.0.signing_service_name", "appsync"),
					resource.TestCheckResourceAttr(resourceName, "type", "HTTP"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeRelationalDatabase(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.db_cluster_identifier", "aws_rds_cluster.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.aws_secret_store_arn", "aws_secretsmanager_secret.test", "arn"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeRelationalDatabaseOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.schema", "mydb"),
					resource.TestCheckResourceAttr(resourceName, "relational_database_config.0.http_endpoint_config.0.region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.db_cluster_identifier", "aws_rds_cluster.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.database_name", "aws_rds_cluster.test", "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "relational_database_config.0.http_endpoint_config.0.aws_secret_store_arn", "aws_secretsmanager_secret.test", "arn"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	iamRoleResourceName := "aws_iam_role.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "lambda_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_config.0.function_arn", lambdaFunctionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "AWS_LAMBDA"),
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
	rName := fmt.Sprintf("tfacctest%d", sdkacctest.RandInt())
	resourceName := "aws_appsync_datasource.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(appsync.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, appsync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroyDataSource,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConfig_typeNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExistsDataSource(resourceName),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
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

func testAccCheckDestroyDataSource(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_datasource" {
			continue
		}

		apiID, name, err := tfappsync.DecodeID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appsync.GetDataSourceInput{
			ApiId: aws.String(apiID),
			Name:  aws.String(name),
		}

		_, err = conn.GetDataSource(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckExistsDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource has no ID: %s", name)
		}

		apiID, name, err := tfappsync.DecodeID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncConn

		input := &appsync.GetDataSourceInput{
			ApiId: aws.String(apiID),
			Name:  aws.String(name),
		}

		_, err = conn.GetDataSource(input)

		return err
	}
}

func testAccDatasourceConfig_dynamoDBBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "UserId"
  name           = %q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "UserId"
    type = "S"
  }
}

resource "aws_iam_role" "test" {
  name = %q

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
`, rName, rName)
}

func testAccDataSourceConfig_Base_elasticsearch(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %q

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role" "test" {
  name = %q

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
`, rName, rName)
}

func testAccDatasourceConfig_lambdaBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = "%slambda"

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
  function_name = %q
  handler       = "exports.test"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
  name = %q

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
`, rName, rName, rName)
}

func testAccDataSourceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  description = %q
  name        = %q
  type        = "HTTP"

  http_config {
    endpoint = "http://example.com"
  }
}
`, rName, description, rName)
}

func testAccDataSourceConfig_dynamoDBRegion(rName, region string) string {
	return testAccDatasourceConfig_dynamoDBBase(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    region     = %q
    table_name = aws_dynamodb_table.test.name
  }
}
`, rName, rName, region)
}

func testAccDataSourceConfig_dynamoDBUseCallerCredentials(rName string, useCallerCredentials bool) string {
	return testAccDatasourceConfig_dynamoDBBase(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "AWS_IAM"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    table_name             = aws_dynamodb_table.test.name
    use_caller_credentials = %t
  }
}
`, rName, rName, useCallerCredentials)
}

func testAccDataSourceConfig_elasticSearchRegion(rName, region string) string {
	return testAccDataSourceConfig_Base_elasticsearch(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_ELASTICSEARCH"

  elasticsearch_config {
    endpoint = "https://${aws_elasticsearch_domain.test.endpoint}"
    region   = %q
  }
}
`, rName, rName, region)
}

func testAccDataSourceConfig_httpEndpoint(rName, endpoint string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = %q
  type   = "HTTP"

  http_config {
    endpoint = %q
  }
}
`, rName, rName, endpoint)
}

func testAccDataSourceConfig_typeDynamoDB(rName string) string {
	return testAccDatasourceConfig_dynamoDBBase(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_DYNAMODB"

  dynamodb_config {
    table_name = aws_dynamodb_table.test.name
  }
}
`, rName, rName)
}

func testAccDataSourceConfig_typeElasticsearch(rName string) string {
	return testAccDataSourceConfig_Base_elasticsearch(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AMAZON_ELASTICSEARCH"

  elasticsearch_config {
    endpoint = "https://${aws_elasticsearch_domain.test.endpoint}"
  }
}
`, rName, rName)
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

func testAccDataSourceConfigBaseRelationalDatabase(rName string) string {
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
  cluster_identifier              = %[1]q
  engine_mode                     = "serverless"
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true

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
`, rName)
}

func testAccDataSourceConfig_typeRelationalDatabase(rName string) string {
	return testAccDataSourceConfigBaseRelationalDatabase(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccDataSourceConfig_typeRelationalDatabaseOptions(rName string) string {
	return testAccDataSourceConfigBaseRelationalDatabase(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccDataSourceConfig_typeLambda(rName string) string {
	return testAccDatasourceConfig_lambdaBase(rName) + fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id           = aws_appsync_graphql_api.test.id
  name             = %q
  service_role_arn = aws_iam_role.test.arn
  type             = "AWS_LAMBDA"

  lambda_config {
    function_arn = aws_lambda_function.test.arn
  }
}
`, rName, rName)
}

func testAccDataSourceConfig_typeNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %q
}

resource "aws_appsync_datasource" "test" {
  api_id = aws_appsync_graphql_api.test.id
  name   = %q
  type   = "NONE"
}
`, rName, rName)
}
