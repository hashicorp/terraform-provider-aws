package athena_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
)

func TestAccAthenaDatabase_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "properties.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_properties(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_properties(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
					resource.TestCheckResourceAttr(resourceName, "properties.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "properties.creator", "Jane D."),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_acl(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_acl(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "acl_configuration.0.s3_acl_option", "BUCKET_OWNER_FULL_CONTROL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "acl_configuration", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_encryption(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_kms(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_option", "SSE_KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", "aws_kms_key.test", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy", "encryption_configuration"},
			},
		},
	})
}

func TestAccAthenaDatabase_nameStartsWithUnderscore(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := "_" + sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_nameCantHaveUppercase(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := "A" + sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatabaseConfig_basic(rName, dbName, false),
				ExpectError: regexp.MustCompile(`must be lowercase letters, numbers, or underscore \('_'\)`),
			},
		},
	})
}

func TestAccAthenaDatabase_destroyFailsIfTablesExist(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.test"),
					testAccDatabaseCreateTables(dbName),
					testAccCheckDatabaseDropFails(dbName),
					testAccDatabaseDestroyTables(dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_forceDestroyAlwaysSucceeds(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.test"),
					testAccDatabaseCreateTables(dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_comment(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
					resource.TestCheckResourceAttr(resourceName, "comment", "athena is a goddess"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_unescaped_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)
	resourceName := "aws_athena_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_unescapedComment(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dbName),
					resource.TestCheckResourceAttr(resourceName, "comment", "athena's a goddess"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bucket", "force_destroy"},
			},
		},
	})
}

func TestAccAthenaDatabase_disppears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dbName := sdkacctest.RandString(8)

	resourceName := "aws_athena_database.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, athena.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseConfig_basic(rName, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfathena.ResourceDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatabaseDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_database" {
			continue
		}

		input := &athena.ListDatabasesInput{
			CatalogName: aws.String("AwsDataCatalog"),
		}

		res, err := conn.ListDatabases(input)
		if err != nil {
			return err
		}

		var database *athena.Database
		for _, db := range res.DatabaseList {
			if aws.StringValue(db.Name) == rs.Primary.ID {
				database = db
				break
			}
		}

		if database != nil {
			return fmt.Errorf("Athena database (%s) still exists", rs.Primary.ID)
		}

	}
	return nil
}

func testAccCheckDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s, %v", name, s.RootModule().Resources)
		}
		return nil
	}
}

func testAccDatabaseCreateTables(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf(
				"create external table foo (bar int) location 's3://%s/';", bucketName)),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		return err
	}
}

func testAccDatabaseDestroyTables(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String("drop table foo;"),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		return err
	}
}

func testAccCheckDatabaseDropFails(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf("drop database `%s`;", dbName)),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		if err == nil {
			return fmt.Errorf("drop database unexpectedly succeeded for a database with tables")
		}

		return nil
	}
}

func testAccDatabaseFindBucketName(s *terraform.State, dbName string) (bucket string, err error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aws_athena_database" && rs.Primary.Attributes["name"] == dbName {
			bucket = rs.Primary.Attributes["bucket"]
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
