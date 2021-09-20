package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccDataSourceAwsSnsTopic_basic(t *testing.T) {
	resourceName := "aws_sns_topic.test"
	datasourceName := "data.aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, sns.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSnsTopicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSnsTopicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_sns_topic" "test" {
  name = aws_sns_topic.test.name
}
`, rName)
}
