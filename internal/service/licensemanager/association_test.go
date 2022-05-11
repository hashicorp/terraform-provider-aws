package licensemanager_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
)

func TestAccLicenseManagerAssociation_basic(t *testing.T) {
	var licenseSpecification licensemanager.LicenseSpecification
	resourceName := "aws_licensemanager_association.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName, &licenseSpecification),
					resource.TestCheckResourceAttrPair(resourceName, "license_configuration_arn", "aws_licensemanager_license_configuration.example", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_instance.example", "arn"),
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

func testAccCheckAssociationExists(resourceName string, licenseSpecification *licensemanager.LicenseSpecification) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resourceArn, licenseConfigurationArn, err := tflicensemanager.AssociationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		specification, err := tflicensemanager.AssociationFindSpecification(conn, resourceArn, licenseConfigurationArn)
		if err != nil {
			return err
		}

		if specification == nil {
			return fmt.Errorf("Error retrieving License Manager association (%s): Not found", rs.Primary.ID)
		}

		*licenseSpecification = *specification
		return nil
	}
}

func testAccCheckAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_licensemanager_association" {
			continue
		}

		resourceArn, licenseConfigurationArn, err := tflicensemanager.AssociationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		specification, err := tflicensemanager.AssociationFindSpecification(conn, resourceArn, licenseConfigurationArn)
		if err != nil {
			return err
		}

		if specification != nil {
			return fmt.Errorf("License Manager association %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccLicenseManagerAssociationConfig_basic = `
data "aws_ami" "example" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }
}

resource "aws_instance" "example" {
  ami           = data.aws_ami.example.id
  instance_type = "t2.micro"
}

resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "vCPU"
}

resource "aws_licensemanager_association" "example" {
  license_configuration_arn = aws_licensemanager_license_configuration.example.id
  resource_arn              = aws_instance.example.arn
}
`
