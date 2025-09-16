// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMPCACertificateAuthoritiesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	domain1 := acctest.RandomDomainName()
	domain2 := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthoritiesDataSourceConfig_arn(domain1, domain2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair("data.aws_acmpca_certificate_authorities.test", "arns.*", "aws_acmpca_certificate_authority.test1", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair("data.aws_acmpca_certificate_authorities.test", "arns.*", "aws_acmpca_certificate_authority.test2", names.AttrARN),
				),
			},
		},
	})
}

func testAccCertificateAuthoritiesDataSourceConfig_arn(domain1 string, domain2 string) string {
	return fmt.Sprintf(`
data "aws_acmpca_certificate_authorities" "test" {
  depends_on = [
	aws_acmpca_certificate_authority.test1,
	aws_acmpca_certificate_authority.test2,
  ]
}

resource "aws_acmpca_certificate_authority" "test1" {
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

resource "aws_acmpca_certificate_authority" "test2" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[2]q
    }
  }
}

`, domain1, domain2)
}
