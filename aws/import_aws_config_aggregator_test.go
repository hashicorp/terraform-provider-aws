package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccConfigAggregator_import(t *testing.T) {
	resourceName := "aws_config_aggregator.example"
	rString := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigAggregatorDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccConfigAggregatorConfig_account(rString),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
