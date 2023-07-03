// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransferCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, domainWildcard)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
		},
	})
}

func TestAccTransferCertificate_certificate(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_certificate(caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
		},
	})
}

func TestAccTransferCertificate_certificateChain(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, domainWildcard)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_certificateChain(certificate, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
		},
	})
}

func TestAccTransferCertificate_certificateKey(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_certificatePrivateKey(certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
		},
	})
}

func TestAccTransferCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_certificate(certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTransferCertificate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_tags1(certificate, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
			{
				Config: testAccCertificateConfig_tags2(certificate, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(certificate, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccTransferCertificate_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf transfer.DescribedCertificate
	resourceName := "aws_transfer_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_description(certificate, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "desc1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate", "certificate_chain"},
			},
			{
				Config: testAccCertificateConfig_description(certificate, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "description", "desc2"),
				),
			},
		},
	})
}

func testAccCheckCertificateExists(ctx context.Context, n string, v *transfer.DescribedCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		output, err := tftransfer.FindCertificateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_certificate" {
				continue
			}

			_, err := tftransfer.FindCertificateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCertificateConfig_basic(certificate, privateKey, caCertificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate       = %[1]q
  private_key       = %[2]q
  certificate_chain = %[3]q
  usage             = "SIGNING"
}
`, certificate, privateKey, caCertificate)
}

func testAccCertificateConfig_certificate(certificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[1]q
  usage       = "SIGNING"
}
`, certificate)
}

func testAccCertificateConfig_certificateChain(certificate, caCertificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate       = %[1]q
  certificate_chain = %[2]q
  usage             = "SIGNING"
}
`, certificate, caCertificate)
}

func testAccCertificateConfig_certificatePrivateKey(certificate, privateKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[1]q
  private_key = %[2]q
  usage       = "SIGNING"
}
`, certificate, privateKey)
}

func testAccCertificateConfig_tags1(certificate, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[1]q
  usage       = "SIGNING"

  tags = {
    %[2]q = %[3]q
  }
}
`, certificate, tagKey1, tagValue1)
}

func testAccCertificateConfig_tags2(certificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[1]q
  usage       = "SIGNING"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, certificate, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCertificateConfig_description(certificate, description string) string {
	return fmt.Sprintf(`
resource "aws_transfer_certificate" "test" {
  certificate = %[1]q
  usage       = "SIGNING"
  description = %[2]q
}
`, certificate, description)
}
