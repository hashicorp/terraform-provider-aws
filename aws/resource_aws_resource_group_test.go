package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSResourceGroup_importBasic(t *testing.T) {
	resourceName := "aws_resource_group.test"
	n := fmt.Sprintf("test-group-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSResourceGroupConfig_basic(n),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSResourceGroup_basic(t *testing.T) {
	resourceName := "aws_resource_group.test"
	n := fmt.Sprintf("test-group-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSResourceGroupConfig_basic(n),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAWSResourceGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource group name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).resourcegroupsconn

		_, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSResourceGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resource_group" "test" {
  name        = "%s"
  description = "Hello World"

  resource_query {
    query = <<JSON
{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Stage",
      "Values": ["Test"]
    }
  ]
}
JSON
  }
}
`, rName)
}
