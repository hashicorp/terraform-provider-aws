// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIoTCACertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	resourceName := "aws_iot_ca_certificate.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, iot.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_basic(caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_auto_registration", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate_pem"),
					resource.TestCheckResourceAttr(resourceName, "certificate_mode", "SNI_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_version"),
					resource.TestCheckResourceAttrSet(resourceName, "generation_id"),
					resource.TestCheckResourceAttr(resourceName, "registration_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "validity.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_after"),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_before"),
				),
			},
		},
	})
}

func testAccCheckCACertificateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		_, err := tfiot.FindCACertificateByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCACertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_ca_certificate" {
				continue
			}

			_, err := tfiot.FindCACertificateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT CA Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCACertificateConfig_basic(caCertificate string) string {
	return fmt.Sprintf(`
resource "aws_iot_ca_certificate" "test" {
  active                  = true
  allow_auto_registration = true
  ca_certificate_pem      = "%[1]s"
  certificate_mode        = "SNI_ONLY"
}
`, acctest.TLSPEMEscapeNewlines(caCertificate))
}
