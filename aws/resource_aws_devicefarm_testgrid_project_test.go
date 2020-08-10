package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDeviceFarmTestGridProject_basic(t *testing.T) {
	var proj devicefarm.TestGridProject
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_testgrid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmTestGridProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmTestGridProjectExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "devicefarm", regexp.MustCompile(`testgrid-project:.+`)),
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

func TestAccAWSDeviceFarmTestGridProject_disappears(t *testing.T) {
	var proj devicefarm.TestGridProject
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_devicefarm_testgrid_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDeviceFarmTestGridProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeviceFarmTestGridProjectConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeviceFarmTestGridProjectExists(resourceName, &proj),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDevicefarmTestgridProject(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDeviceFarmTestGridProjectExists(n string, v *devicefarm.TestGridProject) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).devicefarmconn
		resp, err := conn.GetTestGridProject(
			&devicefarm.GetTestGridProjectInput{ProjectArn: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if resp.TestGridProject == nil {
			return fmt.Errorf("DeviceFarmTestGridProject not found")
		}

		*v = *resp.TestGridProject

		return nil
	}
}

func testAccCheckDeviceFarmTestGridProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).devicefarmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_devicefarm_testgrid_project" {
			continue
		}

		// Try to find the resource
		resp, err := conn.GetTestGridProject(
			&devicefarm.GetTestGridProjectInput{ProjectArn: aws.String(rs.Primary.ID)})
		if err == nil {
			if resp.TestGridProject != nil {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		if isAWSErr(err, devicefarm.ErrCodeNotFoundException, "") {
			return nil
		}
	}

	return nil
}

func testAccDeviceFarmTestGridProjectConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_devicefarm_testgrid_project" "test" {
  name = %[1]q
}
`, rName)
}
