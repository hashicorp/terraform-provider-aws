package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSVpcEndpointConnectionNotification_importBasic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_endpoint_connection_notification.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointConnectionNotificationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConnectionNotificationBasicConfig(lbName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
