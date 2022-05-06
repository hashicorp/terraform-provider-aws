package appconfig_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
)

func TestAccAppConfigApplication_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAppConfigApplication_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappconfig.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppConfigApplication_updateName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
				),
			},
			{
				Config: testAccApplicationNameConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func TestAccAppConfigApplication_updateDescription(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := sdkacctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDescriptionConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationDescriptionConfig(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test Description Removal
				Config: testAccApplicationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
				),
			},
		},
	})
}

func TestAccAppConfigApplication_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
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
				Config: testAccApplicationTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccApplicationTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appconfig_application" {
			continue
		}

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)

		if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading AppConfig Application (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("AppConfig Application (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckApplicationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn

		input := &appconfig.GetApplicationInput{
			ApplicationId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetApplication(input)

		if err != nil {
			return fmt.Errorf("error reading AppConfig Application (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Application (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccApplicationNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccApplicationDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name        = %q
  description = %q
}
`, rName, description)
}

func testAccApplicationTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
