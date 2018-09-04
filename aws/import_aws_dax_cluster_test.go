package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDAXCluster_importBasic(t *testing.T) {
	resourceName := "aws_dax_cluster.test"
	rString := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDAXClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDAXClusterConfig(rString),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
