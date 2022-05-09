package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccS3BucketAccelerateConfiguration_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusEnabled),
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

func TestAccS3BucketAccelerateConfiguration_update(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusEnabled),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusSuspended),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusSuspended),
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

func TestAccS3BucketAccelerateConfiguration_disappears(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(bucketName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketAccelerateConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketAccelerateConfiguration_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withAcceleration(rName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acceleration_status", s3.BucketAccelerateStatusEnabled),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(rName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusEnabled),
				),
			},
		},
	})
}

func TestAccS3BucketAccelerateConfiguration_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_accelerate_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketAccelerateConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withAcceleration(rName, s3.BucketAccelerateStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "acceleration_status", s3.BucketAccelerateStatusEnabled),
				),
			},
			{
				Config: testAccBucketAccelerateConfigurationBasicConfig(rName, s3.BucketAccelerateStatusSuspended),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketAccelerateConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "status", s3.BucketAccelerateStatusSuspended),
				),
			},
		},
	})
}

func testAccCheckBucketAccelerateConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_accelerate_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketAccelerateConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketAccelerateConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket accelerate configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 Bucket accelerate configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketAccelerateConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetBucketAccelerateConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketAccelerateConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket accelerate configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket accelerate configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketAccelerateConfigurationBasicConfig(bucketName, status string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_accelerate_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  status = %[2]q
}
`, bucketName, status)
}
