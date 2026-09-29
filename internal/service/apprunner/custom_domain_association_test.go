// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerCustomDomainAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"
	serviceResourceName := "aws_apprunner_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "dns_target", regexache.MustCompile(fmt.Sprintf(`^[0-9a-z]+\.%s\.awsapprunner\.com$`, acctest.Region()))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", serviceResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "pending_certificate_dns_validation"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(append(
						domainValidationRecords(domain),
						wwwValidationRecord(domain),
					))),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapprunner.ResourceCustomDomainAssociation(), resourceName),
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

func TestAccAppRunnerCustomDomainAssociation_disappears_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"
	resource2Name := "aws_apprunner_custom_domain_association.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_multiple(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapprunner.ResourceCustomDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resource2Name, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(resource2Name, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_WWWSubdomain_false(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain(rName, domain, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(
						domainValidationRecords(domain),
					)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_WWWSubdomain_true(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain(rName, domain, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtTrue),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(append(
						domainValidationRecords(domain),
						wwwValidationRecord(domain),
					))),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_DomainName_Wildcard_WWWSubdomain_default(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domainName := acctest.NewDomainName(root).RandomSubdomain(t)
	wildcard := domainName.Subdomain("*").String()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomDomainAssociationConfig_basic(rName, wildcard),
				ExpectError: regexache.MustCompile(`enable_www_subdomain must be false for wildcard domains`),
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_DomainName_Wildcard_WWWSubdomain_false(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domainName := acctest.NewDomainName(root).RandomSubdomain(t)
	wildcard := domainName.Subdomain("*").String()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain(rName, wildcard, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcard),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(
						domainValidationRecords(domainName.String()),
					)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_DomainName_Wildcard_WWWSubdomain_true(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domainName := acctest.NewDomainName(root).RandomSubdomain(t)
	wildcard := domainName.Subdomain("*").String()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCustomDomainAssociationConfig_wwwSubdomain(rName, wildcard, true),
				ExpectError: regexache.MustCompile(`enable_www_subdomain must be false for wildcard domains`),
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_Route53Records_WWWSubdomain_false(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain_route53Records(rName, root, domain, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(
						domainValidationRecords(domain),
					)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_Route53Records_WWWSubdomain_true(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain_route53Records(rName, root, domain, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtTrue),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(append(
						domainValidationRecords(domain),
						wwwValidationRecord(domain),
					))),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_Route53Records_Wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domainName := acctest.NewDomainName(root).RandomSubdomain(t)
	wildcard := domainName.Subdomain("*").String()
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_wwwSubdomain_route53Records(rName, root, wildcard, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcard),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_validation_records"), knownvalue.SetExact(
						domainValidationRecords(domainName.String()),
					)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	root := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	domain := acctest.RandomSubdomainForRoot(t, root)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"
	resource2Name := "aws_apprunner_custom_domain_association.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_multiple(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					testAccCheckCustomDomainAssociationExists(ctx, t, resource2Name),
					resource.TestCheckResourceAttrPair(resourceName, "dns_target", resource2Name, "dns_target"),
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

func testAccCheckCustomDomainAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_custom_domain_association" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

			_, err := tfapprunner.FindCustomDomainByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["service_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner Custom Domain Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomDomainAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

		_, err := tfapprunner.FindCustomDomainByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["service_arn"])

		return err
	}
}

func domainValidationRecords(domainName string) []knownvalue.Check {
	return []knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf(`^_[0-9a-f]{32}\.%s\.$`, regexp.QuoteMeta(domainName)))),
			names.AttrStatus: tfknownvalue.StringExact(awstypes.CertificateValidationRecordStatusPendingValidation),
			names.AttrType:   knownvalue.StringExact("CNAME"),
			names.AttrValue:  knownvalue.StringRegexp(regexache.MustCompile(`^_[0-9a-f]{32}\.[a-z]+\.acm-validations\.aws\.$`)),
		}),
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf(`^_[0-9a-f]{32}\.[0-9a-z]{31}\.%s\.$`, regexp.QuoteMeta(domainName)))),
			names.AttrStatus: tfknownvalue.StringExact(awstypes.CertificateValidationRecordStatusPendingValidation),
			names.AttrType:   knownvalue.StringExact("CNAME"),
			names.AttrValue:  knownvalue.StringRegexp(regexache.MustCompile(`^_[0-9a-f]{32}\.[a-z]+\.acm-validations\.aws\.$`)),
		}),
	}
}

func wwwValidationRecord(domainName string) knownvalue.Check {
	return knownvalue.ObjectExact(map[string]knownvalue.Check{
		names.AttrName:   knownvalue.StringRegexp(regexache.MustCompile(fmt.Sprintf(`^_[0-9a-f]{32}\.www\.%s\.$`, regexp.QuoteMeta(domainName)))),
		names.AttrStatus: tfknownvalue.StringExact(awstypes.CertificateValidationRecordStatusPendingValidation),
		names.AttrType:   knownvalue.StringExact("CNAME"),
		names.AttrValue:  knownvalue.StringRegexp(regexache.MustCompile(`^_[0-9a-f]{32}\.[a-z]+\.acm-validations\.aws\.$`)),
	})
}

func testAccCustomDomainAssociationConfig_basic(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_custom_domain_association" "test" {
  domain_name = %[2]q
  service_arn = aws_apprunner_service.test.arn
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName, domain)
}

func testAccCustomDomainAssociationConfig_wwwSubdomain(rName, domain string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_apprunner_custom_domain_association" "test" {
  domain_name = %[2]q
  service_arn = aws_apprunner_service.test.arn

  enable_www_subdomain = %[3]t
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName, domain, enabled)
}

func testAccCustomDomainAssociationConfig_wwwSubdomain_route53Records(rName, root, domain string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccCustomDomainAssociationConfig_wwwSubdomain(rName, domain, enabled),
		fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name = %[1]q
}

resource "aws_route53_record" "dns_target" {
  zone_id = data.aws_route53_zone.test.zone_id
  name    = aws_apprunner_custom_domain_association.test.domain_name
  type    = "CNAME"
  ttl     = 300

  records = [aws_apprunner_custom_domain_association.test.dns_target]
}

resource "aws_route53_record" "www" {
  count = aws_apprunner_custom_domain_association.test.enable_www_subdomain ? 1 : 0

  zone_id = data.aws_route53_zone.test.zone_id
  name    = "www.${aws_apprunner_custom_domain_association.test.domain_name}"
  type    = "CNAME"
  ttl     = 300

  records = [aws_apprunner_custom_domain_association.test.dns_target]
}

locals {
  certificate_validation_records = tolist(aws_apprunner_custom_domain_association.test.certificate_validation_records)
}

resource "aws_route53_record" "validation" {
  count = aws_apprunner_custom_domain_association.test.enable_www_subdomain ? 3 : 2

  zone_id = data.aws_route53_zone.test.zone_id
  name    = local.certificate_validation_records[count.index].name
  type    = "CNAME"
  ttl     = 300

  records = [local.certificate_validation_records[count.index].value]
}
`, root))
}

func testAccCustomDomainAssociationConfig_multiple(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_custom_domain_association" "test" {
  domain_name = %[2]q
  service_arn = aws_apprunner_service.test.arn
}

resource "aws_apprunner_custom_domain_association" "test2" {
  domain_name = "other-%[2]s"
  service_arn = aws_apprunner_service.test.arn
}

resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}
`, rName, domain)
}
