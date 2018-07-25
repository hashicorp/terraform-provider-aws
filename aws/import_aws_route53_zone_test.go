package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRoute53Zone_importBasic(t *testing.T) {
	resourceName := "aws_route53_zone.main"

	rString := acctest.RandString(8)
	zoneName := fmt.Sprintf("%s.terraformtest.com", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53ZoneConfig(zoneName),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSRoute53Zone_importTwoVpcs(t *testing.T) {
	resourceName := "aws_route53_zone.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccRoute53ZoneAssociationConfig,
			},
			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Ignoring vpc_id, it can either be the one from the zone resource or
				// the one from the zone association resource
				ImportStateVerifyIgnore: []string{"force_destroy", "vpc_id"},
			},
		},
	})
}
