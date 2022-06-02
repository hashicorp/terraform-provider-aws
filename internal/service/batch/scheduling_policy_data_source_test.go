package batch_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/batch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccBatchSchedulingPolicyDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_scheduling_policy.test"
	datasourceName := "data.aws_batch_scheduling_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBatchSchedulingPolicyDataSourceConfig_Basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "fair_share_policy.#", resourceName, "fair_share_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "fair_share_policy.0.compute_reservation", resourceName, "fair_share_policy.0.compute_reservation"),
					resource.TestCheckResourceAttrPair(datasourceName, "fair_share_policy.0.share_decay_seconds", resourceName, "fair_share_policy.0.share_decay_seconds"),
					resource.TestCheckResourceAttrPair(datasourceName, "fair_share_policy.0.share_distribution.#", resourceName, "fair_share_policy.0.share_distribution.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccBatchSchedulingPolicyDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test" {
  name = %[1]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }

  tags = {
    "Name" = "Test Batch Scheduling Policy"
  }
}
`, rName)
}

func testAccBatchSchedulingPolicyDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(testAccBatchSchedulingPolicyDataSourceConfig(rName) + `
data "aws_batch_scheduling_policy" "test" {
  arn = aws_batch_scheduling_policy.test.arn
}
`)
}
