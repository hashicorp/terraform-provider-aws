// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontDistributionTenant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_cloudfront_distribution_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("cloudfront", regexache.MustCompile(`distribution-tenant/dt_[0-9A-Za-z]+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("connection_group_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("etag"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_cloudfront_distribution_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudfront.ResourceDistributionTenant, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_customCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_cloudfront_distribution_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_customCertificate(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_customCertificateWithWebACL(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_cloudfront_distribution_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_customCertificateWithWebACL(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccCloudFrontDistributionTenant_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var tenant awstypes.DistributionTenant
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_cloudfront_distribution_tenant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDistributionTenantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionTenantConfig_tags1(rName, rootDomain, domain, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(resourceName, names.AttrName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"wait_for_deployment",
					names.AttrStatus,
				},
			},
			{
				Config: testAccDistributionTenantConfig_tags2(rName, rootDomain, domain, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccDistributionTenantConfig_tags1(rName, rootDomain, domain, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionTenantExists(ctx, t, resourceName, &tenant),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckDistributionTenantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudfront_distribution_tenant" {
				continue
			}

			_, err := tfcloudfront.FindDistributionTenantByIdentifier(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFront Distribution Tenant (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDistributionTenantExists(ctx context.Context, t *testing.T, n string, v *awstypes.DistributionTenant) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFrontClient(ctx)

		output, err := tfcloudfront.FindDistributionTenantByIdentifier(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output.DistributionTenant

		return nil
	}
}

func testAccDistributionTenantConfig_baseCertificate(rootDomain, domain string) string {
	return acctest.ConfigCompose(testAccRegionProviderConfig(), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  depends_on = [aws_route53_record.test]

  certificate_arn = aws_acm_certificate.test.arn
}

data "aws_acm_certificate" "test" {
  domain      = %[1]q
  most_recent = true

  depends_on = [aws_acm_certificate_validation.test]
}
`, domain, rootDomain))
}

func testAccDistributionTenantConfig_basic(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_baseCertificate(rootDomain, tenantDomain), fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = true
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "www.example.com"
    id          = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_multitenant_distribution.test.id
  domain {
    domain = %[2]q
  }
  name    = %[1]q
  enabled = false

  parameter {
    name  = "origin_domain"
    value = "www.example.com"
  }
}
`, rName, tenantDomain))
}

func testAccDistributionTenantConfig_customCertificate(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_baseCertificate(rootDomain, tenantDomain), fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = true
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "www.example.com"
    id          = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
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

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_multitenant_distribution.test.id
  domain {
    domain = %[2]q
  }
  name    = %[1]q
  enabled = false

  parameter {
    name  = "origin_domain"
    value = "www.example.com"
  }

  customizations {
    geo_restriction {
      locations        = ["US", "CA"]
      restriction_type = "whitelist"
    }

    certificate {
      arn = data.aws_acm_certificate.test.arn
    }
  }
}
`, rName, tenantDomain))
}

func testAccDistributionTenantConfig_customCertificateWithWebACL(rName, rootDomain, tenantDomain string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_baseCertificate(rootDomain, tenantDomain), fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = true
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "www.example.com"
    id          = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
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

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_multitenant_distribution.test.id
  domain {
    domain = %[2]q
  }
  name    = %[1]q
  enabled = false

  parameter {
    name  = "origin_domain"
    value = "www.example.com"
  }

  customizations {
    geo_restriction {
      locations        = ["US", "CA"]
      restriction_type = "whitelist"
    }

    certificate {
      arn = data.aws_acm_certificate.test.arn
    }

    web_acl {
      action = "override"
      arn    = aws_wafv2_web_acl.test.arn
    }
  }
}

resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = "tftest"
  scope       = "CLOUDFRONT"

  default_action {
    allow {
      custom_request_handling {
        insert_header {
          name  = "X-WebACL-Test"
          value = "test"
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, tenantDomain))
}

func testAccDistributionTenantConfig_tags1(rName, rootDomain, tenantDomain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_baseCertificate(rootDomain, tenantDomain), fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = true
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "www.example.com"
    id          = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_multitenant_distribution.test.id
  domain {
    domain = %[2]q
  }
  name    = %[1]q
  enabled = false

  parameter {
    name  = "origin_domain"
    value = "www.example.com"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, tenantDomain, tagKey1, tagValue1))
}

func testAccDistributionTenantConfig_tags2(rName, rootDomain, tenantDomain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDistributionTenantConfig_baseCertificate(rootDomain, tenantDomain), fmt.Sprintf(`
resource "aws_cloudfront_multitenant_distribution" "test" {
  enabled = true
  comment = "Test multi-tenant distribution"

  origin {
    domain_name = "www.example.com"
    id          = "test"

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "http-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "test"
    viewer_protocol_policy = "allow-all"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # AWS Managed CachingDisabled policy

    allowed_methods {
      items          = ["GET", "HEAD"]
      cached_methods = ["GET", "HEAD"]
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn = data.aws_acm_certificate.test.arn
    ssl_support_method  = "sni-only"
  }

  tenant_config {
    parameter_definition {
      name = "origin_domain"
      definition {
        string_schema {
          required = true
          comment  = "Origin domain parameter for tenants"
        }
      }
    }
  }
}

resource "aws_cloudfront_distribution_tenant" "test" {
  distribution_id = aws_cloudfront_multitenant_distribution.test.id
  domain {
    domain = %[2]q
  }
  name    = %[1]q
  enabled = false

  parameter {
    name  = "origin_domain"
    value = "www.example.com"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, tenantDomain, tagKey1, tagValue1, tagKey2, tagValue2))
}
