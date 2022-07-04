package acmpca_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccACMPCAPermission_Valid(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acmpca.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_valid(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
				),
			},
		},
	})
}

func TestAccACMPCAPermission_InvalidPrincipal(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acmpca.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionConfig_invalidPrincipal(),
				ExpectError: regexp.MustCompile("Error: expected principal to be one of .*, got .*"),
			},
		},
	})
}

func TestAccACMPCAPermission_InvalidActionsEntry(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acmpca.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionConfig_invalidActionsEntry(),
				ExpectError: regexp.MustCompile("Error: expected actions.1 to be one of .*, got .*"),
			},
		},
	})
}

func testAccCheckPermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_permission" {
			continue
		}

		input := &acmpca.ListPermissionsInput{
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.ListPermissions(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
				tfawserr.ErrCodeEquals(err, acmpca.ErrCodeInvalidStateException) {
				return nil
			}
			return err
		}

		if output != nil {
			return fmt.Errorf("ACMPCA Permission %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckPermissionExists(resourceName string, permission *acmpca.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn
		input := &acmpca.ListPermissionsInput{
			CertificateAuthorityArn: aws.String(rs.Primary.Attributes["certificate_authority_arn"]),
		}

		output, err := conn.ListPermissions(input)

		if err != nil {
			return err
		}

		if output == nil || output.Permissions == nil {
			return fmt.Errorf("ACMPCA Permission %q does not exist", rs.Primary.ID)
		}

		*permission = *output.Permissions[0]

		return nil
	}
}

func testAccCertificateAuthority() string {
	return `
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}
`
}

func testAccPermissionConfig_valid() string {
	return acctest.ConfigCompose(
		testAccCertificateAuthority(),
		`
resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`)
}

func testAccPermissionConfig_invalidPrincipal() string {
	return acctest.ConfigCompose(
		testAccCertificateAuthority(),
		`
resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "notacm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`)
}

func testAccPermissionConfig_invalidActionsEntry() string {
	return acctest.ConfigCompose(
		testAccCertificateAuthority(),
		`
resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCert", "GetCertificate", "ListPermissions"]
}
`)
}
