// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseLayerVersionARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		arn         string
		wantLayer   string
		wantVersion int64
		wantErr     bool
	}{
		{
			name:        "valid ARN",
			arn:         "arn:aws:lambda:us-east-1:123456789012:layer:test-layer:1", // lintignore:AWSAT003,AWSAT005
			wantLayer:   "arn:aws:lambda:us-east-1:123456789012:layer:test-layer",   // lintignore:AWSAT003,AWSAT005
			wantVersion: 1,
			wantErr:     false,
		},
		{
			name:    "no version",
			arn:     "arn:aws:lambda:us-east-1:123456789012:layer:test-layer", // lintignore:AWSAT003,AWSAT005
			wantErr: true,
		},
		{
			name:    "invalid format",
			arn:     "invalid-arn",
			wantErr: true,
		},
		{
			name:    "invalid version",
			arn:     "arn:aws:lambda:us-east-1:123456789012:layer:test-layer:abc", // lintignore:AWSAT003,AWSAT005
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			layer, version, err := tflambda.ParseLayerVersionARN(tt.arn)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLayerVersionARN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if layer != tt.wantLayer {
					t.Errorf("parseLayerVersionARN() layer = %v, want %v", layer, tt.wantLayer)
				}
				if version != tt.wantVersion {
					t.Errorf("parseLayerVersionARN() version = %v, want %v", version, tt.wantVersion)
				}
			}
		})
	}
}

func TestLayerNameFromARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layerArn string
		want     string
	}{
		{
			name:     "valid ARN",
			layerArn: "arn:aws:lambda:us-east-1:123456789012:layer:test-layer", // lintignore:AWSAT003,AWSAT005
			want:     "test-layer",
		},
		{
			name:     "invalid ARN",
			layerArn: "invalid-arn",
			want:     "",
		},
		{
			name:     "short ARN",
			layerArn: "arn:aws:lambda", // lintignore:AWSAT005
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tflambda.LayerNameFromARN(tt.layerArn); got != tt.want {
				t.Errorf("layerNameFromARN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccLambdaLayerVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_runtimes.%", resourceName, "compatible_runtimes.%s"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_info", resourceName, "license_info"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_arn", resourceName, "layer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "code_sha256", resourceName, "code_sha256"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_hash", resourceName, "code_sha256"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_code_size", resourceName, "source_code_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_profile_version_arn", resourceName, "signing_profile_version_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "signing_job_arn", resourceName, "signing_job_arn"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_version(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_version(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_runtime(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_runtimes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_architectures(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_architecturesX86(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionDataSourceConfig_architecturesARM(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionDataSourceConfig_architecturesX86ARM(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
			{
				Config: testAccLayerVersionDataSourceConfig_architecturesNone(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "compatible_architectures", resourceName, "compatible_architectures"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_arn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_arn", resourceName, "layer_arn"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_arnWithoutVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_layer_version.test"
	resourceName := "aws_lambda_layer_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDataSourceConfig_arnWithoutVersion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_name", resourceName, "layer_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "layer_arn", resourceName, "layer_arn"),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersionDataSource_arnCrossAccountWithoutVersionError(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccLayerVersionDataSourceConfig_arnCrossAccountWithoutVersion(),
				ExpectError: regexache.MustCompile(`unable to list layer versions.*cross-account.*version number`),
			},
		},
	})
}

func testAccLayerVersionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test.layer_name
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_version(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

resource "aws_lambda_layer_version" "test_two" {
  filename            = "test-fixtures/lambdatest_modified.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test_two.layer_name
  version    = aws_lambda_layer_version.test.version
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_runtimes(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["python3.12"]
}

resource "aws_lambda_layer_version" "test_two" {
  filename            = "test-fixtures/lambdatest_modified.zip"
  layer_name          = aws_lambda_layer_version.test.layer_name
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name         = aws_lambda_layer_version.test_two.layer_name
  compatible_runtime = "python3.12"
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_architecturesX86(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs20.x"]
  compatible_architectures = ["x86_64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "x86_64"
}

`, rName)
}

func testAccLayerVersionDataSourceConfig_architecturesARM(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs20.x"]
  compatible_architectures = ["arm64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "arm64"
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_architecturesX86ARM(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs20.x"]
  compatible_architectures = ["x86_64", "arm64"]
}

data "aws_lambda_layer_version" "test" {
  layer_name              = aws_lambda_layer_version.test.layer_name
  compatible_architecture = "arm64"
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_architecturesNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_name = aws_lambda_layer_version.test.layer_name
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_arn(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_version_arn = aws_lambda_layer_version.test.arn
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_arnWithoutVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = ["nodejs20.x"]
}

data "aws_lambda_layer_version" "test" {
  layer_version_arn = aws_lambda_layer_version.test.layer_arn
}
`, rName)
}

func testAccLayerVersionDataSourceConfig_arnCrossAccountWithoutVersion() string {
	// lintignore:AWSAT003,AWSAT005
	return `
data "aws_lambda_layer_version" "test" {
  # Datadog's public layer - we don't have ListLayerVersions permission
  layer_version_arn = "arn:aws:lambda:us-east-1:464622532012:layer:Datadog-Python312"
}
`
}
