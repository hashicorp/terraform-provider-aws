package s3_test

import (
	"fmt"
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

func TestAccS3BucketRequestPaymentConfiguration_Basic_BucketOwner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketRequestPaymentConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerBucketOwner),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "payer", s3.PayerBucketOwner),
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

func TestAccS3BucketRequestPaymentConfiguration_Basic_Requester(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketRequestPaymentConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "payer", s3.PayerRequester),
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

func TestAccS3BucketRequestPaymentConfiguration_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketRequestPaymentConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerBucketOwner),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
				),
			},
		},
	})
}

func TestAccS3BucketRequestPaymentConfiguration_migrate_noChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketRequestPaymentConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_requestPayer(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "request_payer", s3.PayerRequester),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", s3.PayerRequester),
				),
			},
		},
	})
}

func TestAccS3BucketRequestPaymentConfiguration_migrate_withChange(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_request_payment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketRequestPaymentConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_requestPayer(rName, s3.PayerRequester),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "request_payer", s3.PayerRequester),
				),
			},
			{
				Config: testAccBucketRequestPaymentConfigurationConfig_basic(rName, s3.PayerBucketOwner),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketRequestPaymentConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "payer", s3.PayerBucketOwner),
				),
			},
		},
	})
}

func testAccCheckBucketRequestPaymentConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_request_payment_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketRequestPaymentInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketRequestPayment(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 bucket request payment configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && aws.StringValue(output.Payer) != s3.PayerBucketOwner {
			return fmt.Errorf("S3 bucket request payment configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketRequestPaymentConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetBucketRequestPaymentInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketRequestPayment(input)

		if err != nil {
			return fmt.Errorf("error getting S3 bucket request payment configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket request payment configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketRequestPaymentConfigurationConfig_basic(rName, payer string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_request_payment_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  payer  = %[2]q
}
`, rName, payer)
}
