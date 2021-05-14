package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppConfigHostedConfigurationVersion_basic(t *testing.T) {
	var out appconfig.GetHostedConfigurationVersionOutput
	resourceName := "aws_appconfig_hosted_configuration_version.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName, &out),
					testAccCheckAWSAppConfigHostedConfigurationVersionARN(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "description", "hosted configuration version description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAppConfigHostedConfigurationVersionImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppConfigHostedConfigurationVersion_disappears(t *testing.T) {
	var out appconfig.GetHostedConfigurationVersionOutput
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName, &out),
					testAccCheckAWSAppConfigHostedConfigurationVersionDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppConfigHostedConfigurationVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_hosted_configuration_version" {
			continue
		}
		versionNumber, err := strconv.Atoi(rs.Primary.Attributes["version_number"])
		if err != nil {
			return fmt.Errorf("failed to convert version_number into int (%s): %w", rs.Primary.Attributes["version_number"], err)
		}

		input := &appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(rs.Primary.Attributes["application_id"]),
			ConfigurationProfileId: aws.String(rs.Primary.Attributes["configuration_profile_id"]),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		}

		output, err := conn.GetHostedConfigurationVersion(input)

		if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AppConfig HostedConfigurationVersion (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigHostedConfigurationVersionDisappears(hcv *appconfig.GetHostedConfigurationVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		_, err := conn.DeleteHostedConfigurationVersion(&appconfig.DeleteHostedConfigurationVersionInput{
			ApplicationId:          hcv.ApplicationId,
			ConfigurationProfileId: hcv.ConfigurationProfileId,
			VersionNumber:          hcv.VersionNumber,
		})

		return err
	}
}

func testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName string, hcv *appconfig.GetHostedConfigurationVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		versionNumber, err := strconv.Atoi(rs.Primary.Attributes["version_number"])
		if err != nil {
			return fmt.Errorf("failed to convert version_number into int (%s): %w", rs.Primary.Attributes["version_number"], err)
		}

		output, err := conn.GetHostedConfigurationVersion(&appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(rs.Primary.Attributes["application_id"]),
			ConfigurationProfileId: aws.String(rs.Primary.Attributes["configuration_profile_id"]),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		})
		if err != nil {
			return err
		}

		*hcv = *output

		return nil
	}
}

func testAccCheckAWSAppConfigHostedConfigurationVersionARN(resourceName string, hcv *appconfig.GetHostedConfigurationVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appconfig", fmt.Sprintf("application/%s/configurationprofile/%s/hostedconfigurationversion/%d", aws.StringValue(hcv.ApplicationId), aws.StringValue(hcv.ConfigurationProfileId), aws.Int64Value(hcv.VersionNumber)))(s)
	}
}

func testAccAWSAppConfigHostedConfigurationVersion() string {
	appName := acctest.RandomWithPrefix("tf-acc-test")
	profileName := acctest.RandomWithPrefix("tf-acc-test")
	return testAccAWSAppConfigApplicationName(appName, "test") + testAccAWSAppConfigConfigurationProfile(profileName, "test") + `
resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.id
  description              = "hosted configuration version description"
  content_type             = "application/json"
  content = jsonencode({
    foo = "foo"
  })
}
`
}

func testAccAWSAppConfigHostedConfigurationVersionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return rs.Primary.ID, nil
	}
}
