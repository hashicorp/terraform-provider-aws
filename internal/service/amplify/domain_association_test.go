// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamplify "github.com/hashicorp/terraform-provider-aws/internal/service/amplify"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// To test this resource, use a domain that is hosted in Route 53 in the same account and set the environment
// variable "AMPLIFY_DOMAIN_NAME" to the domain name.

func testAccDomainAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_settings.0.certificate_verification_dns_record"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.custom_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "AMPLIFY_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccDomainAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamplify.ResourceDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomainAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var (
		app    types.App
		domain types.DomainAssociation
	)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_basic(rName, domainName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, "aws_amplify_app.test", &app),
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
						"verified":       acctest.CtTrue,
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
			{
				PreConfig: domainAssociationStatusAvailablePreConfig(ctx, t, &app, &domain),
				Config:    testAccDomainAssociationConfig_updated(rName, domainName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
						"verified":       acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    fmt.Sprintf("%s-2", rName),
						names.AttrPrefix: "www",
						// "verified":       acctest.CtTrue, // Even though we're waiting for verification, this isn't getting verified
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccDomainAssociation_certificateSettings_Managed(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var (
		app    types.App
		domain types.DomainAssociation
	)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_certificateSettings_Managed(rName, domainName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, "aws_amplify_app.test", &app),
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "AMPLIFY_MANAGED"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_settings.0.certificate_verification_dns_record"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.custom_certificate_arn", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
			{
				PreConfig: domainAssociationStatusAvailablePreConfig(ctx, t, &app, &domain),
				Config:    testAccDomainAssociationConfig_certificateSettings_Custom(rName, domainName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.certificate_verification_dns_record", ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_settings.0.custom_certificate_arn", "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "CUSTOM"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccDomainAssociation_certificateSettings_Custom(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var (
		app    types.App
		domain types.DomainAssociation
	)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(t)
			acctest.PreCheckAlternateRegion(t, endpoints.UsEast1RegionID) // ACM certificate must be created in us-east-1
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_certificateSettings_Custom(rName, domainName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, "aws_amplify_app.test", &app),
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.certificate_verification_dns_record", ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_settings.0.custom_certificate_arn", "aws_acm_certificate.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "CUSTOM"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
			{
				PreConfig: domainAssociationStatusAvailablePreConfig(ctx, t, &app, &domain),
				Config:    testAccDomainAssociationConfig_certificateSettings_Managed(rName, domainName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "AMPLIFY_MANAGED"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_settings.0.certificate_verification_dns_record"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.custom_certificate_arn", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccDomainAssociation_CreateWithSubdomain(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AMPLIFY_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var domain types.DomainAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_amplify_domain_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AmplifyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAssociationConfig_WithSubdomain(rName, domainName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainAssociationExists(ctx, resourceName, &domain),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "amplify", regexache.MustCompile(fmt.Sprintf(`apps/.+/domains/%s$`, domainName))),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_settings.0.certificate_verification_dns_record"),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.custom_certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "certificate_settings.0.type", "AMPLIFY_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_sub_domain", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "sub_domain.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "",
						"verified":       acctest.CtTrue, // Even though we're waiting for verification, this isn't getting verified
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sub_domain.*", map[string]string{
						"branch_name":    rName,
						names.AttrPrefix: "www",
						// "verified":       acctest.CtTrue, // Even though we're waiting for verification, this isn't getting verified
					}),
					resource.TestCheckResourceAttr(resourceName, "wait_for_verification", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_verification"},
			},
		},
	})
}

func testAccCheckDomainAssociationExists(ctx context.Context, n string, v *types.DomainAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		output, err := tfamplify.FindDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes[names.AttrDomainName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_amplify_domain_association" {
				continue
			}

			_, err := tfamplify.FindDomainAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_id"], rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amplify Domain Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDomainAssociationConfig_basic(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}

func testAccDomainAssociationConfig_updated(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  sub_domain {
    branch_name = aws_amplify_branch.test2.branch_name
    prefix      = "www"
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}

func testAccDomainAssociationConfig_certificateSettings_Managed(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  certificate_settings {
    type = "AMPLIFY_MANAGED"
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}
`, rName, domainName, enableAutoSubDomain, waitForVerification)
}

func testAccDomainAssociationConfig_certificateSettings_Custom(rName, domainName string, enableAutoSubDomain bool, waitForVerification bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  certificate_settings {
    type                   = "CUSTOM"
    custom_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  }

  enable_auto_sub_domain = %[3]t
  wait_for_verification  = %[4]t
}

resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

data "aws_route53_zone" "test" {
  provider = "awsalternate"

  name         = %[2]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  provider = "awsalternate"

  domain_name       = %[2]q
  validation_method = "DNS"
}

resource "aws_route53_record" "test" {
  provider = "awsalternate"

  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  provider = "awsalternate"

  depends_on      = [aws_route53_record.test]
  certificate_arn = aws_acm_certificate.test.arn
}
  `, rName, domainName, enableAutoSubDomain, waitForVerification))
}

func testAccDomainAssociationConfig_WithSubdomain(rName, domainName string, waitForVerification bool) string {
	return fmt.Sprintf(`
resource "aws_amplify_domain_association" "test" {
  app_id      = aws_amplify_app.test.id
  domain_name = %[2]q

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = ""
  }

  sub_domain {
    branch_name = aws_amplify_branch.test.branch_name
    prefix      = "www"
  }

  wait_for_verification = %[3]t
}

resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}
`, rName, domainName, waitForVerification)
}

// In practice, we don't seem to need to wait for the Domain Association to be `AVAILABLE` for the purposes of deploying infrastructure.
// Since subsequent modifications to a Domain Association cannot occur until it is `AVAILABLE`, wait during tests.
func domainAssociationStatusAvailablePreConfig(ctx context.Context, t *testing.T, app *types.App, domain *types.DomainAssociation) func() {
	return func() {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyClient(ctx)

		_, err := tfamplify.WaitDomainAssociationAvailable(ctx, conn, aws.ToString(app.AppId), aws.ToString(domain.DomainName))
		if err != nil {
			t.Fatalf("waiting for Amplify Domain Association (%s/%s) to be available: %s", aws.ToString(app.AppId), aws.ToString(domain.DomainName), err)
		}
	}
}
