package licensemanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/licensemanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflicensemanager "github.com/hashicorp/terraform-provider-aws/internal/service/licensemanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLicenseManagerAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "license_configuration_arn", "aws_licensemanager_license_configuration.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_instance.test", "arn"),
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

func TestAccLicenseManagerAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_licensemanager_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflicensemanager.ResourceAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No License Manager Association ID is set")
		}

		resourceARN, licenseConfigurationARN, err := tflicensemanager.AssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn

		err = tflicensemanager.FindAssociation(context.TODO(), conn, resourceARN, licenseConfigurationARN)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_licensemanager_association" {
			continue
		}

		resourceARN, licenseConfigurationARN, err := tflicensemanager.AssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		err = tflicensemanager.FindAssociation(context.TODO(), conn, resourceARN, licenseConfigurationARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("License Manager Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"

  tags = {
    Name = %[1]q
  }
}

resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "vCPU"
}

resource "aws_licensemanager_association" "test" {
  license_configuration_arn = aws_licensemanager_license_configuration.test.id
  resource_arn              = aws_instance.test.arn
}
`, rName))
}
