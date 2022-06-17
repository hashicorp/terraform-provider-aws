package s3_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketLogging_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
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

func TestAccS3BucketLogging_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketLogging(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketLogging_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	targetBucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
				),
			},
			{
				// Test updating "target_prefix"
				Config: testAccBucketLoggingConfig_update(rName, rName, "tmp/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "tmp/"),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test updating "target_bucket" and "target_prefix"
				Config: testAccBucketLoggingConfig_update(rName, targetBucketName, "log/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
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

func TestAccS3BucketLogging_TargetGrantByID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByID(rName, s3.BucketLogsPermissionFullControl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.BucketLogsPermissionFullControl,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.id", "data.aws_canonical_user_id.current", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.display_name", "data.aws_canonical_user_id.current", "display_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByID(rName, s3.BucketLogsPermissionRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeCanonicalUser,
						"permission":     s3.BucketLogsPermissionRead,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "target_grant.*.grantee.0.display_name", "data.aws_canonical_user_id.current", "display_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_TargetGrantByEmail(t *testing.T) {
	rEmail, ok := os.LookupEnv("AWS_S3_BUCKET_LOGGING_AMAZON_CUSTOMER_BY_EMAIL")

	if !ok {
		acctest.Skip(t, "'AWS_S3_BUCKET_LOGGING_AMAZON_CUSTOMER_BY_EMAIL' not set, skipping test.")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByEmail(rName, rEmail, s3.BucketLogsPermissionFullControl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":               "1",
						"grantee.0.email_address": rEmail,
						"grantee.0.type":          s3.TypeAmazonCustomerByEmail,
						"permission":              s3.BucketLogsPermissionFullControl,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByEmail(rName, rEmail, s3.BucketLogsPermissionRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":       "1",
						"grantee.0.email": rEmail,
						"grantee.0.type":  s3.TypeAmazonCustomerByEmail,
						"permission":      s3.BucketLogsPermissionRead,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_TargetGrantByGroup(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketLoggingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketLoggingConfig_targetGrantByGroup(rName, s3.BucketLogsPermissionFullControl),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeGroup,
						"permission":     s3.BucketLogsPermissionFullControl,
					}),
					testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_targetGrantByGroup(rName, s3.BucketLogsPermissionRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
						"grantee.#":      "1",
						"grantee.0.type": s3.TypeGroup,
						"permission":     s3.BucketLogsPermissionRead,
					}),
					testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_grant.#", "0"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_migrate_loggingNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_logging(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.#", "1"),
					resource.TestCheckResourceAttrPair(bucketResourceName, "logging.0.target_bucket", "aws_s3_bucket.log_bucket", "id"),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.0.target_prefix", "log/"),
				),
			},
			{
				Config: testAccBucketLoggingConfig_migrate(bucketName, "log/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "log/"),
				),
			},
		},
	})
}

func TestAccS3BucketLogging_migrate_loggingWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_logging(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.#", "1"),
					resource.TestCheckResourceAttrPair(bucketResourceName, "logging.0.target_bucket", "aws_s3_bucket.log_bucket", "id"),
					resource.TestCheckResourceAttr(bucketResourceName, "logging.0.target_prefix", "log/"),
				),
			},
			{
				Config: testAccBucketLoggingConfig_migrate(bucketName, "tmp/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketLoggingExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "target_bucket", "aws_s3_bucket.log_bucket", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_prefix", "tmp/"),
				),
			},
		},
	})
}

func testAccCheckBucketLoggingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_logging" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketLoggingInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketLogging(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Logging (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && output.LoggingEnabled != nil {
			return fmt.Errorf("S3 Bucket Logging (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketLoggingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketLoggingInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketLogging(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Logging (%s): %w", rs.Primary.ID, err)
		}

		if output == nil || output.LoggingEnabled == nil {
			return fmt.Errorf("S3 Bucket Logging (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketLoggingTargetGrantGranteeURI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		uri := fmt.Sprintf("http://acs.%s/groups/s3/LogDelivery", acctest.PartitionDNSSuffix())
		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_grant.*", map[string]string{
			"grantee.0.uri": uri,
		})(s)
	}
}

func testAccBucketLoggingConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"
}
`, rName)
}

func testAccBucketLoggingConfig_update(rName, targetBucketName, targetPrefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[2]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = %[3]q
}
`, rName, targetBucketName, targetPrefix)
}

func testAccBucketLoggingConfig_targetGrantByID(rName, permission string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}


resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}


resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      id   = data.aws_canonical_user_id.current.id
      type = "CanonicalUser"
    }
    permission = %[2]q
  }
}
`, rName, permission)
}

func testAccBucketLoggingConfig_targetGrantByEmail(rName, email, permission string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      email_address = %[2]q
      type          = "AmazonCustomerByEmail"
    }
    permission = %[3]q
  }
}
`, rName, email, permission)
}

func testAccBucketLoggingConfig_targetGrantByGroup(rName, permission string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = "log/"

  target_grant {
    grantee {
      type = "Group"
      # Test with S3 log delivery group
      uri = "http://acs.${data.aws_partition.current.dns_suffix}/groups/s3/LogDelivery"
    }
    permission = %[2]q
  }
}
`, rName, permission)
}

func testAccBucketLoggingConfig_migrate(rName, targetPrefix string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "log_bucket" {
  bucket = "%[1]s-log"
}

resource "aws_s3_bucket_acl" "log_bucket_acl" {
  bucket = aws_s3_bucket.log_bucket.id
  acl    = "log-delivery-write"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_logging" "test" {
  bucket = aws_s3_bucket.test.id

  target_bucket = aws_s3_bucket.log_bucket.id
  target_prefix = %[2]q
}
`, rName, targetPrefix)
}
