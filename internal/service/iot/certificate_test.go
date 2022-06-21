package iot_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIoTCertificate_csr(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_csr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func TestAccIoTCertificate_Keys_certificate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_keys,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func TestAccIoTCertificate_Keys_existingCertificate(t *testing.T) {
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "testcert")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iot.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_existingCertificate(acctest.TLSPEMEscapeNewlines(certificate)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func testAccCheckCertificateDestroy_basic(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_certificate" {
			continue
		}

		// Try to find the Cert
		DescribeCertOpts := &iot.DescribeCertificateInput{
			CertificateId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCertificate(DescribeCertOpts)

		if err == nil {
			if resp.CertificateDescription != nil {
				return fmt.Errorf("Device Certificate still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}

	}

	return nil
}

var testAccCertificateConfig_csr = `
resource "aws_iot_certificate" "foo_cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}
`

var testAccCertificateConfig_keys = `
resource "aws_iot_certificate" "foo_cert" {
  active = true
}
`

func testAccCertificateConfig_existingCertificate(pem string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "foo_cert" {
  active          = true
  certificate_pem = "%[1]s"
}
`, pem)
}
