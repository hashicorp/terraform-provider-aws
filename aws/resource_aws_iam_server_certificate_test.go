package aws

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMServerCertificate_basic(t *testing.T) {
	var cert iam.ServerCertificate
	rInt := acctest.RandInt()
	var certBody string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServerCertConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists("aws_iam_server_certificate.test_cert", &cert),
					getCertBody(&certBody),
					testAccCheckAWSServerCertAttributes(&cert, &certBody),
				),
			},
		},
	})
}

func TestAccAWSIAMServerCertificate_name_prefix(t *testing.T) {
	var cert iam.ServerCertificate
	var certBody string
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServerCertConfig_random(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists("aws_iam_server_certificate.test_cert", &cert),
					getCertBody(&certBody),
					testAccCheckAWSServerCertAttributes(&cert, &certBody),
				),
			},
		},
	})
}

func TestAccAWSIAMServerCertificate_disappears(t *testing.T) {
	var cert iam.ServerCertificate
	rInt := acctest.RandInt()

	testDestroyCert := func(*terraform.State) error {
		// reach out and DELETE the Cert
		conn := testAccProvider.Meta().(*AWSClient).iamconn
		_, err := conn.DeleteServerCertificate(&iam.DeleteServerCertificateInput{
			ServerCertificateName: cert.ServerCertificateMetadata.ServerCertificateName,
		})

		if err != nil {
			return fmt.Errorf("Error destroying cert in test: %s", err)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServerCertConfig_random(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists("aws_iam_server_certificate.test_cert", &cert),
					testDestroyCert,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMServerCertificate_file(t *testing.T) {
	var _cert iam.ServerCertificate

	rInt := acctest.RandInt()
	unix, err := ioutil.ReadFile("test-fixtures/iam-ssl-unix-line-endings.pem")
	if err != nil {
		t.Fatalf("error loading test file: %s", err)
	}
	certBodyUnix := string(unix)
	win, err := ioutil.ReadFile("test-fixtures/iam-ssl-windows-line-endings.pem.winfile")
	if err != nil {
		t.Fatalf("error loading test file: %s", err)
	}
	certBodyWin := string(win)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMServerCertConfig_file(rInt, certBodyUnix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists("aws_iam_server_certificate.test_cert", &_cert),
				),
			},
			{
				Config: testAccIAMServerCertConfig_file(rInt, certBodyWin),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertExists("aws_iam_server_certificate.test_cert", &_cert),
				),
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

		conn := testAccProvider.Meta().(*AWSClient).iamconn
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

func getCertBody(body *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "tls_self_signed_cert" {
				continue
			}

			*body = rs.Primary.Attributes["cert_pem"]
		}
		return nil
	}
}

func testAccCheckAWSServerCertAttributes(cert *iam.ServerCertificate, certBody *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*cert.ServerCertificateMetadata.ServerCertificateName, "terraform-test-cert") {
			return fmt.Errorf("Bad Server Cert Name: %s", *cert.ServerCertificateMetadata.ServerCertificateName)
		}

		if *cert.CertificateBody != strings.TrimSpace(*certBody) {
			return fmt.Errorf("Bad Server Cert body\n\t expected: %s\n\tgot: %s\n", *certBody, *cert.CertificateBody)
		}
		return nil
	}
}

func testAccCheckIAMServerCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

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

var certBody = fmt.Sprintf(`
-----BEGIN CERTIFICATE-----
MIIDBjCCAe4CCQCGWwBmOiHQdTANBgkqhkiG9w0BAQUFADBFMQswCQYDVQQGEwJB
VTETMBEGA1UECBMKU29tZS1TdGF0ZTEhMB8GA1UEChMYSW50ZXJuZXQgV2lkZ2l0
cyBQdHkgTHRkMB4XDTE2MDYyMTE2MzM0MVoXDTE3MDYyMTE2MzM0MVowRTELMAkG
A1UEBhMCQVUxEzARBgNVBAgTClNvbWUtU3RhdGUxITAfBgNVBAoTGEludGVybmV0
IFdpZGdpdHMgUHR5IEx0ZDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AL+LFlsCJG5txZp4yuu+lQnuUrgBXRG+irQqcTXlV91Bp5hpmRIyhnGCtWxxDBUL
xrh4WN3VV/0jDzKT976oLgOy3hj56Cdqf+JlZ1qgMN5bHB3mm3aVWnrnsLbBsfwZ
SEbk3Kht/cE1nK2toNVW+rznS3m+eoV3Zn/DUNwGlZr42hGNs6ETn2jURY78ETqR
mW47xvjf86eIo7vULHJaY6xyarPqkL8DZazOmvY06hUGvGwGBny7gugfXqDG+I8n
cPBsGJGSAmHmVV8o0RCB9UjY+TvSMQRpEDoVlvyrGuglsD8to/4+7UcsuDGlRYN6
jmIOC37mOi/jwRfWL1YUa4MCAwEAATANBgkqhkiG9w0BAQUFAAOCAQEAPDxTH0oQ
JjKXoJgkmQxurB81RfnK/NrswJVzWbOv6ejcbhwh+/ZgJTMc15BrYcxU6vUW1V/i
Z7APU0qJ0icECACML+a2fRI7YdLCTiPIOmY66HY8MZHAn3dGjU5TeiUflC0n0zkP
mxKJe43kcYLNDItbfvUDo/GoxTXrC3EFVZyU0RhFzoVJdODlTHXMVFCzcbQEBrBJ
xKdShCEc8nFMneZcGFeEU488ntZoWzzms8/QpYrKa5S0Sd7umEU2Kwu4HTkvUFg/
CqDUFjhydXxYRsxXBBrEiLOE5BdtJR1sH/QHxIJe23C9iHI2nS1NbLziNEApLwC4
GnSud83VUo9G9w==
-----END CERTIFICATE-----`)

const testAccTLSServerCert = `
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

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
}
`

func testAccIAMServerCertConfig(rInt int) string {
	return fmt.Sprintf(`
%s

resource "aws_iam_server_certificate" "test_cert" {
  name = "terraform-test-cert-%d"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}
`, testAccTLSServerCert, rInt)
}

func testAccIAMServerCertConfig_random(rInt int) string {
	return fmt.Sprintf(`
%s 

resource "aws_iam_server_certificate" "test_cert" {
  name_prefix = "terraform-test-cert"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}
`, testAccTLSServerCert)
}

// iam-ssl-unix-line-endings
func testAccIAMServerCertConfig_file(rInt int, fName string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test_cert" {
  name = "terraform-test-cert-%d"
  certificate_body = <<EOF
%s
EOF

	private_key =  <<EOF
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
