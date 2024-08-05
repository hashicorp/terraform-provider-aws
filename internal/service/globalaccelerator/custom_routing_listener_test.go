// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorCustomRoutingListener_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingListener
	resourceName := "aws_globalaccelerator_custom_routing_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingListenerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingListenerExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "port_range.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_range.*", map[string]string{
						"from_port": "443",
						"to_port":   "443",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_range.*", map[string]string{
						"from_port": "10000",
						"to_port":   "30000",
					}),
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

func TestAccGlobalAcceleratorCustomRoutingListener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingListener
	resourceName := "aws_globalaccelerator_custom_routing_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingListenerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingListenerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingListenerExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglobalaccelerator.ResourceCustomRoutingListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomRoutingListenerExists(ctx context.Context, n string, v *awstypes.CustomRoutingListener) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		output, err := tfglobalaccelerator.FindCustomRoutingListenerByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCustomRoutingListenerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_custom_routing_listener" {
				continue
			}

			_, err := tfglobalaccelerator.FindCustomRoutingListenerByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Custom Routing Listener %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCustomRoutingListenerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.id

  port_range {
    from_port = 443
    to_port   = 443
  }

  port_range {
    from_port = 10000
    to_port   = 30000
  }
}
`, rName)
}
