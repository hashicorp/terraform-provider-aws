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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3BucketCorsConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketCorsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCorsConfigurationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
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

func TestAccS3BucketCorsConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketCorsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCorsConfigurationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketCorsConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketCorsConfiguration_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketCorsConfigurationDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccBucketCorsConfigurationCompleteConfig_SingleRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
						"expose_headers.#":  "1",
						"id":                rName,
						"max_age_seconds":   "3000",
					}),
				),
			},
			{
				Config: testAccBucketCorsConfigurationConfig_MultipleRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketCorsConfigurationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
				),
			},
		},
	})
}

func TestAccS3BucketCorsConfiguration_SingleRule(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketCorsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCorsConfigurationCompleteConfig_SingleRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
						"expose_headers.#":  "1",
						"id":                rName,
						"max_age_seconds":   "3000",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.expose_headers.*", "ETag"),
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

func TestAccS3BucketCorsConfiguration_MultipleRules(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketCorsConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketCorsConfigurationConfig_MultipleRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "3",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_headers.*", "*"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "*"),
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

func TestAccS3BucketCorsConfiguration_migrate_corsRuleNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withCORS(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCorsConfigurationConfig_Migrate_CorsRuleNoChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_headers.#": "1",
						"allowed_methods.#": "2",
						"allowed_origins.#": "1",
						"expose_headers.#":  "2",
						"max_age_seconds":   "3000",
					}),
				),
			},
		},
	})
}

func TestAccS3BucketCorsConfiguration_migrate_corsRuleWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_cors_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_withCORS(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_headers.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_methods.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.allowed_origins.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.expose_headers.#", "2"),
					resource.TestCheckResourceAttr(bucketResourceName, "cors_rule.0.max_age_seconds", "3000"),
				),
			},
			{
				Config: testAccBucketCorsConfigurationConfig_Migrate_CorsRuleWithChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketCorsConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cors_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "cors_rule.*", map[string]string{
						"allowed_methods.#": "1",
						"allowed_origins.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_rule.*.allowed_origins.*", "https://www.example.com"),
				),
			},
		},
	})
}

func testAccCheckBucketCorsConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_cors_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketCorsInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketCors(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, tfs3.ErrCodeNoSuchCORSConfiguration) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket CORS configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 Bucket CORS configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketCorsConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetBucketCorsInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		corsResponse, err := tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			return conn.GetBucketCors(input)
		}, tfs3.ErrCodeNoSuchCORSConfiguration)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket CORS configuration (%s): %w", rs.Primary.ID, err)
		}

		if output, ok := corsResponse.(*s3.GetBucketCorsOutput); !ok || output == nil || len(output.CORSRules) == 0 {
			return fmt.Errorf("S3 Bucket CORS configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketCorsConfigurationBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}

func testAccBucketCorsConfigurationCompleteConfig_SingleRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["ETag"]
    id              = %[1]q
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCorsConfigurationConfig_MultipleRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST", "DELETE"]
    allowed_origins = ["https://www.example.com"]
  }

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
  }
}
`, rName)
}

func testAccBucketCorsConfigurationConfig_Migrate_CorsRuleNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = ["https://www.example.com"]
    expose_headers  = ["x-amz-server-side-encryption", "ETag"]
    max_age_seconds = 3000
  }
}
`, rName)
}

func testAccBucketCorsConfigurationConfig_Migrate_CorsRuleWithChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_cors_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  cors_rule {
    allowed_methods = ["PUT"]
    allowed_origins = ["https://www.example.com"]
  }
}
`, rName)
}
