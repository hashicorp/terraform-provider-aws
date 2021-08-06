package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsAppStreamStack_basic(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigBasic(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
				),
			},
			{
				Config:            testAccAwsAppStreamStackConfigBasic(stackName),
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
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigBasic(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppStreamStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppStreamStack_withTags(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a fleet"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigWithTags(stackName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
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
			return fmt.Errorf("appstream fleet %q does not exist", rs.Primary.ID)
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
			return fmt.Errorf("appstream fleet %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamStackConfigBasic(stackName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test_fleet" {
  name       = %[1]q
}
`, stackName)
}

func testAccAwsAppStreamStackConfigWithTags(stackName, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name       = %[1]q
  compute_capacity {
    desired_instances = 1
  }
  description                         = %[2]q
  display_name                        = %[1]q
  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }
  user_settings {
    action  = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    enabled = true
  }
  user_settings {
    action  = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    enabled = true
  }
  user_settings {
    action  = "FILE_UPLOAD"
    enabled = true
  }
  user_settings {
    action  = "FILE_DOWNLOAD"
    enabled = true
  }
  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }
  tags = {
    Key = "value"
  }
}
`, stackName, description)
}
