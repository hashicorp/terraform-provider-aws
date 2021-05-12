package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppConfigConfigurationProfile_basic(t *testing.T) {
	var profile appconfig.GetConfigurationProfileOutput
	profileName := acctest.RandomWithPrefix("tf-acc-test")
	profileDesc := acctest.RandomWithPrefix("desc")
	resourceName := "aws_appconfig_configuration_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigConfigurationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigConfigurationProfile(profileName, profileDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigConfigurationProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "name", profileName),
					testAccCheckAWSAppConfigConfigurationProfileARN(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", profileDesc),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAppConfigConfigurationProfileImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSAppConfigConfigurationProfile_disappears(t *testing.T) {
	var profile appconfig.GetConfigurationProfileOutput

	profileName := acctest.RandomWithPrefix("tf-acc-test")
	profileDesc := acctest.RandomWithPrefix("desc")
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigConfigurationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigConfigurationProfile(profileName, profileDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigConfigurationProfileExists(resourceName, &profile),
					testAccCheckAWSAppConfigConfigurationProfileDisappears(&profile),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigConfigurationProfile_Tags(t *testing.T) {
	var profile appconfig.GetConfigurationProfileOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appconfig.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigConfigurationProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigConfigurationProfileTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigConfigurationProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSAppConfigConfigurationProfileImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigConfigurationProfileTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigConfigurationProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigConfigurationProfileTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigConfigurationProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppConfigConfigurationProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_configuration_profile" {
			continue
		}

		input := &appconfig.GetConfigurationProfileInput{
			ApplicationId:          aws.String(rs.Primary.Attributes["application_id"]),
			ConfigurationProfileId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetConfigurationProfile(input)

		if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AppConfig ConfigurationProfile (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSAppConfigConfigurationProfileDisappears(profile *appconfig.GetConfigurationProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		_, err := conn.DeleteConfigurationProfile(&appconfig.DeleteConfigurationProfileInput{
			ApplicationId:          profile.ApplicationId,
			ConfigurationProfileId: profile.Id,
		})

		return err
	}
}

func testAccCheckAWSAppConfigConfigurationProfileExists(resourceName string, profile *appconfig.GetConfigurationProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		output, err := conn.GetConfigurationProfile(&appconfig.GetConfigurationProfileInput{
			ApplicationId:          aws.String(rs.Primary.Attributes["application_id"]),
			ConfigurationProfileId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*profile = *output

		return nil
	}
}

func testAccCheckAWSAppConfigConfigurationProfileARN(resourceName string, profile *appconfig.GetConfigurationProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appconfig", fmt.Sprintf("application/%s/configurationprofile/%s", aws.StringValue(profile.ApplicationId), aws.StringValue(profile.Id)))(s)
	}
}

func testAccAWSAppConfigConfigurationProfile(profileName, profileDesc string) string {
	appName := acctest.RandomWithPrefix("tf-acc-test")
	appDesc := acctest.RandomWithPrefix("desc")
	return testAccAWSAppConfigApplicationName(appName, appDesc) + fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  name           = %[1]q
  description    = %[2]q
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
  validators {
    type = "JSON_SCHEMA"
    content = jsonencode({
      "$schema" = "http://json-schema.org/draft-04/schema#"
      title = "$id$"
      description = "BasicFeatureToggle-1"
      type = "object"
      additionalProperties = false
      patternProperties = {
        "[^\\s]+$" = {
          type = "boolean"
        }
      }
      minProperties = 1
    })
  }
}
`, profileName, profileDesc)
}

func testAccAWSAppConfigConfigurationProfileTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSAppConfigApplicationTags1(rName, tagKey1, tagValue1) + fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAppConfigConfigurationProfileTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSAppConfigApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2) + fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAppConfigConfigurationProfileImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["application_id"], rs.Primary.ID), nil
	}
}
