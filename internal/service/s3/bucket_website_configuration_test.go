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

func TestAccS3BucketWebsiteConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttrSet(resourceName, "website_domain"),
					resource.TestCheckResourceAttrSet(resourceName, "website_endpoint"),
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

func TestAccS3BucketWebsiteConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketWebsiteConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_update(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "error_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "error_document.0.key", "error.html"),
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

func TestAccS3BucketWebsiteConfiguration_Redirect(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_redirect(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3_bucket.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "redirect_all_requests_to.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redirect_all_requests_to.0.host_name", "example.com"),
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

func TestAccS3BucketWebsiteConfiguration_RoutingRule_ConditionAndRedirect(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleOptionalRedirection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                        "1",
						"condition.0.key_prefix_equals":      "docs/",
						"redirect.#":                         "1",
						"redirect.0.replace_key_prefix_with": "documents/",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleRedirectErrors(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#": "1",
						"condition.0.http_error_code_returned_equals": "404",
						"redirect.#":                         "1",
						"redirect.0.replace_key_prefix_with": "report-404",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleRedirectToPage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   "1",
						"condition.0.key_prefix_equals": "images/",
						"redirect.#":                    "1",
						"redirect.0.replace_key_with":   "errorpage.html",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
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

func TestAccS3BucketWebsiteConfiguration_RoutingRule_MultipleRules(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleMultipleRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   "1",
						"condition.0.key_prefix_equals": "docs/",
						"redirect.#":                    "1",
						"redirect.0.replace_key_with":   "errorpage.html",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   "1",
						"condition.0.key_prefix_equals": "images/",
						"redirect.#":                    "1",
						"redirect.0.replace_key_with":   "errorpage.html",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_RoutingRule_RedirectOnly(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleRedirectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"redirect.#":                  "1",
						"redirect.0.protocol":         s3.ProtocolHttps,
						"redirect.0.replace_key_with": "errorpage.html",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
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

func TestAccS3BucketWebsiteConfiguration_RoutingRules_ConditionAndRedirect(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
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

func TestAccS3BucketWebsiteConfiguration_RoutingRules_ConditionAndRedirectWithEmptyString(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
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

func TestAccS3BucketWebsiteConfiguration_RoutingRules_updateConditionAndRedirect(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_RoutingRuleToRoutingRules(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketWebsiteConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleOptionalRedirection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithIndexDocumentNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_website(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "website.0.index_document", "index.html"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentNoChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithIndexDocumentWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_website(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", "1"),
					resource.TestCheckResourceAttr(bucketResourceName, "website.0.index_document", "index.html"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "other.html"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithRoutingRuleNoChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRoutingRules(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", "1"),
					resource.TestCheckResourceAttrSet(bucketResourceName, "website.0.routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleNoChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithRoutingRuleWithChange(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRoutingRules(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", "1"),
					resource.TestCheckResourceAttrSet(bucketResourceName, "website.0.routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.0.protocol", s3.ProtocolHttps),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.0.replace_key_with", "errorpage.html"),
				),
			},
		},
	})
}

func testAccCheckBucketWebsiteConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_website_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketWebsiteInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketWebsite(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, tfs3.ErrCodeNoSuchWebsiteConfiguration) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 bucket website configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 bucket website configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckBucketWebsiteConfigurationExists(resourceName string) resource.TestCheckFunc {
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

		input := &s3.GetBucketWebsiteInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketWebsite(input)

		if err != nil {
			return fmt.Errorf("error getting S3 bucket website configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("S3 Bucket website configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBucketWebsiteConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  index_document {
    suffix = "index.html"
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}


resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_redirect(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}


resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id
  redirect_all_requests_to {
    host_name = "example.com"
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_routingRuleOptionalRedirection(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}


resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      key_prefix_equals = "docs/"
    }
    redirect {
      replace_key_prefix_with = "documents/"
    }
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_routingRuleRedirectErrors(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHVMEBSAMI(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}


resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      http_error_code_returned_equals = "404"
    }
    redirect {
      replace_key_prefix_with = "report-404"
    }
  }
}
`, rName))
}

func testAccBucketWebsiteConfigurationConfig_routingRuleRedirectToPage(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}


resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      key_prefix_equals = "images/"
    }
    redirect {
      replace_key_with = "errorpage.html"
    }
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_routingRuleRedirectOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    redirect {
      protocol         = "https"
      replace_key_with = "errorpage.html"
    }
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_routingRuleMultipleRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      key_prefix_equals = "images/"
    }
    redirect {
      replace_key_with = "errorpage.html"
    }
  }

  routing_rule {
    condition {
      key_prefix_equals = "docs/"
    }
    redirect {
      replace_key_with = "errorpage.html"
    }
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(bucketName, replaceKeyPrefixWith string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rules = <<EOF
[
  {
    "Condition": {
      "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
      "ReplaceKeyPrefixWith": %[2]q
    }
  }
]
EOF
}
`, bucketName, replaceKeyPrefixWith)
}

func testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "other.html"
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleNoChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    condition {
      key_prefix_equals = "docs/"
    }
    redirect {
      replace_key_prefix_with = "documents/"
    }
  }
}
`, rName)
}

func testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleChange(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "bucket" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }

  routing_rule {
    redirect {
      protocol         = "https"
      replace_key_with = "errorpage.html"
    }
  }
}
`, rName)
}
