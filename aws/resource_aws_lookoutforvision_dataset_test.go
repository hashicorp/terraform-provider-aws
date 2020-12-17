package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lookoutforvision"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lookoutforvision/finder"
)

func init() {
	resource.AddTestSweepers("aws_lookoutforvision_dataset", &resource.Sweeper{
		Name: "aws_lookoutforvision_dataset",
		Dependencies: []string{
			"aws_lookoutforvision_project",
		},
		F: testSweepLookoutForVisionProjects,
	})
}

func TestAccAWSLookoutForVisionDataset_basic(t *testing.T) {
	var trainDataset lookoutforvision.DescribeDatasetOutput
	var testDataset lookoutforvision.DescribeDatasetOutput
	projectName := acctest.RandomWithPrefix("tf-acc-test-project")
	trainDatasetResourceName := "aws_lookoutforvision_dataset.train"
	testDatasetResourceName := "aws_lookoutforvision_dataset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLookoutForVisionDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLookoutForVisionDatasetBasicConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLookoutForVisionDatasetExists(trainDatasetResourceName, &trainDataset),
					testAccCheckAWSLookoutForVisionDatasetExists(testDatasetResourceName, &testDataset),
					resource.TestCheckResourceAttr(trainDatasetResourceName, "project", projectName),
					resource.TestCheckResourceAttr(testDatasetResourceName, "project", projectName),
					resource.TestCheckResourceAttr(trainDatasetResourceName, "dataset_type", "train"),
					resource.TestCheckResourceAttr(testDatasetResourceName, "dataset_type", "test"),
				),
			},
		},
	})
}

func testAccCheckAWSLookoutForVisionDatasetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lookoutforvisionconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lookoutforvision_dataset" {
			continue
		}

		projectName := rs.Primary.Attributes["project"]
		datasetType := rs.Primary.Attributes["dataset_type"]
		_, err := finder.DatasetByProjectAndType(conn, projectName, datasetType)
		if err == nil {
			return fmt.Errorf("Lookout for Vision dataset %q for project %q still exists", datasetType, projectName)
		}
	}

	return nil
}

func testAccCheckAWSLookoutForVisionDatasetExists(n string, dataset *lookoutforvision.DescribeDatasetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lookoutforvisionconn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		projectName := rs.Primary.Attributes["project"]
		datasetType := rs.Primary.Attributes["dataset_type"]

		if projectName == "" {
			return fmt.Errorf("No Lookout for Vision project is set")
		}

		if datasetType == "" {
			return fmt.Errorf("No Lookout for Vision dataset type is set")
		}

		resp, err := finder.DatasetByProjectAndType(conn, projectName, datasetType)
		if err != nil {
			return err
		}
		*dataset = *resp

		return nil
	}
}

func testAccAWSLookoutForVisionDatasetBasicConfig(projectName string) string {
	return fmt.Sprintf(`
resource "aws_lookoutforvision_project" "demo" {
  name = %[1]q
}

resource "aws_lookoutforvision_dataset" "train" {
  project = aws_lookoutforvision_project.demo.name
  dataset_type = "train"
}

resource "aws_lookoutforvision_dataset" "test" {
  project = aws_lookoutforvision_project.demo.name
  dataset_type = "test"
}
`, projectName)
}
