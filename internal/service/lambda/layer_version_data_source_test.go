package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaLayerVersionDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionBasicDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_runtimes.%", resourceName, "compatible_runtimes.%s"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_info", resourceName, "license_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_arn", resourceName, "layer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_hash", resourceName, "source_code_hash"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_size", resourceName, "source_code_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_profile_version_arn", resourceName, "signing_profile_version_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_job_arn", resourceName, "signing_job_arn"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_version(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionVersionDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_runtime(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionRuntimesDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_architectures(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionArchitecturesX86DataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionArchitecturesARMDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionArchitecturesX86ARMDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionArchitecturesNoneDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
		},
	})
}

func testAccLayerVersionBasicDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs12.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test.layer_name
}
`, rName)
}

func testAccLayerVersionVersionDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_layer_version" "test_two" {
  filename            = "test-fixtures/lambdatest_modified.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs12.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test_two.layer_name
  version    = aws_lambda_layer_version.test.version
}
`, rName)
}

func testAccLayerVersionRuntimesDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["go1.x"]
}

resource "aws_lambda_layer_version" "test_two" {
  filename            = "test-fixtures/lambdatest_modified.zip"
  layer_name          = aws_lambda_layer_version.test.layer_name
  compatible_runtimes = ["nodejs12.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name         = aws_lambda_layer_version.test_two.layer_name
  compatible_runtime = "go1.x"
}
`, rName)
}

func testAccLayerVersionArchitecturesX86DataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs12.x"]
  compatible_architectures = ["x86_64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "x86_64"
}

`, rName)
}

func testAccLayerVersionArchitecturesARMDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs12.x"]
  compatible_architectures = ["arm64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "arm64"
}
`, rName)
}

func testAccLayerVersionArchitecturesX86ARMDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs12.x"]
  compatible_architectures = ["x86_64", "arm64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "arm64"
}
`, rName)
}

func testAccLayerVersionArchitecturesNoneDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs12.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test.layer_name
}
`, rName)
}
