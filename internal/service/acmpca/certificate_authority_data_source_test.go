// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMPCACertificateAuthorityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateAuthorityDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccCertificateAuthorityDataSourceConfig_arn(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCertificate, resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCertificateChain, resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_signing_request", resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_storage_security_standard", resourceName, "key_storage_security_standard"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.#", resourceName, "revocation_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.#", resourceName, "revocation_configuration.0.crl_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.enabled", resourceName, "revocation_configuration.0.crl_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "serial", resourceName, "serial"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(datasourceName, "usage_mode", resourceName, "usage_mode"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthorityDataSource_s3ObjectACL(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateAuthorityDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccCertificateAuthorityDataSourceConfig_s3ObjectACLARN(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCertificate, resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCertificateChain, resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_signing_request", resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_storage_security_standard", resourceName, "key_storage_security_standard"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.#", resourceName, "revocation_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.#", resourceName, "revocation_configuration.0.crl_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.enabled", resourceName, "revocation_configuration.0.crl_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl"),
					resource.TestCheckResourceAttrPair(datasourceName, "serial", resourceName, "serial"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(datasourceName, "usage_mode", resourceName, "usage_mode"),
				),
			},
		},
	})
}

func testAccCertificateAuthorityDataSourceConfig_arn(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "wrong" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = aws_acmpca_certificate_authority.test.arn
}
`, commonName)
}

func testAccCertificateAuthorityDataSourceConfig_s3ObjectACLARN(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "wrong" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = aws_acmpca_certificate_authority.test.arn
}
`, commonName)
}

// lintignore:AWSAT003,AWSAT005
const testAccCertificateAuthorityDataSourceConfig_nonExistent = `
data "aws_acmpca_certificate_authority" "test" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/tf-acc-test-does-not-exist"
}
`
