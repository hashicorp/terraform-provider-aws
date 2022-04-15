package cloudwatchevidently_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccProject_basic(t *testing.T) {
	var project cloudwatchevidently.GetProjectOutput

	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_cloudwatchevidently_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_basic(rName, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectExists(resourceName, &project),
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProjectConfig_basic(rName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "active_experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "active_launch_count"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrSet(resourceName, "experiment_count"),
					resource.TestCheckResourceAttrSet(resourceName, "feature_count"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrSet(resourceName, "launch_count"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Test Project"),
				),
			},
		},
	})
}

func testAccCheckProjectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEvidentlyConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatchevidently_project" {
			continue
		}

		input := &cloudwatchevidently.GetProjectInput{
			Project: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetProject(input)

		if err == nil {
			if aws.StringValue(resp.Project.Arn) == rs.Primary.ID {
				return fmt.Errorf("Project '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckProjectExists(name string, project *cloudwatchevidently.GetProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEvidentlyConn
		input := &cloudwatchevidently.GetProjectInput{
			Project: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetProject(input)

		if err != nil {
			return err
		}

		*project = *resp

		return nil
	}
}

func testAccProjectConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchevidently_project" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    "Key1" = "Test Project"
  }
}
`, rName, description)
}
