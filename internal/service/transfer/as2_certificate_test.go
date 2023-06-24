package transfer_test

import (
	"context"
	"fmt"

	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
	resourceName := "aws_transfer_as2_certificate.test"
	//commonName := "example.com"
	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	//certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)
	//certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, domainWildcard)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testCertificate_basic(rName, certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_transfer_as2_certificate.test"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testCertificate_certificate(caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_transfer_as2_certificate.test"
	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, domainWildcard)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testCertificate_certificatechain(certificate, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_transfer_as2_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testCertificate_certificatekey(certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRFC3339(resourceName, "active_date"),
					acctest.CheckResourceAttrRFC3339(resourceName, "inactive_date"),
					resource.TestCheckResourceAttr(resourceName, "usage", "SIGNING"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_transfer_as2_certificate.test"
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, acctest.RandomSubdomain())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, transfer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testCertificate_certificate(certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCertificate_basic(rName string, certificate string, key string, caCertificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate       = %[2]q
  private_key       = %[3]q
  certificate_chain = %[4]q
  usage             = "SIGNING"
}
`, rName, certificate, key, caCertificate)
}

func testCertificate_certificate(certificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate = %[1]q
  usage       = "SIGNING"
}
`, certificate)
}

func testCertificate_certificatechain(certificate string, caCertificate string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate       = %[1]q
  certificate_chain = %[2]q
  usage             = "SIGNING"
}
`, certificate, caCertificate)
}

func testCertificate_certificatekey(certificate string, key string) string {
	return fmt.Sprintf(`
resource "aws_transfer_as2_certificate" "test" {
  certificate = %[1]q
  private_key = %[2]q
  usage       = "SIGNING"
}
`, certificate, key)
}

func testAccCheckCertificateExists(ctx context.Context, n string, v *transfer.DescribedCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Certificate ID is set")
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
			if rs.Type != "aws_transfer_as2_certificate" {
				continue
			}

			_, err := tftransfer.FindCertificateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}
