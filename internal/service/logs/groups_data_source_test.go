package logs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLogsGroupsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "data.aws_cloudwatch_log_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test1", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test1", "name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test2", "name"),
				),
			},
		},
	})
}

func TestAccLogsGroupsDataSource_noPrefix(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "data.aws_cloudwatch_log_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupsDataSourceConfig_noPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test1", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "arns.*", "aws_cloudwatch_log_group.test2", "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test1", "name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "log_group_names.*", "aws_cloudwatch_log_group.test2", "name"),
				),
			},
		},
	})
}

func testAccGroupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test1" {
  name = "%[1]s/1"
}

resource aws_cloudwatch_log_group "test2" {
  name = "%[1]s/2"
}

data aws_cloudwatch_log_groups "test" {
  log_group_name_prefix = %[1]q

  depends_on = [aws_cloudwatch_log_group.test1,aws_cloudwatch_log_group.test2]
}
`, rName)
}

func testAccGroupsDataSourceConfig_noPrefix(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test1" {
  name = "%[1]s/1"
}

resource aws_cloudwatch_log_group "test2" {
  name = "%[1]s/2"
}

data aws_cloudwatch_log_groups "test" {
  depends_on = [aws_cloudwatch_log_group.test1,aws_cloudwatch_log_group.test2]
}
`, rName)
}
