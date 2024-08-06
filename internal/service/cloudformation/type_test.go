// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", sdkacctest.RandString(8))
	zipPath := testAccTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_name(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudformation", regexache.MustCompile(fmt.Sprintf("type/resource/%s/.+", strings.ReplaceAll(typeName, "::", "-")))),
					resource.TestCheckResourceAttr(resourceName, "default_version_id", ""),
					resource.TestCheckResourceAttr(resourceName, "deprecated_status", string(awstypes.DeprecatedStatusLive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "An example resource schema demonstrating some basic constructs and validation rules."),
					resource.TestCheckResourceAttr(resourceName, names.AttrExecutionRoleARN, ""),
					resource.TestCheckResourceAttr(resourceName, "is_default_version", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "provisioning_type", string(awstypes.ProvisioningTypeFullyMutable)),
					resource.TestCheckResourceAttr(resourceName, "schema_handler_package", fmt.Sprintf("s3://%s/test", rName)),
					resource.TestMatchResourceAttr(resourceName, names.AttrSchema, regexache.MustCompile(`^\{.*`)),
					resource.TestCheckResourceAttr(resourceName, "source_url", "https://github.com/aws-cloudformation/aws-cloudformation-rpdk.git"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RegistryTypeResource)),
					acctest.CheckResourceAttrRegionalARN(resourceName, "type_arn", "cloudformation", fmt.Sprintf("type/resource/%s", strings.ReplaceAll(typeName, "::", "-"))),
					resource.TestCheckResourceAttr(resourceName, "type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "visibility", string(awstypes.VisibilityPrivate)),
					resource.TestMatchResourceAttr(resourceName, "version_id", regexache.MustCompile(`.+`)),
				),
			},
		},
	})
}

func TestAccCloudFormationType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", sdkacctest.RandString(8))
	zipPath := testAccTypeZipGenerator(t, typeName)
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_name(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceType(), resourceName),
					// Verify Delete error handling
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationType_executionRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", sdkacctest.RandString(8))
	zipPath := testAccTypeZipGenerator(t, typeName)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_executionRoleARN(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrExecutionRoleARN, iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCloudFormationType_logging(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	typeName := fmt.Sprintf("HashiCorp::TerraformAwsProvider::TfAccTest%s", sdkacctest.RandString(8))
	zipPath := testAccTypeZipGenerator(t, typeName)
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTypeConfig_logging(rName, zipPath, typeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTypeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.0.log_group_name", cloudwatchLogGroupResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "logging_config.0.log_role_arn", iamRoleResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckTypeExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		_, err := tfcloudformation.FindTypeByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set" {
				continue
			}

			_, err := tfcloudformation.FindTypeByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation Type %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// Since the CloudFormation resource schema type name must match the RegisterType TypeName
// and we randomize the latter to avoid testing race conditions, this function sets up a
// temporary directory and returns a string to the generated zip file.
//
// This is essentially the equivalent of running `cfn submit --dry-run` with the correct
// type names injected. Only .rpdk-config, handler.zip, and schema.json are required from
// the following zip structure that is created:
// .
// ├── .cfn_metadata.json
// ├── .rpdk-config
// ├── Makefile
// ├── cmd
// │   ├── main.go
// │   └── resource
// │       ├── model.go
// │       └── resource.go
// ├── handler.zip
// └── schema.json
//
// Where handler.zip has the single Go binary file from bin, handler, at the root. This
// binary file is large (near 10MB as of this writing), so it is in the .gitignore file.
// The test skip messaging includes directions for how to locally build the binary.
func testAccTypeZipGenerator(t *testing.T, typeName string) string {
	t.Helper()
	tempDir := t.TempDir()
	_, currentFilePath, _, _ := runtime.Caller(0)

	sourceDirectoryPath := filepath.Join(filepath.Dir(currentFilePath), "test-fixtures", "examplecompany-exampleservice-exampleresource")
	targetDirectoryPath := filepath.Join(tempDir, "examplecompany-exampleservice-exampleresource")
	targetZipFilePath := filepath.Join(tempDir, "examplecompany-exampleservice-exampleresource.zip")

	err := os.Mkdir(targetDirectoryPath, 0777)

	if err != nil {
		t.Fatalf("error making directory %s: %s", targetDirectoryPath, err)
	}

	sourceRpdkConfigFilePath := filepath.Join(sourceDirectoryPath, ".rpdk-config")
	targetRpdkConfigFilePath := filepath.Join(targetDirectoryPath, ".rpdk-config")

	err = testAccTypeCopyFileWithTypeNameReplacement(sourceRpdkConfigFilePath, targetRpdkConfigFilePath, typeName)

	if err != nil {
		t.Fatal(err)
	}

	sourceBinDirectoryPath := filepath.Join(sourceDirectoryPath, "bin")
	sourceBinHandlerFilePath := filepath.Join(sourceBinDirectoryPath, "handler")
	targetHandlerZipFilePath := filepath.Join(targetDirectoryPath, "handler.zip")

	err = testAccTypeBuildHandlerZip(sourceBinHandlerFilePath, targetHandlerZipFilePath)

	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("%s does not exist. To generate locally: cd %s && make build", sourceBinHandlerFilePath, sourceDirectoryPath)
	}

	if err != nil {
		t.Fatal(err)
	}

	sourceSchemaFilePath := filepath.Join(sourceDirectoryPath, "examplecompany-exampleservice-exampleresource.json")
	targetSchemaFilePath := filepath.Join(targetDirectoryPath, "schema.json")

	err = testAccTypeCopyFileWithTypeNameReplacement(sourceSchemaFilePath, targetSchemaFilePath, typeName)

	if err != nil {
		t.Fatal(err)
	}

	err = testAccTypeBuildDirectoryZip(targetDirectoryPath, targetZipFilePath)

	if err != nil {
		t.Fatal(err)
	}

	return targetZipFilePath
}

func testAccTypeBuildDirectoryZip(sourceDirectoryPath string, targetZipFilePath string) error {
	targetZipFile, err := os.Create(targetZipFilePath)

	if err != nil {
		return fmt.Errorf("error creating file %s: %w", targetZipFilePath, err)
	}

	targetZipFileWriter := zip.NewWriter(targetZipFile)

	targetDirectoryFiles, err := os.ReadDir(sourceDirectoryPath)

	if err != nil {
		return fmt.Errorf("error listing %s directory files: %w", sourceDirectoryPath, err)
	}

	for _, targetDirectoryFile := range targetDirectoryFiles {
		sourceFilePath := filepath.Join(sourceDirectoryPath, targetDirectoryFile.Name())

		sourceFile, err := os.Open(sourceFilePath)

		if err != nil {
			return fmt.Errorf("error reading file %s: %w", sourceFilePath, err)
		}

		targetFile, err := targetZipFileWriter.Create(sourceFilePath)

		if err != nil {
			return fmt.Errorf("error creating file %s in zip: %w", sourceFilePath, err)
		}

		_, err = io.Copy(targetFile, sourceFile)

		if err != nil {
			return fmt.Errorf("error copying file %s into zip: %w", sourceFilePath, err)
		}
	}

	err = targetZipFileWriter.Close()

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetZipFilePath, err)
	}

	return nil
}

