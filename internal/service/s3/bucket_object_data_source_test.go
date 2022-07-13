package s3_test

// WARNING: This code is DEPRECATED and will be removed in a future release!!
// DO NOT apply fixes or enhancements to the data source in this file.
// INSTEAD, apply fixes and enhancements to the data source in "object_data_source_test.go".

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketObjectDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_basicViaAccessPoint(t *testing.T) {
	var dsObj, rObj s3.GetObjectOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	dataSourceName := "data.aws_s3_object.test"
	resourceName := "aws_s3_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_basicViaAccessPoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					testAccCheckObjectExists(resourceName, &rObj),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key", resourceName, "key"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_readableBody(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_readableBody(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "yes"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_kmsEncrypted(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_kmsEncrypted(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "22"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sse_kms_key_id", resourceName, "kms_key_id"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "Keep Calm and Carry On"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_bucketKeyEnabled(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_keyEnabled(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "22"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sse_kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_key_enabled", resourceName, "bucket_key_enabled"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "Keep Calm and Carry On"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_allParams(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "25"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_id", resourceName, "version_id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket_key_enabled", resourceName, "bucket_key_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_control", resourceName, "cache_control"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_disposition", resourceName, "content_disposition"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_encoding", resourceName, "content_encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_language", resourceName, "content_language"),
					// Encryption is off
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttr(dataSourceName, "sse_kms_key_id", ""),
					// Supported, but difficult to reproduce in short testing time
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_class", resourceName, "storage_class"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration", ""),
					// Currently unsupported in aws_s3_object resource
					resource.TestCheckResourceAttr(dataSourceName, "expires", ""),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_redirect_location", resourceName, "website_redirect"),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_objectLockLegalHoldOff(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_lockLegalHoldOff(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_objectLockLegalHoldOn(t *testing.T) {
	rInt := sdkacctest.RandInt()
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName := "data.aws_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_lockLegalHoldOn(rInt, retainUntilDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_leadingSlash(t *testing.T) {
	var rObj s3.GetObjectOutput
	var dsObj1, dsObj2, dsObj3 s3.GetObjectOutput

	resourceName := "aws_s3_object.object"
	dataSourceName1 := "data.aws_s3_object.obj1"
	dataSourceName2 := "data.aws_s3_object.obj2"
	dataSourceName3 := "data.aws_s3_object.obj3"

	rInt := sdkacctest.RandInt()
	resourceOnlyConf, conf := testAccBucketObjectDataSourceConfig_leadingSlash(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:test-config-funcs-correct-form
				Config: resourceOnlyConf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
				),
			},
			{ // nosemgrep:test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExistsDataSource(dataSourceName1, &dsObj1),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName1, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),

					testAccCheckObjectExistsDataSource(dataSourceName2, &dsObj2),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName2, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),

					testAccCheckObjectExistsDataSource(dataSourceName3, &dsObj3),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName3, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName3, "body", "yes"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_multipleSlashes(t *testing.T) {
	var rObj1, rObj2 s3.GetObjectOutput
	var dsObj1, dsObj2, dsObj3 s3.GetObjectOutput

	resourceName1 := "aws_s3_object.object1"
	resourceName2 := "aws_s3_object.object2"
	dataSourceName1 := "data.aws_s3_object.obj1"
	dataSourceName2 := "data.aws_s3_object.obj2"
	dataSourceName3 := "data.aws_s3_object.obj3"

	rInt := sdkacctest.RandInt()
	resourceOnlyConf, conf := testAccBucketObjectDataSourceConfig_multipleSlashes(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{ // nosemgrep:test-config-funcs-correct-form
				Config: resourceOnlyConf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName1, &rObj1),
					testAccCheckObjectExists(resourceName2, &rObj2),
				),
			},
			{ // nosemgrep:test-config-funcs-correct-form
				Config: conf,
				Check: resource.ComposeTestCheckFunc(

					testAccCheckObjectExistsDataSource(dataSourceName1, &dsObj1),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "content_type", resourceName1, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),

					testAccCheckObjectExistsDataSource(dataSourceName2, &dsObj2),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "content_type", resourceName1, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),

					testAccCheckObjectExistsDataSource(dataSourceName3, &dsObj3),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "content_type", resourceName2, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName3, "body", "no"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectDataSource_singleSlashAsKey(t *testing.T) {
	var dsObj s3.GetObjectOutput
	dataSourceName := "data.aws_s3_object.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectDataSourceConfig_singleSlashAsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
				),
			},
		},
	})
}

func testAccBucketObjectDataSourceConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = "tf-testing-obj-%[1]d"
  content = "Hello World"
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_basicViaAccessPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = "Hello World"
}

data "aws_s3_object" "test" {
  bucket = aws_s3_access_point.test.arn
  key    = aws_s3_object.test.key
}
`, rName)
}

func testAccBucketObjectDataSourceConfig_readableBody(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "tf-testing-obj-%[1]d-readable"
  content      = "yes"
  content_type = "text/plain"
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_kmsEncrypted(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_kms_key" "example" {
  description             = "TF Acceptance Test KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "tf-testing-obj-%[1]d-encrypted"
  content      = "Keep Calm and Carry On"
  content_type = "text/plain"
  kms_key_id   = aws_kms_key.example.arn
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_keyEnabled(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_kms_key" "example" {
  description             = "TF Acceptance Test KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_object" "object" {
  bucket             = aws_s3_bucket.object_bucket.bucket
  key                = "tf-testing-obj-%[1]d-encrypted"
  content            = "Keep Calm and Carry On"
  content_type       = "text/plain"
  kms_key_id         = aws_kms_key.example.arn
  bucket_key_enabled = true
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.object_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "tf-testing-obj-%[1]d-all-params"

  content             = <<CONTENT
{
  "msg": "Hi there!"
}
CONTENT
  content_type        = "application/unknown"
  cache_control       = "no-cache"
  content_disposition = "attachment"
  content_encoding    = "identity"
  content_language    = "en-GB"

  tags = {
    Key1 = "Value 1"
  }
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_lockLegalHoldOff(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.object_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "tf-testing-obj-%[1]d"
  content                       = "Hello World"
  object_lock_legal_hold_status = "OFF"
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt)
}

func testAccBucketObjectDataSourceConfig_lockLegalHoldOn(randInt int, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"

  object_lock_enabled = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.object_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  # Must have bucket versioning enabled first
  depends_on = [aws_s3_bucket_versioning.test]

  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "tf-testing-obj-%[1]d"
  content                       = "Hello World"
  force_destroy                 = true
  object_lock_legal_hold_status = "ON"
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = "%[2]s"
}

data "aws_s3_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_object.object.key
}
`, randInt, retainUntilDate)
}

func testAccBucketObjectDataSourceConfig_leadingSlash(randInt int) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "//tf-testing-obj-%[1]d-readable"
  content      = "yes"
  content_type = "text/plain"
}
`, randInt)

	both := fmt.Sprintf(`
%[1]s

data "aws_s3_object" "obj1" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "tf-testing-obj-%[2]d-readable"
}

data "aws_s3_object" "obj2" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "/tf-testing-obj-%[2]d-readable"
}

data "aws_s3_object" "obj3" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "//tf-testing-obj-%[2]d-readable"
}
`, resources, randInt)

	return resources, both
}

func testAccBucketObjectDataSourceConfig_multipleSlashes(randInt int) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_object" "object1" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "first//second///third//"
  content      = "yes"
  content_type = "text/plain"
}

# Without a trailing slash.
resource "aws_s3_object" "object2" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "/first////second/third"
  content      = "no"
  content_type = "text/plain"
}
`, randInt)

	both := fmt.Sprintf(`
%s

data "aws_s3_object" "obj1" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first/second/third/"
}

data "aws_s3_object" "obj2" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first//second///third//"
}

data "aws_s3_object" "obj3" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first/second/third"
}
`, resources)

	return resources, both
}

func testAccBucketObjectDataSourceConfig_singleSlashAsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/"
}
`, rName)
}
