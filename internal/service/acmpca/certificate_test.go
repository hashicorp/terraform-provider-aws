// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfacmpca "github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMPCACertificate_rootCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_root(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificateChain, ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA512WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/RootCACertificate/V1"),
					resource.TestCheckNoResourceAttr(resourceName, "api_passthrough"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccACMPCACertificate_rootCertificateWithAPIPassthrough(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_rootWithAPIPassthrough(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					testAccCheckCertificateExtension(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificateChain, ""),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA512WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/RootCACertificate_APIPassthrough/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccACMPCACertificate_subordinateCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	rootCertificateAuthorityResourceName := "aws_acmpca_certificate_authority.root"
	subordinateCertificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subordinate(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", rootCertificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_signing_request", subordinateCertificateAuthorityResourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "YEARS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA512WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/SubordinateCACertificate_PathLen0/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccACMPCACertificate_endEntityCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(t, 4096, csrDomain)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_endEntity(domain, acctest.TLSPEMEscapeNewlines(csr)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccACMPCACertificate_Validity_endDate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(t, 4096, csrDomain)
	domain := acctest.RandomDomainName()
	later := time.Now().Add(time.Minute * 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_validityEndDate(domain, acctest.TLSPEMEscapeNewlines(csr), later),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", later),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "END_DATE"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func TestAccACMPCACertificate_Validity_absolute(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	csrDomain := acctest.RandomDomainName()
	csr, _ := acctest.TLSRSAX509CertificateRequestPEM(t, 4096, csrDomain)
	domain := acctest.RandomDomainName()
	later := time.Now().Add(time.Minute * 10).Unix()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_validityAbsolute(domain, acctest.TLSPEMEscapeNewlines(csr), later),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+/certificate/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttr(resourceName, "certificate_signing_request", csr),
					resource.TestCheckResourceAttr(resourceName, "validity.0.value", strconv.FormatInt(later, 10)),
					resource.TestCheckResourceAttr(resourceName, "validity.0.type", "ABSOLUTE"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm", "SHA256WITHRSA"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "template_arn", "acm-pca", "template/EndEntityCertificate/V1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"api_passthrough",
					"certificate_signing_request",
					"signing_algorithm",
					"template_arn",
					"validity",
				},
			},
		},
	})
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acmpca_certificate" {
				continue
			}

			_, err := tfacmpca.FindCertificateByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["certificate_authority_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if errs.IsAErrorMessageContains[*types.InvalidStateException](err, "not in the correct state to have issued certificates") {
				// This is returned when checking root certificates and the certificate has not been associated with the certificate authority
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ACM PCA Certificate %s still exists", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		_, err := tfacmpca.FindCertificateByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["certificate_authority_arn"])

		return err
	}
}

func testAccCheckCertificateExtension(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		block, _ := pem.Decode([]byte(rs.Primary.Attributes[names.AttrCertificate]))
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("Failed to parse certificate: %w", err)
		}

		if len(cert.PermittedDNSDomains) != 1 {
			return fmt.Errorf("Permitted DNS Domains expected to have 1 element, got %d", len(cert.PermittedDNSDomains))
		}

		expectedPermittedDNSDomain := ".permitted.test"
		if cert.PermittedDNSDomains[0] != expectedPermittedDNSDomain {
			return fmt.Errorf("Expected permitted DNS domain: %s, got: %s", expectedPermittedDNSDomain, cert.PermittedDNSDomains[0])
		}

		return nil
	}
}

func testAccCertificateConfig_root(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_partition" "current" {}
`, domain)
}

func testAccCertificateConfig_rootWithAPIPassthrough(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate_APIPassthrough/V1"

  validity {
    type  = "YEARS"
    value = 1
  }

  api_passthrough = jsonencode({
    Extensions = {
      CustomExtensions = [
        {
          ObjectIdentifier = "2.5.29.30",
          Value            = "MBWgEzARgg8ucGVybWl0dGVkLnRlc3Q=",
          Critical         = true
        },
      ]
    }
  })
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_partition" "current" {}
`, domain)
}