func testAccTypeBuildHandlerZip(sourceBinHandlerFilePath string, targetZipFilePath string) error {
	sourceBinHandlerFile, err := os.Open(sourceBinHandlerFilePath)

	if err != nil {
		return fmt.Errorf("error reading file %s: %w", sourceBinHandlerFilePath, err)
	}

	targetZipFile, err := os.Create(targetZipFilePath)

	if err != nil {
		return fmt.Errorf("error creating file %s: %w", targetZipFilePath, err)
	}

	targetZipWriter := zip.NewWriter(targetZipFile)

	targetHandlerFile, err := targetZipWriter.Create("handler")

	if err != nil {
		return fmt.Errorf("error creating file %s in zip: %w", "handler", err)
	}

	_, err = io.Copy(targetHandlerFile, sourceBinHandlerFile)

	if err != nil {
		return fmt.Errorf("error copying file %s into zip: %w", sourceBinHandlerFilePath, err)
	}

	err = targetZipWriter.Close()

	if err != nil {
		return fmt.Errorf("error writing zip %s: %w", targetZipFilePath, err)
	}

	err = targetZipFile.Close()

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetZipFilePath, err)
	}

	return nil
}

func testAccTypeCopyFileWithTypeNameReplacement(sourceFilePath string, targetFilePath string, targetTypeName string) error {
	sourceTypeName := "ExampleCompany::ExampleService::ExampleResource"
	sourceFile, err := os.ReadFile(sourceFilePath)

	if err != nil {
		return fmt.Errorf("error reading file %s: %w", sourceFilePath, err)
	}

	err = os.WriteFile(targetFilePath, []byte(strings.ReplaceAll(string(sourceFile), sourceTypeName, targetTypeName)), 0644)

	if err != nil {
		return fmt.Errorf("error writing file %s: %w", targetFilePath, err)
	}

	return nil
}

func testAccTypeConfig_base(rName string, zipPath string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test"
  source = %[2]q
}
`, rName, zipPath)
}

func testAccTypeConfig_executionRoleARN(rName string, zipPath string, typeName string) string {
	return acctest.ConfigCompose(
		testAccTypeConfig_base(rName, zipPath),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "cloudformation.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudformation_type" "test" {
  execution_role_arn     = aws_iam_role.test.arn
  schema_handler_package = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[2]q
}
`, rName, typeName))
}

func testAccTypeConfig_logging(rName string, zipPath string, typeName string) string {
	return acctest.ConfigCompose(
		testAccTypeConfig_base(rName, zipPath),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "cloudformation.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudformation_type" "test" {
  schema_handler_package = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[2]q

  logging_config {
    log_group_name = aws_cloudwatch_log_group.test.name
    log_role_arn   = aws_iam_role.test.arn
  }
}
`, rName, typeName))
}

func testAccTypeConfig_name(rName string, zipPath string, typeName string) string {
	return acctest.ConfigCompose(
		testAccTypeConfig_base(rName, zipPath),
		fmt.Sprintf(`
resource "aws_cloudformation_type" "test" {
  schema_handler_package = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  type                   = "RESOURCE"
  type_name              = %[1]q
}
`, typeName))
}
