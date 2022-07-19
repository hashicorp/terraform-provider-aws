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

func TestAccIAMServiceSpecificCredential_basic(t *testing.T) {
	var cred iam.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, "user_name", "aws_iam_user.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttrSet(resourceName, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_multi(t *testing.T) {
	var cred iam.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	resourceName2 := "aws_iam_service_specific_credential.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_multi(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					resource.TestCheckResourceAttrPair(resourceName, "user_name", "aws_iam_user.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "service_name", "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttrSet(resourceName, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName, "service_specific_credential_id"),
					resource.TestCheckResourceAttrPair(resourceName2, "user_name", "aws_iam_user.test", "name"),
					resource.TestCheckResourceAttr(resourceName2, "service_name", "codecommit.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName2, "status", "Active"),
					resource.TestCheckResourceAttrSet(resourceName2, "service_user_name"),
					resource.TestCheckResourceAttrSet(resourceName2, "service_specific_credential_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password"},
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_status(t *testing.T) {
	var cred iam.ServiceSpecificCredentialMetadata

	resourceName := "aws_iam_service_specific_credential.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_password"},
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Active"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
				),
			},
			{
				Config: testAccServiceSpecificCredentialConfig_status(rName, "Inactive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
				),
			},
		},
	})
}

func TestAccIAMServiceSpecificCredential_disappears(t *testing.T) {
	var cred iam.ServiceSpecificCredentialMetadata
	resourceName := "aws_iam_service_specific_credential.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckServiceSpecificCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSpecificCredentialConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceSpecificCredentialExists(resourceName, &cred),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceServiceSpecificCredential(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceServiceSpecificCredential(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceSpecificCredentialExists(n string, cred *iam.ServiceSpecificCredentialMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Server Cert ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		serviceName, userName, credId, err := tfiam.DecodeServiceSpecificCredentialId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindServiceSpecificCredential(conn, serviceName, userName, credId)
		if err != nil {
			return err
		}

		*cred = *output

		return nil
	}
}

func testAccCheckServiceSpecificCredentialDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_service_specific_credential" {
			continue
		}

		serviceName, userName, credId, err := tfiam.DecodeServiceSpecificCredentialId(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := tfiam.FindServiceSpecificCredential(conn, serviceName, userName, credId)

		if tfresource.NotFound(err) {
			continue
		}

		if output != nil {
			return fmt.Errorf("IAM Service Specific Credential (%s) still exists", rs.Primary.ID)
		}

	}

	return nil
}

func testAccServiceSpecificCredentialConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}
`, rName)
}

func testAccServiceSpecificCredentialConfig_multi(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}

resource "aws_iam_service_specific_credential" "test2" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
}
`, rName)
}

func testAccServiceSpecificCredentialConfig_status(rName, status string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_service_specific_credential" "test" {
  service_name = "codecommit.amazonaws.com"
  user_name    = aws_iam_user.test.name
  status       = %[2]q
}
`, rName, status)
}
