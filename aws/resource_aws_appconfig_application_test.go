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

func TestAccAWSAppConfigApplication_basic(t *testing.T) {
	var application appconfig.GetApplicationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rDesc := acctest.RandomWithPrefix("desc")
	resourceName := "aws_appconfig_application.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationName(rName, rDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckAWSAppConfigApplicationARN(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", rDesc),
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

func TestAccAWSAppConfigApplication_disappears(t *testing.T) {
	var application appconfig.GetApplicationOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rDesc := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationName(rName, rDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName, &application),
					testAccCheckAWSAppConfigApplicationDisappears(&application),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAppConfigApplication_Tags(t *testing.T) {
	var application appconfig.GetApplicationOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAppConfigApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppConfigApplicationTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAppConfigApplicationTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAppConfigApplicationTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppConfigApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAppConfigApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appconfigconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_application" {
			continue
		}

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)

		if isAWSErr(err, appconfig.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil {
			return fmt.Errorf("AppConfig Application (%s) still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSAppConfigApplicationDisappears(application *appconfig.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		input := &appconfig.DeleteApplicationInput{
			ApplicationId: aws.String(*application.Id),
		}

		_, err := conn.DeleteApplication(input)

		return err
	}
}

func testAccCheckAWSAppConfigApplicationExists(resourceName string, application *appconfig.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appconfigconn

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)
		if err != nil {
			return err
		}

		*application = *output

		return nil
	}
}

func testAccCheckAWSAppConfigApplicationARN(resourceName string, application *appconfig.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "appconfig", fmt.Sprintf("application/%s", aws.StringValue(application.Id)))(s)
	}
}

func testAccAWSAppConfigApplicationName(rName, rDesc string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
	name        = %[1]q
	description = %[2]q
}
`, rName, rDesc)
}

func testAccAWSAppConfigApplicationTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAppConfigApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
