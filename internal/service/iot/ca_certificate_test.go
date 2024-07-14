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

func TestAccIoTCACertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	resourceName := "aws_iot_ca_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_basic(caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_auto_registration", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate_pem"),
					resource.TestCheckResourceAttr(resourceName, "certificate_mode", "SNI_ONLY"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_version"),
					resource.TestCheckResourceAttrSet(resourceName, "generation_id"),
					resource.TestCheckResourceAttr(resourceName, "registration_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validity.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_after"),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_before"),
				),
			},
		},
	})
}

func TestAccIoTCACertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	resourceName := "aws_iot_ca_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_basic(caCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceCACertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTCACertificate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	resourceName := "aws_iot_ca_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_tags1(caCertificate, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccCACertificateConfig_tags2(caCertificate, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCACertificateConfig_tags1(caCertificate, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTCACertificate_defaultMode(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_ca_certificate.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_defaultMode(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "allow_auto_registration", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "ca_certificate_pem"),
					resource.TestCheckResourceAttr(resourceName, "certificate_mode", "DEFAULT"),
					resource.TestCheckResourceAttrSet(resourceName, "customer_version"),
					resource.TestCheckResourceAttrSet(resourceName, "generation_id"),
					resource.TestCheckResourceAttr(resourceName, "registration_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validity.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_after"),
					resource.TestCheckResourceAttrSet(resourceName, "validity.0.not_before"),
				),
			},
			{
				Config: testAccCACertificateConfig_defaultMode(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "allow_auto_registration", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccIoTCACertificate_registrationConfig(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iot_ca_certificate.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.4",
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckCACertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCACertificateConfig_registrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCACertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "registration_config.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "registration_config.0.role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_config.0.template_body"),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindCACertificateByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckCACertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

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

func testAccCACertificateConfig_tags1(caCertificate, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_ca_certificate" "test" {
  active                  = true
  allow_auto_registration = true
  ca_certificate_pem      = "%[1]s"
  certificate_mode        = "SNI_ONLY"

  tags = {
    %[2]q = %[3]q
  }
}
`, acctest.TLSPEMEscapeNewlines(caCertificate), tagKey1, tagValue1)
}

func testAccCACertificateConfig_tags2(caCertificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_ca_certificate" "test" {
  active                  = true
  allow_auto_registration = true
  ca_certificate_pem      = "%[1]s"
  certificate_mode        = "SNI_ONLY"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, acctest.TLSPEMEscapeNewlines(caCertificate), tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCACertificateConfig_defaultMode(active, allowAutoRegistration bool) string {
	return fmt.Sprintf(`
resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem
  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
  is_ca_certificate = true
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  private_key_pem = tls_private_key.verification.private_key_pem
  subject {
    common_name = data.aws_iot_registration_code.test.registration_code
  }
}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_locally_signed_cert" "verification" {
  cert_request_pem      = tls_cert_request.verification.cert_request_pem
  ca_private_key_pem    = tls_private_key.ca.private_key_pem
  ca_cert_pem           = tls_self_signed_cert.ca.cert_pem
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_iot_ca_certificate" "test" {
  active                       = %[1]t
  allow_auto_registration      = %[2]t
  ca_certificate_pem           = tls_self_signed_cert.ca.cert_pem
  certificate_mode             = "DEFAULT"
  verification_certificate_pem = tls_locally_signed_cert.verification.cert_pem
}

data "aws_iot_registration_code" "test" {}
`, active, allowAutoRegistration)
}

func testAccCACertificateConfig_registrationConfig_iamRole() string {
	return `
resource "aws_iam_role" "test" {
  name = "test_iot_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "iot.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}
`
}

func testAccCACertificateConfig_registrationConfig() string {
	return acctest.ConfigCompose(testAccCACertificateConfig_registrationConfig_iamRole(), `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "tls_self_signed_cert" "ca" {
  private_key_pem = tls_private_key.ca.private_key_pem
  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
  is_ca_certificate = true
}

resource "tls_private_key" "ca" {
  algorithm = "RSA"
}

resource "tls_cert_request" "verification" {
  private_key_pem = tls_private_key.verification.private_key_pem
  subject {
    common_name = data.aws_iot_registration_code.test.registration_code
  }
}

resource "tls_private_key" "verification" {
  algorithm = "RSA"
}

resource "tls_locally_signed_cert" "verification" {
  cert_request_pem      = tls_cert_request.verification.cert_request_pem
  ca_private_key_pem    = tls_private_key.ca.private_key_pem
  ca_cert_pem           = tls_self_signed_cert.ca.cert_pem
  validity_period_hours = 12
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_iot_ca_certificate" "test" {
  active                       = true
  allow_auto_registration      = true
  ca_certificate_pem           = tls_self_signed_cert.ca.cert_pem
  certificate_mode             = "DEFAULT"
  verification_certificate_pem = tls_locally_signed_cert.verification.cert_pem
  registration_config {
    role_arn      = aws_iam_role.test.arn
    template_body = <<EOF
{
  "Parameters": {
    "AWS::IoT::Certificate::CommonName": {
      "Type": "String"
    },
    "AWS::IoT::Certificate::SerialNumber": {
      "Type": "String"
    },
    "AWS::IoT::Certificate::Country": {
      "Type": "String"
    },
    "AWS::IoT::Certificate::Id": {
      "Type": "String"
    }
  },
  "Resources": {
    "thing": {
      "Type":"AWS::IoT::Thing",
      "Properties": {
        "ThingName": {
          "Ref": "AWS::IoT::Certificate::CommonName"
        },
        "AttributePayload": {
          "version":"v1",
          "serialNumber": {
            "Ref": "AWS::IoT::Certificate::SerialNumber"
          }
        },
        "ThingTypeName": "lightBulb-versionA",
        "ThingGroups": [
          "v1-lightbulbs",
          {
            "Ref": "AWS::IoT::Certificate::Country"
          }
        ]
      },
      "OverrideSettings": {
        "AttributePayload": "MERGE",
        "ThingTypeName": "REPLACE",
        "ThingGroups": "DO_NOTHING"
      }
    },
    "certificate": {
      "Type": "AWS::IoT::Certificate",
      "Properties": {
        "CertificateId": {
          "Ref": "AWS::IoT::Certificate::Id"
        },
        "Status": "ACTIVE"
      }
    },
    "policy": {
      "Type": "AWS::IoT::Policy",
      "Properties": {
        "PolicyDocument":"{ \"Version\": \"2012-10-17\", \"Statement\": [{ \"Effect\": \"Allow\", \"Action\":[\"iot:Publish\"], \"Resource\": [\"arn:${data.aws_partition.current.partition}:iot:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:topic/foo/bar\"] }] }"
      }
    }
  } 
}
    EOF
  }
}

data "aws_iot_registration_code" "test" {}
`)
}
