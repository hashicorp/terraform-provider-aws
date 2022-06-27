package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccS3ObjectCopy_basic(t *testing.T) {
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"
	sourceName := "aws_s3_object.source"
	key := "HundBegraven"
	sourceKey := "WshngtnNtnls"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName2),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
					resource.TestCheckResourceAttrPair(resourceName, "etag", sourceName, "etag"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_BucketKeyEnabled_bucket(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_bucketKeyEnabledBucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "true"),
				),
			},
		},
	})
}

func TestAccS3ObjectCopy_BucketKeyEnabled_object(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_object_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_bucketKeyEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket_key_enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckObjectCopyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_object_copy" {
			continue
		}

		_, err := conn.HeadObject(
			&s3.HeadObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err == nil {
			return fmt.Errorf("AWS S3 Object still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckObjectCopyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Object ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
		_, err := conn.GetObject(
			&s3.GetObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err != nil {
			return fmt.Errorf("S3 Object error: %s", err)
		}

		return nil
	}
}

func testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, key string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = %[1]q
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = %[2]q
  content = "Ingen ko på isen"
}

resource "aws_s3_bucket" "target" {
  bucket = %[3]q
}

resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[4]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"

  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
    type        = "Group"
    permissions = ["READ"]
  }
}
`, rName1, sourceKey, rName2, key)
}

func testAccObjectCopyConfig_bucketKeyEnabledBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  content = "Ingen ko på isen"
  key     = "test"
}

resource "aws_s3_bucket" "target" {
  bucket = "%[1]s-target"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.target.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_object_copy" "test" {
  # Must have bucket SSE enabled first
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  bucket = aws_s3_bucket.target.bucket
  key    = "test"
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, rName)
}

func testAccObjectCopyConfig_bucketKeyEnabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Encrypts test objects"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"
}

resource "aws_s3_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  content = "Ingen ko på isen"
  key     = "test"
}

resource "aws_s3_bucket" "target" {
  bucket = "%[1]s-target"
}

resource "aws_s3_object_copy" "test" {
  bucket             = aws_s3_bucket.target.bucket
  bucket_key_enabled = true
  key                = "test"
  kms_key_id         = aws_kms_key.test.arn
  source             = "${aws_s3_bucket.source.bucket}/${aws_s3_object.source.key}"
}
`, rName)
}
