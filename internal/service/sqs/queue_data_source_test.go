package sqs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSQSQueueDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_sqs_queue.test"
	datasourceName := "data.aws_sqs_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccQueueCheckDataSource(datasourceName, resourceName),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccSQSQueueDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_sqs_queue.test"
	datasourceName := "data.aws_sqs_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccQueueCheckDataSource(datasourceName, resourceName),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Empty", ""),
				),
			},
		},
	})
}

func testAccQueueCheckDataSource(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		sqsQueueRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"name",
		}

		for _, attrName := range attrNames {
			if rs.Primary.Attributes[attrName] != sqsQueueRs.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					rs.Primary.Attributes[attrName],
					sqsQueueRs.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccQueueDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "wrong" {
  name = "%[1]s_wrong"
}

resource "aws_sqs_queue" "test" {
  name = "%[1]s"
}

data "aws_sqs_queue" "by_name" {
  name = aws_sqs_queue.test.name
}
`, rName)
}

func testAccQueueDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = "%[1]s"

  tags = {
    Environment = "Production"
    Foo         = "Bar"
    Empty       = ""
  }
}

data "aws_sqs_queue" "by_name" {
  name = aws_sqs_queue.test.name
}
`, rName)
}
