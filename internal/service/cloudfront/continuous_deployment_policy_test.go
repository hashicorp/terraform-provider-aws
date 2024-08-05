// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultDomain = "www.example.com"
)

func TestAccCloudFrontContinuousDeploymentPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudfront.GetContinuousDeploymentPolicyOutput
	var stagingDistribution awstypes.Distribution
	var productionDistribution awstypes.Distribution
	resourceName := "aws_cloudfront_continuous_deployment_policy.test"
	stagingDistributionResourceName := "aws_cloudfront_distribution.staging"
	productionDistributionResourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContinuousDeploymentPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContinuousDeploymentPolicyConfig_init(defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, stagingDistributionResourceName, &stagingDistribution),
					testAccCheckDistributionExists(ctx, productionDistributionResourceName, &productionDistribution),
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
				),
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "staging_distribution_dns_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "staging_distribution_dns_names.0.quantity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "staging_distribution_dns_names.0.items.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "staging_distribution_dns_names.0.items.0", stagingDistributionResourceName, names.AttrDomainName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                  "SingleWeight",
						"single_weight_config.#":        acctest.Ct1,
						"single_weight_config.0.weight": "0.01",
					}),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(productionDistributionResourceName, "continuous_deployment_policy_id", resourceName, names.AttrID),
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

func TestAccCloudFrontContinuousDeploymentPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudfront.GetContinuousDeploymentPolicyOutput
	var stagingDistribution awstypes.Distribution
	var productionDistribution awstypes.Distribution
	resourceName := "aws_cloudfront_continuous_deployment_policy.test"
	stagingDistributionResourceName := "aws_cloudfront_distribution.staging"
	productionDistributionResourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContinuousDeploymentPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContinuousDeploymentPolicyConfig_init(defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, stagingDistributionResourceName, &stagingDistribution),
					testAccCheckDistributionExists(ctx, productionDistributionResourceName, &productionDistribution),
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcloudfront.ResourceContinuousDeploymentPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontContinuousDeploymentPolicy_trafficConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudfront.GetContinuousDeploymentPolicyOutput
	var stagingDistribution awstypes.Distribution
	var productionDistribution awstypes.Distribution
	resourceName := "aws_cloudfront_continuous_deployment_policy.test"
	stagingDistributionResourceName := "aws_cloudfront_distribution.staging"
	productionDistributionResourceName := "aws_cloudfront_distribution.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContinuousDeploymentPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContinuousDeploymentPolicyConfig_init(defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, stagingDistributionResourceName, &stagingDistribution),
					testAccCheckDistributionExists(ctx, productionDistributionResourceName, &productionDistribution),
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
				),
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleWeight(false, "0.01", 300, 600, defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                                                   "SingleWeight",
						"single_weight_config.#":                                         acctest.Ct1,
						"single_weight_config.0.weight":                                  "0.01",
						"single_weight_config.0.session_stickiness_config.#":             acctest.Ct1,
						"single_weight_config.0.session_stickiness_config.0.idle_ttl":    "300",
						"single_weight_config.0.session_stickiness_config.0.maximum_ttl": "600",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleWeight(true, "0.02", 600, 1200, defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                                                   "SingleWeight",
						"single_weight_config.#":                                         acctest.Ct1,
						"single_weight_config.0.weight":                                  "0.02",
						"single_weight_config.0.session_stickiness_config.#":             acctest.Ct1,
						"single_weight_config.0.session_stickiness_config.0.idle_ttl":    "600",
						"single_weight_config.0.session_stickiness_config.0.maximum_ttl": "1200",
					}),
				),
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleHeader(false, "aws-cf-cd-test", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                  "SingleHeader",
						"single_header_config.#":        acctest.Ct1,
						"single_header_config.0.header": "aws-cf-cd-test",
						"single_header_config.0.value":  "test",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleHeader(true, "aws-cf-cd-test2", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                  "SingleHeader",
						"single_header_config.#":        acctest.Ct1,
						"single_header_config.0.header": "aws-cf-cd-test2",
						"single_header_config.0.value":  "test2",
					}),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/33338
