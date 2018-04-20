package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsBatchJobQueue(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_job_queue.test"
	datasourceName := "data.aws_batch_job_queue.by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsBatchJobQueueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsBatchJobQueueCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwsBatchJobQueueCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no data source called %s", datasourceName)
		}

		jobQueueRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"name",
			"state",
			"priority",
		}

		for _, attrName := range attrNames {
			if ds.Primary.Attributes[attrName] != jobQueueRs.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					ds.Primary.Attributes[attrName],
					jobQueueRs.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceAwsBatchJobQueueConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  name = "%[1]s"
  state = "ENABLED"
  priority = 1
  compute_environments = []
}

resource "aws_batch_job_queue" "wrong" {
  name = "%[1]s_wrong"
  state = "ENABLED"
  priority = 2
  compute_environments = []
}

data "aws_batch_job_queue" "by_name" {
  name = "${aws_batch_job_queue.test.name}"
}
`, rName)
}
