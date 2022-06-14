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

func TestAccBatchSchedulingPolicy_basic(t *testing.T) {
	var schedulingPolicy1 batch.SchedulingPolicyDetail
	resourceName := "aws_batch_scheduling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchedulingPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulingPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchedulingPolicyExists(resourceName, &schedulingPolicy1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.compute_reservation", "1"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.share_decay_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.share_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// add one more share_distribution block
				Config: testAccSchedulingPolicyConfig_basic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchedulingPolicyExists(resourceName, &schedulingPolicy1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.compute_reservation", "1"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.share_decay_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "fair_share_policy.0.share_distribution.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func TestAccBatchSchedulingPolicy_disappears(t *testing.T) {
	var schedulingPolicy1 batch.SchedulingPolicyDetail
	resourceName := "aws_batch_scheduling_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSchedulingPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulingPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchedulingPolicyExists(resourceName, &schedulingPolicy1),
					acctest.CheckResourceDisappears(acctest.Provider, tfbatch.ResourceSchedulingPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSchedulingPolicyExists(n string, sp *batch.SchedulingPolicyDetail) resource.TestCheckFunc {
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

func testAccCheckSchedulingPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_scheduling_policy" {
			continue
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn
		sp, err := GetSchedulingPolicyNoContext(conn, rs.Primary.ID)
		if err == nil {
			if sp != nil {
				return fmt.Errorf("Error: Scheduling Policy still exists")
			}
		}
		return nil
	}
	return nil
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

func testAccSchedulingPolicyConfig_basic(rName string) string {
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

func testAccSchedulingPolicyConfig_basic2(rName string) string {
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
