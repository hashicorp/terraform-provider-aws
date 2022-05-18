package s3_test

import (
	"fmt"
	"testing"
	"time"

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

func TestAccS3BucketPublicAccessBlock_basic(t *testing.T) {
	var config s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("aws_s3_bucket.bucket"),
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "bucket", name),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
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

func TestAccS3BucketPublicAccessBlock_disappears(t *testing.T) {
	var config s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config),
					testAccCheckBucketPublicAccessBlockDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_Disappears_bucket(t *testing.T) {
	var config s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"
	bucketResourceName := "aws_s3_bucket.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucket(), bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_blockPublicACLs(t *testing.T) {
	var config1, config2, config3 s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "true", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config2),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "true", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config3),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_blockPublicPolicy(t *testing.T) {
	var config1, config2, config3 s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "true", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config2),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "true", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config3),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_ignorePublicACLs(t *testing.T) {
	var config1, config2, config3 s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "true", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config2),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "true", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config3),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
		},
	})
}

func TestAccS3BucketPublicAccessBlock_restrictPublicBuckets(t *testing.T) {
	var config1, config2, config3 s3.PublicAccessBlockConfiguration
	name := fmt.Sprintf("tf-test-bucket-%d", sdkacctest.RandInt())
	resourceName := "aws_s3_bucket_public_access_block.bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config1),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config2),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
				),
			},
			{
				Config: testAccBucketPublicAccessBlockConfig(name, "false", "false", "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketPublicAccessBlockExists(resourceName, &config3),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
		},
	})
}

func testAccCheckBucketPublicAccessBlockExists(n string, config *s3.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		input := &s3.GetPublicAccessBlockInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		var output *s3.GetPublicAccessBlockOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.GetPublicAccessBlock(input)

			if tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchPublicAccessBlockConfiguration) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		if output == nil || output.PublicAccessBlockConfiguration == nil {
			return fmt.Errorf("S3 Bucket Public Access Block not found")
		}

		*config = *output.PublicAccessBlockConfiguration

		return nil
	}
}

func testAccCheckBucketPublicAccessBlockDisappears(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		deleteInput := &s3.DeletePublicAccessBlockInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeletePublicAccessBlock(deleteInput); err != nil {
			return err
		}

		getInput := &s3.GetPublicAccessBlockInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		return resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.GetPublicAccessBlock(getInput)

			if tfawserr.ErrCodeEquals(err, tfs3.ErrCodeNoSuchPublicAccessBlockConfiguration) {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("S3 Bucket Public Access Block (%s) still exists", rs.Primary.ID))
		})
	}
}

func testAccBucketPublicAccessBlockConfig(bucketName, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q

  tags = {
    TestName = %[1]q
  }
}

resource "aws_s3_bucket_public_access_block" "bucket" {
  bucket = aws_s3_bucket.bucket.bucket

  block_public_acls       = %[2]q
  block_public_policy     = %[3]q
  ignore_public_acls      = %[4]q
  restrict_public_buckets = %[5]q
}
`, bucketName, blockPublicAcls, blockPublicPolicy, ignorePublicAcls, restrictPublicBuckets)
}