func TestAccCloudFrontContinuousDeploymentPolicy_domainChange(t *testing.T) {
	ctx := acctest.Context(t)
	var policy cloudfront.GetContinuousDeploymentPolicyOutput
	var stagingDistribution awstypes.Distribution
	var productionDistribution awstypes.Distribution
	resourceName := "aws_cloudfront_continuous_deployment_policy.test"
	stagingDistributionResourceName := "aws_cloudfront_distribution.staging"
	productionDistributionResourceName := "aws_cloudfront_distribution.test"
	domain1 := fmt.Sprintf("%s.example.com", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))
	domain2 := fmt.Sprintf("%s.example.com", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContinuousDeploymentPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContinuousDeploymentPolicyConfig_init(domain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionExists(ctx, stagingDistributionResourceName, &stagingDistribution),
					testAccCheckDistributionExists(ctx, productionDistributionResourceName, &productionDistribution),
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
				),
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleWeight(true, "0.01", 300, 600, domain1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                                                   "SingleWeight",
						"single_weight_config.#":                                         acctest.Ct1,
						"single_weight_config.0.weight":                                  "0.01",
						"single_weight_config.0.session_stickiness_config.#":             acctest.Ct1,
						"single_weight_config.0.session_stickiness_config.0.idle_ttl":    "300",
						"single_weight_config.0.session_stickiness_config.0.maximum_ttl": "600",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(stagingDistributionResourceName, "origin.*", map[string]string{
						names.AttrDomainName: domain1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(productionDistributionResourceName, "origin.*", map[string]string{
						names.AttrDomainName: domain1,
					}),
				),
			},
			{
				Config: testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleWeight(true, "0.01", 300, 600, domain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContinuousDeploymentPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "traffic_config.*", map[string]string{
						names.AttrType:                                                   "SingleWeight",
						"single_weight_config.#":                                         acctest.Ct1,
						"single_weight_config.0.weight":                                  "0.01",
						"single_weight_config.0.session_stickiness_config.#":             acctest.Ct1,
						"single_weight_config.0.session_stickiness_config.0.idle_ttl":    "300",
						"single_weight_config.0.session_stickiness_config.0.maximum_ttl": "600",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(stagingDistributionResourceName, "origin.*", map[string]string{
						names.AttrDomainName: domain2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(productionDistributionResourceName, "origin.*", map[string]string{
						names.AttrDomainName: domain2,
					}),
				),
			},
		},
	})
}

func testAccCheckContinuousDeploymentPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_continuous_deployment_policy" {
				continue
			}

			_, err := tfcloudfront.FindContinuousDeploymentPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Continuous Deployment Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckContinuousDeploymentPolicyExists(ctx context.Context, n string, v *cloudfront.GetContinuousDeploymentPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindContinuousDeploymentPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccContinuousDeploymentPolicyConfigBase_staging(domain string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "staging" {
  enabled          = true
  retain_on_delete = false
  staging          = true

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = %[1]q
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, domain)
}

// The initial production distribution must be created _without_ the continuous
// deployment policy attached. Example error:
//
// InvalidArgument: Continuous deployment policy is not supported during distribution creation.
func testAccContinuousDeploymentPolicyConfigBase_productionInit(domain string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = true
  retain_on_delete = false

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = %[1]q
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, domain)
}

func testAccContinuousDeploymentPolicyConfigBase_production(domain string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_distribution" "test" {
  enabled          = true
  retain_on_delete = false

  continuous_deployment_policy_id = aws_cloudfront_continuous_deployment_policy.test.id

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"

    forwarded_values {
      query_string = false

      cookies {
        forward = "all"
      }
    }
  }

  origin {
    domain_name = %[1]q
    origin_id   = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}
`, domain)
}

// testAccContinuousDeploymentPolicyConfig_init initializes the staging and production
// distributions and creates the continuous deployment policy, but does not yet
// associate it with the production distribution. Association with a production distribution
// must be done in sunsequent steps, so this will always be the first configuration deployed.
//
// A fortunate side effect of the initial deployment without a production distribution
// association is that it allows the the policy to be "disappeared" out of band safely,
// avoiding errors like the following:
//
// ContinuousDeploymentPolicyInUse: The specified continuous deployment policy is
// currently associated with a distribution.
func testAccContinuousDeploymentPolicyConfig_init(domain string) string {
	return acctest.ConfigCompose(
		testAccContinuousDeploymentPolicyConfigBase_staging(domain),
		testAccContinuousDeploymentPolicyConfigBase_productionInit(domain),
		`
resource "aws_cloudfront_continuous_deployment_policy" "test" {
  enabled = false

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleWeight"
    single_weight_config {
      weight = "0.01"
    }
  }
}
`)
}

func testAccContinuousDeploymentPolicyConfig_basic() string {
	return acctest.ConfigCompose(
		testAccContinuousDeploymentPolicyConfigBase_staging(defaultDomain),
		testAccContinuousDeploymentPolicyConfigBase_production(defaultDomain),
		`
resource "aws_cloudfront_continuous_deployment_policy" "test" {
  enabled = false

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleWeight"
    single_weight_config {
      weight = "0.01"
    }
  }
}
`)
}

func testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleWeight(enabled bool, weight string, idleTTL, maxTTL int, domain string) string {
	return acctest.ConfigCompose(
		testAccContinuousDeploymentPolicyConfigBase_staging(domain),
		testAccContinuousDeploymentPolicyConfigBase_production(domain),
		fmt.Sprintf(`
resource "aws_cloudfront_continuous_deployment_policy" "test" {
  enabled = %[1]t

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleWeight"
    single_weight_config {
      weight = %[2]q
      session_stickiness_config {
        idle_ttl    = %[3]d
        maximum_ttl = %[4]d
      }
    }
  }
}
`, enabled, weight, idleTTL, maxTTL))
}

func testAccContinuousDeploymentPolicyConfig_TrafficConfig_singleHeader(enabled bool, header, value string) string {
	return acctest.ConfigCompose(
		testAccContinuousDeploymentPolicyConfigBase_staging(defaultDomain),
		testAccContinuousDeploymentPolicyConfigBase_production(defaultDomain),
		fmt.Sprintf(`
resource "aws_cloudfront_continuous_deployment_policy" "test" {
  enabled = %[1]t

  staging_distribution_dns_names {
    items    = [aws_cloudfront_distribution.staging.domain_name]
    quantity = 1
  }

  traffic_config {
    type = "SingleHeader"
    single_header_config {
      header = %[2]q
      value  = %[3]q
    }
  }
}
`, enabled, header, value))
}
