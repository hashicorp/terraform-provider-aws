package s3_test

import (
	"fmt"
	"regexp"
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

func TestAccS3BucketVersioning_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
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

func TestAccS3BucketVersioning_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketVersioning(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_disappears_bucket(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketVersioning_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusSuspended),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
		},
	})
}

// TestAccBucketVersioning_MFADelete can only test for a "Disabled"
// mfa_delete configuration as the "mfa" argument is required if it's enabled
func TestAccS3BucketVersioning_MFADelete(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_mfaDelete(rName, s3.MFADeleteDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.mfa_delete", s3.MFADeleteDisabled),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
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

func TestAccS3BucketVersioning_migrate_versioningDisabledNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", "false"),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningDisabledWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", "false"),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningEnabledNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", "true"),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_migrate_versioningEnabledWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioning(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.enabled", "true"),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateEnabled(bucketName, s3.BucketVersioningStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusSuspended),
				),
			},
		},
	})
}

// TestAccS3BucketVersioning_migrate_mfaDeleteNoChange can only test for a "Disabled"
// mfa_delete configuration as the "mfa" argument is required if it's enabled
func TestAccS3BucketVersioning_migrate_mfaDeleteNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_versioningMFADelete(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "versioning.0.mfa_delete", "false"),
				),
			},
			{
				Config: testAccBucketVersioningConfig_migrateMFADelete(bucketName, s3.MFADeleteDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.mfa_delete", s3.MFADeleteDisabled),
				),
			},
		},
	})
}

func TestAccS3BucketVersioning_Status_disabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
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

func TestAccS3BucketVersioning_Status_disabledToEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
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

func TestAccS3BucketVersioning_Status_disabledToSuspended(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", tfs3.BucketVersioningStatusDisabled),
				),
			},
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusSuspended),
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

func TestAccS3BucketVersioning_Status_enabledToDisabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusEnabled),
				),
			},
			{
				Config:      testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				ExpectError: regexp.MustCompile(`versioning_configuration.status cannot be updated from 'Enabled' to 'Disabled'`),
			},
		},
	})
}

func TestAccS3BucketVersioning_Status_suspendedToDisabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_versioning.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketVersioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketVersioningConfig_basic(rName, s3.BucketVersioningStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketVersioningExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "versioning_configuration.0.status", s3.BucketVersioningStatusSuspended),
				),
			},
			{
				Config:      testAccBucketVersioningConfig_basic(rName, tfs3.BucketVersioningStatusDisabled),
				ExpectError: regexp.MustCompile(`versioning_configuration.status cannot be updated from 'Suspended' to 'Disabled'`),
			},
		},
	})
}

func testAccCheckBucketVersioningDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_versioning" {
			continue
		}

		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketVersioning(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 bucket versioning (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && aws.StringValue(output.Status) == s3.ReplicationRuleStatusEnabled {
			return fmt.Errorf("S3 bucket versioning (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketVersioningExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetBucketVersioning(input)

		if err != nil {
			return fmt.Errorf("error getting S3 bucket versioning (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket versioning (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketVersioningConfig_basic(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketVersioningConfig_mfaDelete(rName, mfaDelete string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    mfa_delete = %[2]q
    status     = "Enabled"
  }
}
`, rName, mfaDelete)
}

func testAccBucketVersioningConfig_migrateEnabled(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = %[2]q
  }
}
`, rName, status)
}

func testAccBucketVersioningConfig_migrateMFADelete(rName, mfaDelete string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    mfa_delete = %[2]q
    status     = "Enabled"
  }
}
`, rName, mfaDelete)
}
