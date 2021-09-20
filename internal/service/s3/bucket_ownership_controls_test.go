package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccAWSS3BucketOwnershipControls_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSS3BucketOwnershipControlsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketOwnershipControlsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipBucketOwnerPreferred),
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

func TestAccAWSS3BucketOwnershipControls_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSS3BucketOwnershipControlsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketOwnershipControlsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketOwnershipControls(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3BucketOwnershipControls_disappears_Bucket(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_ownership_controls.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSS3BucketOwnershipControlsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketOwnershipControlsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), s3BucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3BucketOwnershipControls_Rule_ObjectOwnership(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_bucket_ownership_controls.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSS3BucketOwnershipControlsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, s3.ObjectOwnershipObjectWriter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketOwnershipControlsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipObjectWriter),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, s3.ObjectOwnershipBucketOwnerPreferred),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketOwnershipControlsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.object_ownership", s3.ObjectOwnershipBucketOwnerPreferred),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketOwnershipControlsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_ownership_controls" {
			continue
		}

		input := &s3.GetBucketOwnershipControlsInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetBucketOwnershipControls(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Bucket Ownership Controls (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSS3BucketOwnershipControlsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		input := &s3.GetBucketOwnershipControlsInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetBucketOwnershipControls(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSS3BucketOwnershipControlsConfig_Rule_ObjectOwnership(rName, objectOwnership string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    object_ownership = %[2]q
  }
}
`, rName, objectOwnership)
}