func testAccCertificateConfig_subordinate(domain string) string {
	return acctest.ConfigCompose(
		testAccCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/SubordinateCACertificate_PathLen0/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "SUBORDINATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "sub.%[1]s"
    }
  }
}
`, domain))
}

func testAccCertificateConfig_endEntity(domain, csr string) string {
	return acctest.ConfigCompose(
		testAccCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "DAYS"
    value = 1
  }
}
`, csr))
}

func testAccCertificateConfig_validityEndDate(domain, csr, expiry string) string {
	return acctest.ConfigCompose(
		testAccCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "END_DATE"
    value = %[2]q
  }
}
`, csr, expiry))
}

func testAccCertificateConfig_validityAbsolute(domain, csr string, expiry int64) string {
	return acctest.ConfigCompose(
		testAccCertificateBaseRootCAConfig(domain),
		fmt.Sprintf(`
resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = "%[1]s"
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/EndEntityCertificate/V1"

  validity {
    type  = "ABSOLUTE"
    value = %[2]d
  }
}
`, csr, expiry))
}

func testAccCertificateBaseRootCAConfig(domain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "root" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority_certificate" "root" {
  certificate_authority_arn = aws_acmpca_certificate_authority.root.arn

  certificate       = aws_acmpca_certificate.root.certificate
  certificate_chain = aws_acmpca_certificate.root.certificate_chain
}

resource "aws_acmpca_certificate" "root" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.root.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

data "aws_partition" "current" {}
  `, domain)
}

func TestValidateTemplateARN(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"arn:aws:acm-pca:::template/EndEntityCertificate/V1",                     // lintignore:AWSAT005
		"arn:aws:acm-pca:::template/SubordinateCACertificate_PathLen0/V1",        // lintignore:AWSAT005
		"arn:aws-us-gov:acm-pca:::template/EndEntityCertificate/V1",              // lintignore:AWSAT005
		"arn:aws-us-gov:acm-pca:::template/SubordinateCACertificate_PathLen0/V1", // lintignore:AWSAT005
	}
	for _, v := range validNames {
		_, errors := tfacmpca.ValidTemplateARN(v, "template_arn")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid ACM PCA ARN: %q", v, errors)
		}
	}

	invalidNames := []string{
		names.AttrARN,
		"arn:aws:s3:::my_corporate_bucket/exampleobject.png",                       // lintignore:AWSAT005
		"arn:aws:acm-pca:us-west-2::template/SubordinateCACertificate_PathLen0/V1", // lintignore:AWSAT003,AWSAT005
		"arn:aws:acm-pca::123456789012:template/EndEntityCertificate/V1",           // lintignore:AWSAT005
		"arn:aws:acm-pca:::not-a-template/SubordinateCACertificate_PathLen0/V1",    // lintignore:AWSAT005
	}
	for _, v := range invalidNames {
		_, errors := tfacmpca.ValidTemplateARN(v, "template_arn")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid ARN", v)
		}
	}
}

func TestExpandValidityValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Type     string
		Value    string
		Expected int64
	}{
		{
			Type:     string(types.ValidityPeriodTypeEndDate),
			Value:    "2021-02-26T16:04:00Z",
			Expected: 20210226160400,
		},
		{
			Type:     string(types.ValidityPeriodTypeEndDate),
			Value:    "2021-02-26T16:04:00-08:00",
			Expected: 20210227000400,
		},
		{
			Type:     string(types.ValidityPeriodTypeAbsolute),
			Value:    "1614385420",
			Expected: 1614385420,
		},
		{
			Type:     string(types.ValidityPeriodTypeYears),
			Value:    acctest.Ct2,
			Expected: 2,
		},
	}

	for _, testcase := range testCases {
		i, _ := tfacmpca.ExpandValidityValue(testcase.Type, testcase.Value)
		if i != testcase.Expected {
			t.Errorf("%s, %q: expected %d, got %d", testcase.Type, testcase.Value, testcase.Expected, i)
		}
	}
}
