package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcsCluster_importBasic(t *testing.T) {
	resourceName := "aws_ecs_cluster.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsClusterDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSEcsCluster,
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
