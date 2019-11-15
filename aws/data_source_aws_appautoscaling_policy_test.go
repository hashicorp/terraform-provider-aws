package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSAppautoscalingPolicyDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_appautoscaling_policy.test"
	resourceName := "aws_appautoscaling_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAppautoscalingPolicyDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_type", dataSourceName, "policy_type"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", dataSourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", dataSourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", dataSourceName, "service_namespace"),
				),
			},
		},
	})
}

var testAppautoscalingPolicyDataSourceConfig = fmt.Sprintf(`
resource "aws_appautoscaling_policy" "test" {
	name               = "tf-autoscaling-test-policy"
	policy_type        = "TargetTrackingScaling"
	resource_id        = "service/clusterName/serviceName"
	scalable_dimension = "ecs:service:DesiredCount"
	service_namespace  = "ecs"
}
`)
