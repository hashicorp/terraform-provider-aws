package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSVpcEndpointService_importBasic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_endpoint_service.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceBasicConfig(lbName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
