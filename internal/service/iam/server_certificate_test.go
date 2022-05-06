package iam_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMServerCertificate_basic(t *testing.T) {
	var cert iam.ServerCertificate

	resourceName := "aws_iam_server_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("server-certificate/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "expiration"),
					acctest.CheckResourceAttrRFC3339(resourceName, "upload_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", strings.TrimSpace(certificate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
		},
	})
}

func TestAccIAMServerCertificate_tags(t *testing.T) {
	var cert iam.ServerCertificate

	resourceName := "aws_iam_server_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_tags1(rName, key, certificate, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
			{
				Config: testAccServerCertificateConfig_tags2(rName, key, certificate, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServerCertificateConfig_tags1(rName, key, certificate, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_Name_prefix(t *testing.T) {
	var cert iam.ServerCertificate

	resourceName := "aws_iam_server_certificate.test"

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_random(key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_disappears(t *testing.T) {
	var cert iam.ServerCertificate
	resourceName := "aws_iam_server_certificate.test"

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_random(key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceServerCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMServerCertificate_file(t *testing.T) {
	var cert iam.ServerCertificate

	rInt := sdkacctest.RandInt()
	unixFile := "test-fixtures/iam-ssl-unix-line-endings.pem"
	winFile := "test-fixtures/iam-ssl-windows-line-endings.pem.winfile"
	resourceName := "aws_iam_server_certificate.test"
	resourceId := fmt.Sprintf("terraform-test-cert-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_file(rInt, unixFile),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           resourceId,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
			{
				Config: testAccServerCertificateConfig_file(rInt, winFile),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
				),
			},
		},
	})
}

func TestAccIAMServerCertificate_path(t *testing.T) {
	var cert iam.ServerCertificate

	resourceName := "aws_iam_server_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerCertificateConfig_path(rName, "/test/", key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "path", "/test/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           rName,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
		},
	})
}

func testAccCheckCertExists(n string, cert *iam.ServerCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server Cert ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		describeOpts := &iam.GetServerCertificateInput{
			ServerCertificateName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.GetServerCertificate(describeOpts)
		if err != nil {
			return err
		}

		*cert = *resp.ServerCertificate

		return nil
	}
}

func testAccCheckServerCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_server_certificate" {
			continue
		}

		// Try to find the Cert
		opts := &iam.GetServerCertificateInput{
			ServerCertificateName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.GetServerCertificate(opts)
		if err == nil {
			if resp.ServerCertificate != nil {
				return fmt.Errorf("Error: Server Cert still exists")
			}

			return nil
		}

	}

	return nil
}

func testAccServerCertificateConfig_basic(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateConfig_random(key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name_prefix      = "tf-acc-test"
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccServerCertificateConfig_path(rName, path, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  path             = "%[2]s"
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}
`, rName, path, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

// iam-ssl-unix-line-endings
func testAccServerCertificateConfig_file(rInt int, fName string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = "terraform-test-cert-%d"
  certificate_body = file("%s")

  private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDKdH6BU9Q0xBVPfeX5NjCC/B2Pm3WsFGnTtRw4abkD+r4to9wD
eYUgjH2yPCyonNOA8mNiCQgDTtaLfbA8LjBYoodt7rgaTO7C0ugRtmTNK96DmYxm
f8Gs5ZS6eC3yeaFv58d1w2mow7tv0+DRk8uXwzVfaaMxoalsCtlLznmZHwIDAQAB
AoGABZj69nBu6ZaSUERW23EYHkcCOjo+Iqfd1TCouxaROv7vyytApgfyGlhIEWmA
gpjzcBlDji5Zvl2rqOesu707MOuJavZvluo+JHy/VIuU+yGUrWuO/QVCu6Jn3yns
vS7g48ConuZ962cTzRPcpPDspONBVOAhVCF33Y8PsnxV0wECQQD5RqeoqxEUupsy
QhrDui0KkYXLdT0uhrEQ69n9rvAiQoHPsiX0MswfEKnj/g9N3VwGLdgWytT0TvcI
8fDPRB4/AkEAz+qF3taX77gB69XRPQwCGWqE1fHIFMwX7QeYdEsk3iRZ0EKVcdp6
vIPCB2Cq4a4eXcaFa/bXen4yeYgyTbeNIQJBAO92dWctdoowPRiJskZmGhC1/Q6X
gH+qenyj5VSy8hInS6anH5i4F6icDGhtzmvhgx6YeaZjkTFkjiG0sb2aVWcCQQDD
WL7UwtzX/xPXB/ril5C1Xo5WESgC2ks0ielkgmGuUYsNEDInWbXtvwGjOuDyz0x6
oRYkfTSxQzabVyqkOGvhAkBtbjUxOD8wgBIjb4T6mAMokQo6PeEAZGUTyPifjJNo
detWVr2WRvgNgQvcRnNPECwfq1RtMJJpavaI3kgeaSxg
-----END RSA PRIVATE KEY-----
EOF
}
`, rInt, fName)
}

func testAccServerCertificateConfig_tags1(rName, key, certificate, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    %[4]q = %[5]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1)
}

func testAccServerCertificateConfig_tags2(rName, key, certificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1, tagKey2, tagValue2)
}
