package lambda_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaFunctionDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionBasicDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "code_signing_config_arn", resourceName, "code_signing_config_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dead_letter_config.#", resourceName, "dead_letter_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ephemeral_storage.#", resourceName, "ephemeral_storage.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ephemeral_storage.0.size", resourceName, "ephemeral_storage.0.size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_name", resourceName, "function_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "handler", resourceName, "handler"),
					resource.TestCheckResourceAttrPair(dataSourceName, "invoke_arn", resourceName, "invoke_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_modified", resourceName, "last_modified"),
					resource.TestCheckResourceAttrPair(dataSourceName, "memory_size", resourceName, "memory_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", resourceName, "qualified_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reserved_concurrent_executions", resourceName, "reserved_concurrent_executions"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role", resourceName, "role"),
					resource.TestCheckResourceAttrPair(dataSourceName, "runtime", resourceName, "runtime"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_job_arn", resourceName, "signing_job_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_profile_version_arn", resourceName, "signing_profile_version_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_hash", resourceName, "source_code_hash"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_size", resourceName, "source_code_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout", resourceName, "timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracing_config.#", resourceName, "tracing_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracing_config.0.mode", resourceName, "tracing_config.0.mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_version(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionVersionDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", resourceName, "qualified_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "qualifier", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_alias(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	lambdaAliasResourceName := "aws_lambda_alias.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionAliasDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualified_arn", lambdaAliasResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualifier", lambdaAliasResourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", lambdaAliasResourceName, "function_version"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_layers(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionLayersDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "layers.#", resourceName, "layers.#"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionVPCDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.#", resourceName, "vpc_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.0.security_group_ids.#", resourceName, "vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_config.0.subnet_ids.#", resourceName, "vpc_config.0.subnet_ids.#"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_environment(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEnvironmentDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.#", resourceName, "environment.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.%", resourceName, "environment.0.variables.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.key1", resourceName, "environment.0.variables.key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "environment.0.variables.key2", resourceName, "environment.0.variables.key2"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_fileSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionFileSystemsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_config.#", resourceName, "file_system_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_config.0.arn", resourceName, "file_system_config.0.arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_config.0.local_mount_path", resourceName, "file_system_config.0.local_mount_path"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_image(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	imageLatestID := os.Getenv("AWS_LAMBDA_IMAGE_LATEST_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccImageLatestPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionImageDataSourceConfig(rName, imageLatestID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "code_signing_config_arn", resourceName, "code_signing_config_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_uri", resourceName, "image_uri"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_architectures(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionArchitecturesDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "architectures", resourceName, "architectures"),
				),
			},
		},
	})
}

func TestAccLambdaFunctionDataSource_ephemeralStorage(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function.test"
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionEphemeralStorageDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ephemeral_storage.#", resourceName, "ephemeral_storage.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ephemeral_storage.0.size", resourceName, "ephemeral_storage.0.size"),
				),
			},
		},
	})
}

func testAccImageLatestPreCheck(t *testing.T) {
	if os.Getenv("AWS_LAMBDA_IMAGE_LATEST_ID") == "" {
		t.Skip("AWS_LAMBDA_IMAGE_LATEST_ID env var must be set for Lambda Function Data Source Image Support acceptance tests.")
	}
}

func testAccFunctionBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "lambda" {
  name = %[1]q
  role = aws_iam_role.lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccFunctionBasicDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionVersionDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  publish       = true
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = 1
}
`, rName))
}

func testAccFunctionAliasDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  publish       = true
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

resource "aws_lambda_alias" "test" {
  name             = "alias-name"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_alias.test.name
}
`, rName))
}

func testAccFunctionLayersDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  layers        = [aws_lambda_layer_version.test.arn]
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionVPCDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Terraform Acceptance Testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
  }
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionEnvironmentDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"

  environment {
    variables = {
      key1 = "value1"
      key2 = "value2"
    }
  }
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionFileSystemsDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Terraform Acceptance Testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_efs_file_system" "efs_for_lambda" {
  tags = {
    Name = "efs_for_lambda"
  }
}

resource "aws_efs_mount_target" "alpha" {
  file_system_id = aws_efs_file_system.efs_for_lambda.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_efs_access_point" "access_point_1" {
  file_system_id = aws_efs_file_system.efs_for_lambda.id

  root_directory {
    path = "/lambda"

    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "777"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "lambdatest.handler"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.access_point_1.arn
    local_mount_path = "/mnt/lambda"
  }

  depends_on = [aws_efs_mount_target.alpha]
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionImageDataSourceConfig(rName, imageID string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  image_uri     = %q
  function_name = %q
  role          = aws_iam_role.lambda.arn
  package_type  = "Image"
  image_config {
    entry_point       = ["/bootstrap-with-handler"]
    command           = ["app.lambda_handler"]
    working_directory = "/var/task"
  }
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, imageID, rName))
}

func testAccFunctionArchitecturesDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}

func testAccFunctionEphemeralStorageDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccFunctionBaseDataSourceConfig(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs12.x"

  ephemeral_storage {
    size = 1024
  }
}

data "aws_lambda_function" "test" {
  function_name = aws_lambda_function.test.function_name
}
`, rName))
}
