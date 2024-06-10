// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3BucketWebsiteConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3.ResourceBucketWebsiteConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
					resource.TestCheckResourceAttr(resourceName, "error_document.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_redirect(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucket, "aws_s3_bucket.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "redirect_all_requests_to.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleOptionalRedirection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                        acctest.Ct1,
						"condition.0.key_prefix_equals":      "docs/",
						"redirect.#":                         acctest.Ct1,
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
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#": acctest.Ct1,
						"condition.0.http_error_code_returned_equals": "404",
						"redirect.#":                         acctest.Ct1,
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
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   acctest.Ct1,
						"condition.0.key_prefix_equals": "images/",
						"redirect.#":                    acctest.Ct1,
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleMultipleRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   acctest.Ct1,
						"condition.0.key_prefix_equals": "docs/",
						"redirect.#":                    acctest.Ct1,
						"redirect.0.replace_key_with":   "errorpage.html",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"condition.#":                   acctest.Ct1,
						"condition.0.key_prefix_equals": "images/",
						"redirect.#":                    acctest.Ct1,
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
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_RoutingRule_RedirectOnly(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleRedirectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "routing_rule.*", map[string]string{
						"redirect.#":                  acctest.Ct1,
						"redirect.0.protocol":         string(types.ProtocolHttps),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_RoutingRuleToRoutingRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRuleOptionalRedirection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_routingRulesConditionAndRedirect(rName, "documents/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "routing_rules"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithIndexDocumentNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_website(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "website.0.index_document", "index.html"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentNoChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "index.html"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithIndexDocumentWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_website(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttr(bucketResourceName, "website.0.index_document", "index.html"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateIndexDocumentChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "index_document.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "index_document.0.suffix", "other.html"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithRoutingRuleNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRoutingRules(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(bucketResourceName, "website.0.routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleNoChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_migrate_websiteWithRoutingRuleWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketResourceName := "aws_s3_bucket.test"
	resourceName := "aws_s3_bucket_website_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_websiteAndRoutingRules(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(ctx, bucketResourceName),
					resource.TestCheckResourceAttr(bucketResourceName, "website.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(bucketResourceName, "website.0.routing_rules"),
				),
			},
			{
				Config: testAccBucketWebsiteConfigurationConfig_migrateRoutingRuleChange(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketWebsiteConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.0.protocol", string(types.ProtocolHttps)),
					resource.TestCheckResourceAttr(resourceName, "routing_rule.0.redirect.0.replace_key_with", "errorpage.html"),
				),
			},
		},
	})
}

func TestAccS3BucketWebsiteConfiguration_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketWebsiteConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccBucketWebsiteConfigurationConfig_directoryBucket(rName),
				ExpectError: regexache.MustCompile(`directory buckets are not supported`),
			},
		},
	})
}

func testAccCheckBucketWebsiteConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_bucket_website_configuration" {
				continue
			}

			bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfs3.FindBucketWebsite(ctx, conn, bucket, expectedBucketOwner)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Bucket Website Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketWebsiteConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		bucket, expectedBucketOwner, err := tfs3.ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfs3.FindBucketWebsite(ctx, conn, bucket, expectedBucketOwner)

		return err
	}
}

func testAccBucketWebsiteConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
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
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
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

func testAccBucketWebsiteConfigurationConfig_directoryBucket(rName string) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), `
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_bucket_website_configuration" "test" {
  bucket = aws_s3_directory_bucket.test.bucket
  index_document {
    suffix = "index.html"
  }
}
`)
}
