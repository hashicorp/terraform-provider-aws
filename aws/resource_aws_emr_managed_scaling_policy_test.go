package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsEmrManagedScalingPolicy_basic(t *testing.T) {
	resourceName := "aws_emr_managed_scaling_policy.testpolicy"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEmrManagedScalingPolicyDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccAWSEmrManagedScalingPolicy_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEmrManagedScalingPolicyExists(resourceName),
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

func testAccAWSEmrManagedScalingPolicy_basic(r int) string {
	return fmt.Sprintf(testAccAWSEmrInstanceGroupBase+`
resource "aws_emr_managed_scaling_policy" "testpolicy" {
  cluster_id = aws_emr_cluster.tf-test-cluster.id
  compute_limits {
    unit_type                       = "Instances"
    minimum_capacity_units          = 1
    maximum_capacity_units          = 2
    maximum_ondemand_capacity_units = 2
    maximum_core_capacity_units     = 2
  }
}
`, r)
}

func testAccCheckAWSEmrManagedScalingPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Managed Scaling Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).emrconn
		resp, err := conn.GetManagedScalingPolicy(&emr.GetManagedScalingPolicyInput{
			ClusterId: aws.String(rs.Primary.Attributes["cluster_id"]),
		})
		if err != nil {
			return err
		}

		if resp.ManagedScalingPolicy == nil {
			return fmt.Errorf("EMR Managed Scaling Policy is empty which shouldn't happen")
		}
		return nil
	}
}

func testAccCheckAWSEmrManagedScalingPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).emrconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_managed_scaling_policy" {
			continue
		}

		resp, err := conn.DescribeSecurityConfiguration(&emr.DescribeSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, "InvalidRequestException", "does not exist") {
			return nil
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("Error: EMR Managed Scaling Policy still exists")
		}

		return nil
	}

	return nil
}
