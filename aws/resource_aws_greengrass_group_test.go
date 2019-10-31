package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassGroup_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassGroupConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("group_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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

func TestAccAWSGreengrassGroup_GroupVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassGroupConfig_groupVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("group_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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

func testAccCheckAWSGreengrassGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_group" {
			continue
		}

		params := &greengrass.ListGroupsInput{}

		out, err := conn.ListGroups(params)
		if err != nil {
			return err
		}
		for _, groupInfo := range out.Groups {
			if *groupInfo.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Group to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassGroupConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_group" "test" {
  name = "group_%s"
}
`, rString)
}

func testAccAWSGreengrassGroupConfig_groupVersion(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_group" "test" {
  name = "group_%s"

  group_version {
	amzn_client_token = "token"
	connector_definition_version_arn = "arn:aws:greengrass:eu-west-1:123456789012:connector-definition-version:test_cd"
  }
}
`, rString)
}
