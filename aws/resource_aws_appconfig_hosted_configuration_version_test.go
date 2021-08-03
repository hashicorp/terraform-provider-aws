package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppConfigHostedConfigurationVersion_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}/configurationprofile/[a-z0-9]{4,7}/hostedconfigurationversion/[0-9]+`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_appconfig_application.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", "aws_appconfig_configuration_profile.test", "configuration_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "content", "{\"foo\":\"bar\"}"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
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

func TestAccAWSAppConfigHostedConfigurationVersion_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigHostedConfigurationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigHostedConfigurationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppconfigHostedConfigurationVersion(), resourceName),
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

		appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		}

		output, err := conn.GetHostedConfigurationVersion(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Hosted Configuration Version (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigHostedConfigurationVersionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		appID, confProfID, versionNumber, err := resourceAwsAppconfigHostedConfigurationVersionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetHostedConfigurationVersion(&appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		})

		if err != nil {
			return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Hosted Configuration Version (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSAppConfigHostedConfigurationVersion(rName string) string {
	return composeConfig(
		testAccAWSAppConfigConfigurationProfileConfigName(rName),
		fmt.Sprintf(`
resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %q
}
`, rName))
}
