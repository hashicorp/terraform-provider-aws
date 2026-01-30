// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMCertificateExport_basic(t *testing.T) {
	ctx := acctest.Context(t)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_export.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateExportConfig_basic(certificate, key, "test-passphrase-123"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExportExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, certificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPrivateKey),
				),
			},
		},
	})
}

func TestAccACMCertificateExport_withCertificateChain(t *testing.T) {
	ctx := acctest.Context(t)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_export.test"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateExportDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateExportConfig_withChain(certificate, key, caCertificate, "test-passphrase-456"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExportExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, certificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCertificateChain),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPrivateKey),
				),
			},
		},
	})
}

func testAccCheckCertificateExportDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acm_certificate" {
				continue
			}

			_, err := tfacm.FindCertificateByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ACM Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCertificateExportExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		arn := rs.Primary.Attributes[names.AttrCertificateARN]
		_, err := tfacm.FindCertificateByARN(ctx, conn, arn)

		return err
	}
}

func testAccCertificateExportConfig_basic(certificate, key, passphrase string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = %[1]q
  private_key      = %[2]q

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_export" "test" {
  certificate_arn = aws_acm_certificate.test.arn
  passphrase      = %[3]q
}
`, certificate, key, passphrase)
}

func testAccCertificateExportConfig_withChain(certificate, key, chain, passphrase string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = %[1]q
  private_key       = %[2]q
  certificate_chain = %[3]q

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_export" "test" {
  certificate_arn = aws_acm_certificate.test.arn
  passphrase      = %[4]q
}
`, certificate, key, chain, passphrase)
}
