package acmpca_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccACMPCAPermission_basic(t *testing.T) {
	var permission acmpca.Permission
	resourceName := "aws_acmpca_permission.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acmpca.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &permission),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "GetCertificate"),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "IssueCertificate"),
					resource.TestCheckTypeSetElemAttr(resourceName, "actions.*", "ListPermissions"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttr(resourceName, "principal", "acm.amazonaws.com"),
					acctest.CheckResourceAttrAccountID(resourceName, "source_account"),
				),
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

func testAccPermissionConfig_basic(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, commonName)
}
