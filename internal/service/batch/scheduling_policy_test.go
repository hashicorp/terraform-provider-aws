package batch_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
)

func testAccCheckBatchSchedulingPolicyExists(n string, sp *batch.SchedulingPolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		log.Printf("State: %#v", s.RootModule().Resources)
		if !ok {
			return fmt.Errorf("Batch Scheduling Policy not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Scheduling Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn
		schedulingPolicy, err := GetSchedulingPolicyNoContext(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if schedulingPolicy == nil {
			return fmt.Errorf("Batch Scheduling Polic not found: %s", n)
		}
		*sp = *schedulingPolicy

		return nil
	}
}


func GetSchedulingPolicyNoContext(conn *batch.Batch, arn string) (*batch.SchedulingPolicyDetail, error) {
	resp, err := conn.DescribeSchedulingPolicies(&batch.DescribeSchedulingPoliciesInput{
		Arns: []*string{aws.String(arn)},
	})
	if err != nil {
		return nil, err
	}

	numSchedulingPolicies := len(resp.SchedulingPolicies)
	switch {
	case numSchedulingPolicies == 0:
		log.Printf("[DEBUG] Scheduling Policy %q is already gone", arn)
		return nil, nil
	case numSchedulingPolicies == 1:
		return resp.SchedulingPolicies[0], nil
	case numSchedulingPolicies > 1:
		return nil, fmt.Errorf("Multiple Scheduling Policy with arn %s", arn)
	}
	return nil, nil
}

func testAccBatchSchedulingPolicyConfigBasic(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test" {
  name = %[1]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }
  }

  tags = {
    "Name" = "Test Batch Scheduling Policy"
  }
}
`, rName))
}

func testAccBatchSchedulingPolicyConfigBasic2(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
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
`, rName))
}
