package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func testAccAwsAppStreamStack_basic(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func testAccAwsAppStreamStack_Name_Generated(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func testAccAwsAppStreamStack_NamePrefix(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigNamePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
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

func testAccAwsAppStreamStack_disappears(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppStreamStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppStreamStack_complete(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigComplete(descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
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

func testAccAwsAppStreamStack_withTags(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigWithTags(descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
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

func testAccCheckAwsAppStreamStackExists(resourceName string, appStreamStack *appstream.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn
		resp, err := conn.DescribeStacks(&appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.Stacks) == 0 {
			return fmt.Errorf("appstream stack %q does not exist", rs.Primary.ID)
		}

		*appStreamStack = *resp.Stacks[0]

		return nil
	}
}

func testAccCheckAwsAppStreamStackDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_stack" {
			continue
		}

		resp, err := conn.DescribeStacks(&appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.Stacks) > 0 {
			return fmt.Errorf("appstream stack %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamStackConfigNameGenerated() string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {}
`)
}

func testAccAwsAppStreamStackConfigNamePrefix(stackName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name_prefix = %[1]q
}
`, stackName)
}

func testAccAwsAppStreamStackConfigComplete(description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  description = %[1]q
  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }
  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }
}
`, description)
}

func testAccAwsAppStreamStackConfigWithTags(description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  description = %[1]q
  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "FILE_UPLOAD"
    permission = "DISABLED"
  }
  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }
  user_settings {
    action     = "PRINTING_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_PASSWORD_SIGNIN"
    permission = "ENABLED"
  }
  user_settings {
    action     = "DOMAIN_SMART_CARD_SIGNIN"
    permission = "ENABLED"
  }
  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }
  tags = {
    Key = "value"
  }
}
`, description)
}
