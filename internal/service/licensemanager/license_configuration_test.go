package licensemanager_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLicenseManagerLicenseConfiguration_basic(t *testing.T) {
	var licenseConfiguration licensemanager.LicenseConfiguration
	resourceName := "aws_licensemanager_license_configuration.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLicenseConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseConfigurationExists(resourceName, &licenseConfiguration),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "license-manager", regexp.MustCompile(`license-configuration:lic-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", "Example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Example"),
					resource.TestCheckResourceAttr(resourceName, "license_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "true"),
					resource.TestCheckResourceAttr(resourceName, "license_counting_type", "Socket"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.0", "#minimumSockets=3"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "barr"),
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

func TestAccLicenseManagerLicenseConfiguration_update(t *testing.T) {
	var licenseConfiguration licensemanager.LicenseConfiguration
	resourceName := "aws_licensemanager_license_configuration.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, licensemanager.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLicenseConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseConfigurationExists(resourceName, &licenseConfiguration),
				),
			},
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseConfigurationExists(resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", "NewName"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "license_count", "123"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.abc", "def"),
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

func testAccCheckLicenseConfigurationExists(resourceName string, licenseConfiguration *licensemanager.LicenseConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn
		resp, err := conn.ListLicenseConfigurations(&licensemanager.ListLicenseConfigurationsInput{
			LicenseConfigurationArns: [](*string){aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error retrieving License Manager license configuration (%s): %s", rs.Primary.ID, err)
		}

		if len(resp.LicenseConfigurations) == 0 {
			return fmt.Errorf("Error retrieving License Manager license configuration (%s): Not found", rs.Primary.ID)
		}

		*licenseConfiguration = *resp.LicenseConfigurations[0]
		return nil
	}
}

func testAccCheckLicenseConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LicenseManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_licensemanager_license_configuration" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetLicenseConfiguration(&licensemanager.GetLicenseConfigurationInput{
			LicenseConfigurationArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeInvalidParameterValueException) {
				continue
			}
			return err
		}
	}

	return nil

}

const testAccLicenseManagerLicenseConfigurationConfig_basic = `
resource "aws_licensemanager_license_configuration" "example" {
  name                     = "Example"
  description              = "Example"
  license_count            = 10
  license_count_hard_limit = true
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]

  tags = {
    foo = "barr"
  }
}
`

const testAccLicenseManagerLicenseConfigurationConfig_update = `
resource "aws_licensemanager_license_configuration" "example" {
  name                  = "NewName"
  license_count         = 123
  license_counting_type = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]

  tags = {
    test = "test"
    abc  = "def"
  }
}
`
