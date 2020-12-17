package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lookoutforvision"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lookoutforvision/finder"
)

func init() {
	resource.AddTestSweepers("aws_lookoutforvision_project", &resource.Sweeper{
		Name: "aws_lookoutforvision_project",
		F:    testSweepLookoutForVisionProjects,
	})
}

func testSweepLookoutForVisionProjects(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).lookoutforvisionconn

	err = conn.ListProjectsPages(&lookoutforvision.ListProjectsInput{},
		func(page *lookoutforvision.ListProjectsOutput, lastPage bool) bool {
			for _, project := range page.Projects {
				name := aws.StringValue(project.ProjectName)

				input := &lookoutforvision.DeleteProjectInput{
					ProjectName: project.ProjectName,
				}

				log.Printf("[INFO] Deleting Lookout for Vision Project: %s", name)
				if _, err := conn.DeleteProject(input); err != nil {
					log.Printf("[ERROR] Error deleting Lookout for Vision Project (%s): %s", name, err)
					continue
				}
			}

			return !lastPage
		})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Lookout for Vision Project sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving Lookout for Vision Projects: %w", err)
	}

	return nil
}

func TestAccAWSLookoutForVisionProject_basic(t *testing.T) {
	var project lookoutforvision.DescribeProjectOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lookoutforvision_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLookoutForVisionProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLookoutForVisionProjectBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLookoutForVisionProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "lookoutvision", fmt.Sprintf("project/%s", rName)),
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

func testAccCheckAWSLookoutForVisionProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lookoutforvisionconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lookoutforvision_project" {
			continue
		}

		project, err := finder.ProjectByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(project.ProjectDescription.ProjectName) == rs.Primary.ID {
			return fmt.Errorf("Lookout for Vision project %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSLookoutForVisionProjectExists(n string, project *lookoutforvision.DescribeProjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lookout for Vision Project ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lookoutforvisionconn
		resp, err := finder.ProjectByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*project = *resp

		return nil
	}
}

func testAccAWSLookoutForVisionProjectBasicConfig(pName string) string {
	return fmt.Sprintf(`
resource "aws_lookoutforvision_project" "test" {
  name = %[1]q
}
`, pName)
}
