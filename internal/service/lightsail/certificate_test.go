// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLightsailCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "lightsail", regexache.MustCompile(`Certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
					// When using a .test domain, Domain Validation Records are not returned
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_subjectAlternativeNames(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())
	subjectAlternativeName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", subjectAlternativeName),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domainName),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_DomainValidationOptions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	// Lightsail will only return Domain Validation Options when using a non-test domain.
	// We need to provide a non-test domain in order to test these values.
	domainName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)
	subjectAlternativeName := fmt.Sprintf("%s.com", acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rName, domainName, subjectAlternativeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domainName,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   subjectAlternativeName,
						"resource_record_type": "CNAME",
					}),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_tags2(rName, domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_keyOnlyTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, acctest.CtKey1, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_tags2(rName, domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(rName, domainName, acctest.CtKey2, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, ""),
				),
			},
		},
	})
}

func TestAccLightsailCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lightsail_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.ACMCertificateRandomSubDomain(acctest.RandomDomainName())

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the Certificate
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)
		_, err := conn.DeleteCertificate(ctx, &lightsail.DeleteCertificateInput{
			CertificateName: aws.String(rName),
		})

		if err != nil {
			return fmt.Errorf("error deleting Lightsail Certificate in disappear test")
		}

		// sleep 7 seconds to give it time, so we don't have to poll
		time.Sleep(7 * time.Second)

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lightsail_certificate" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

			_, err := tflightsail.FindCertificateById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Lightsail, create.ErrActionCheckingDestroyed, tflightsail.ResCertificate, rs.Primary.ID, errors.New("still exists"))
		}

		return nil
	}
}

func testAccCheckCertificateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		respCertificate, err := tflightsail.FindCertificateById(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if respCertificate == nil {
			return fmt.Errorf("Certificate %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCertificateConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
}
`, rName, domainName)
}

func testAccCertificateConfig_subjectAlternativeNames(rName, domainName, san string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name                      = %[1]q
  domain_name               = %[2]q
  subject_alternative_names = [%[3]q]
}
`, rName, domainName, san)
}

func testAccCertificateConfig_tags1(resourceName, domainName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
  tags = {
    %[3]q = %[4]q
  }
}
`, resourceName, domainName, tagKey1, tagValue1)
}

func testAccCertificateConfig_tags2(resourceName, domainName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_certificate" "test" {
  name        = %[1]q
  domain_name = %[2]q
  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, resourceName, domainName, tagKey1, tagValue1, tagKey2, tagValue2)
}
