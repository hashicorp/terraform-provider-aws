package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccIAMSigningCertificate_basic(t *testing.T) {
	var cred iam.SigningCertificate

	resourceName := "aws_iam_signing_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateBasicConfig(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, "user_name", "aws_iam_user.test", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_body"),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIAMSigningCertificate_status(t *testing.T) {
	var cred iam.SigningCertificate

	resourceName := "aws_iam_signing_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateConfigStatus(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSigningCertificateConfigStatus(rName, "Active", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
				),
			},
			{
				Config: testAccSigningCertificateConfigStatus(rName, "Inactive", certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
		},
	})
}

func TestAccIAMSigningCertificate_disappears(t *testing.T) {
	var cred iam.SigningCertificate
	resourceName := "aws_iam_signing_certificate.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSigningCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSigningCertificateBasicConfig(rName, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningCertificateExists(resourceName, &cred),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceSigningCertificate(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceSigningCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSigningCertificateExists(n string, cred *iam.SigningCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server Cert ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		certId, userName, err := tfiam.DecodeSigningCertificateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindSigningCertificate(conn, userName, certId)
		if err != nil {
			return err
		}

		*cred = *output

		return nil
	}
}

func testAccCheckSigningCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_signing_certificate" {
			continue
		}

		certId, userName, err := tfiam.DecodeSigningCertificateId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindSigningCertificate(conn, userName, certId)

		if tfresource.NotFound(err) {
			continue
		}

		if output != nil {
			return fmt.Errorf("IAM Service Specific Credential (%s) still exists", rs.Primary.ID)
		}

	}

	return nil
}

func testAccSigningCertificateBasicConfig(rName, cert string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_signing_certificate" "test" {
  certificate_body = "%[2]s"
  user_name        = aws_iam_user.test.name
}
`, rName, acctest.TLSPEMEscapeNewlines(cert))
}

func testAccSigningCertificateConfigStatus(rName, status, cert string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_signing_certificate" "test" {
  certificate_body = "%[3]s"
  user_name        = aws_iam_user.test.name
  status           = %[2]q
}
`, rName, status, acctest.TLSPEMEscapeNewlines(cert))
}
