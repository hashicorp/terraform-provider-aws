// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTCertificate_csr(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_csr,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, acctest.CtCertificatePEM),
					resource.TestCheckResourceAttrSet(resourceName, "csr"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrPrivateKey),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrPublicKey),
				),
			},
		},
	})
}

func TestAccIoTCertificate_Keys_certificate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_keys,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, acctest.CtCertificatePEM),
					resource.TestCheckNoResourceAttr(resourceName, "csr"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPrivateKey),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPublicKey),
				),
			},
		},
	})
}

func TestAccIoTCertificate_Keys_existingCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "testcert")
	resourceName := "aws_iot_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_existingCertificate(certificate, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, acctest.CtCertificatePEM),
					resource.TestCheckNoResourceAttr(resourceName, "csr"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrPrivateKey),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrPublicKey),
				),
			},
			{
				Config: testAccCertificateConfig_existingCertificate(certificate, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckCertificateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindCertificateByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_certificate" {
				continue
			}

			_, err := tfiot.FindCertificateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccCertificateConfig_csr = `
resource "aws_iot_certificate" "test" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}
`

var testAccCertificateConfig_keys = `
resource "aws_iot_certificate" "test" {
  active = true
}
`

func testAccCertificateConfig_existingCertificate(pem string, active bool) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "test" {
  active          = %[2]t
  certificate_pem = "%[1]s"
}
`, acctest.TLSPEMEscapeNewlines(pem), active)
}
