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

func TestAccACMPCACertificateDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acmpca_certificate.test"
	dataSourceName := "data.aws_acmpca_certificate.test"

	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
			{
				Config: testAccCertificateDataSourceConfig_arn(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCertificate, resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCertificateChain, resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttrPair(dataSourceName, "certificate_authority_arn", resourceName, "certificate_authority_arn"),
				),
			},
		},
	})
}

func testAccCertificateDataSourceConfig_arn(domain string) string {
	return fmt.Sprintf(`
data "aws_acmpca_certificate" "test" {
  arn                       = aws_acmpca_certificate.test.arn
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA256WITHRSA"

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

const testAccCertificateDataSourceConfig_nonExistent = `
data "aws_acmpca_certificate" "test" {
  arn                       = "arn:${data.aws_partition.current.partition}:acm-pca:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:certificate-authority/does-not-exist/certificate/does-not-exist"
  certificate_authority_arn = "arn:${data.aws_partition.current.partition}:acm-pca:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:certificate-authority/does-not-exist"
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}
`
