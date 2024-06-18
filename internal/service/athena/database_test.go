// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaDatabase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "properties.%", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_properties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_properties(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
					resource.TestCheckResourceAttr(resourceName, "properties.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.creator", "Jane D."),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_acl(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_acl(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.0.s3_acl_option", "BUCKET_OWNER_FULL_CONTROL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, "acl_configuration", names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_kms(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_option", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy, names.AttrEncryptionConfiguration},
			},
		},
	})
}

func TestAccAthenaDatabase_nameStartsWithUnderscore(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := "_" + sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_nameCantHaveUppercase(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := "A" + sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_basic(rName, dbName, false),
				ExpectError: regexache.MustCompile(`must be lowercase letters, numbers, or underscore \('_'\)`),
			},
		},
	})
}

func TestAccAthenaDatabase_destroyFailsIfTablesExist(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					testAccDatabaseCreateTables(ctx, dbName),
					testAccCheckDatabaseDropFails(ctx, dbName),
					testAccDatabaseDestroyTables(ctx, dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_forceDestroyAlwaysSucceeds(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					testAccDatabaseCreateTables(ctx, dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_comment(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "athena is a goddess"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_unescaped_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_unescapedComment(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, dbName),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, "athena's a goddess"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrBucket, names.AttrForceDestroy},
			},
		},
	})
}

func TestAccAthenaDatabase_disppears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)

	resourceName := "aws_athena_database.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfathena.ResourceDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_database" {
				continue
			}

			_, err := tfathena.FindDatabaseByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Athena Database %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckDatabaseExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		_, err := tfathena.FindDatabaseByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDatabaseCreateTables(ctx context.Context, dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &types.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf(
				"create external table foo (bar int) location 's3://%s/';", bucketName)),
			ResultConfiguration: &types.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		output, err := conn.StartQueryExecution(ctx, input)

		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(ctx, conn, aws.ToString(output.QueryExecutionId))

		return err
	}
}

func testAccDatabaseDestroyTables(ctx context.Context, dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &types.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String("drop table foo;"),
			ResultConfiguration: &types.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		output, err := conn.StartQueryExecution(ctx, input)

		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(ctx, conn, aws.ToString(output.QueryExecutionId))

		return err
	}
}

func testAccCheckDatabaseDropFails(ctx context.Context, dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &types.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf("drop database `%s`;", dbName)),
			ResultConfiguration: &types.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		output, err := conn.StartQueryExecution(ctx, input)

		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(ctx, conn, aws.ToString(output.QueryExecutionId))

		if err == nil {
			return fmt.Errorf("drop database unexpectedly succeeded for a database with tables")
		}

		return nil
	}
}

func testAccDatabaseFindBucketName(s *terraform.State, dbName string) (bucket string, err error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aws_athena_database" && rs.Primary.Attributes[names.AttrName] == dbName {
			bucket = rs.Primary.Attributes[names.AttrBucket]
			break
		}
	}

	if bucket == "" {
		err = fmt.Errorf("cannot find database %s", dbName)
	}

	return bucket, err
}

func testAccDatabaseConfig_basic(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  force_destroy = %[3]t
}
`, rName, dbName, forceDestroy)
}

func testAccDatabaseConfig_properties(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  force_destroy = %[3]t

  properties = {
    creator = "Jane D."
  }
}
`, rName, dbName, forceDestroy)
}

func testAccDatabaseConfig_acl(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  force_destroy = %[3]t

  acl_configuration {
    s3_acl_option = "BUCKET_OWNER_FULL_CONTROL"
  }
}
`, rName, dbName, forceDestroy)
}

func testAccDatabaseConfig_kms(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 10
  description             = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_athena_database" "test" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  force_destroy = %[3]t

  encryption_configuration {
    encryption_option = "SSE_KMS"
    kms_key           = aws_kms_key.test.arn
  }
}
`, rName, dbName, forceDestroy)
}

func testAccDatabaseConfig_comment(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  comment       = "athena is a goddess"
  force_destroy = %[3]t
}
`, rName, dbName, forceDestroy)
}

func testAccDatabaseConfig_unescapedComment(rName string, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_athena_database" "test" {
  name          = %[2]q
  bucket        = aws_s3_bucket.test.bucket
  comment       = "athena's a goddess"
  force_destroy = %[3]t
}
`, rName, dbName, forceDestroy)
}
