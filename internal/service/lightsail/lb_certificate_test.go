// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLoadBalancerCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lbName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_basic(rName, lbName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lightsail", regexache.MustCompile(`LoadBalancerTlsCertificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
					// When using a .test domain, Domain Validation Records return a single FAILED entry
					resource.TestCheckResourceAttr(resourceName, "domain_validation_records.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccLoadBalancerCertificate_subjectAlternativeNames(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lbName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())
	subjectAlternativeName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName, lbName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", subjectAlternativeName),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
				),
			},
		},
	})
}

func testAccLoadBalancerCertificate_domainValidationRecords(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lbName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// Lightsail will only return Domain Validation Options when using a non-test domain.
	// We need to provide a non-test domain in order to test these values.
	domainName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)
	subjectAlternativeName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName, lbName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, t, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_records.*", map[string]string{
						names.AttrDomainName:   domainName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_records.*", map[string]string{
						names.AttrDomainName:   subjectAlternativeName,
						"resource_record_type": "CNAME",
					}),
				),
			},
		},
	})
}

func testAccLoadBalancerCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_lb_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	lbName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerCertificateConfig_basic(rName, lbName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerCertificateExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflightsail.ResourceLoadBalancerCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLoadBalancerCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_lb_certificate" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

			_, err := tflightsail.FindLoadBalancerCertificateById(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResLoadBalancerCertificate, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccCheckLoadBalancerCertificateExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LightsailClient(ctx)

		respCertificate, err := tflightsail.FindLoadBalancerCertificateById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if respCertificate == nil {
			return fmt.Errorf("Load Balancer Certificate %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccLoadBalancerCertificateConfigBase(lbName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_lb" "test" {
  name              = %[1]q
  health_check_path = "/"
  instance_port     = "80"
}
`, lbName)
}

func testAccLoadBalancerCertificateConfig_basic(rName string, lbName string, domainName string) string {
	return acctest.ConfigCompose(
		testAccLoadBalancerCertificateConfigBase(lbName),
		fmt.Sprintf(`
resource "aws_lightsail_lb_certificate" "test" {
  name        = %[1]q
  lb_name     = aws_lightsail_lb.test.id
  domain_name = %[2]q
}
`, rName, domainName))
}

func testAccLoadBalancerCertificateConfig_subjectAlternativeNames(rName string, lbName string, domainName string, san string) string {
	return acctest.ConfigCompose(
		testAccLoadBalancerCertificateConfigBase(lbName),
		fmt.Sprintf(`
resource "aws_lightsail_lb_certificate" "test" {
  name                      = %[1]q
  lb_name                   = aws_lightsail_lb.test.id
  domain_name               = %[2]q
  subject_alternative_names = [%[3]q]
}
`, rName, domainName, san))
}
