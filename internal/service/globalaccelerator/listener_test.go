// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorListener_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_listener.example"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, t, resourceName),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "globalaccelerator", regexache.MustCompile(`accelerator/`+verify.UUIDRegexPattern+`/listener/[a-z0-9]{8}$`)),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "NONE"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "TCP"),
					resource.TestCheckResourceAttr(resourceName, "port_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_range.*", map[string]string{
						"from_port": "80",
						"to_port":   "81",
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

func TestAccGlobalAcceleratorListener_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_listener.example"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglobalaccelerator.ResourceListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorListener_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_listener.example"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rName),
			},
			{
				Config: testAccListenerConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "SOURCE_IP"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "UDP"),
					resource.TestCheckResourceAttr(resourceName, "port_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "port_range.*", map[string]string{
						"from_port": "443",
						"to_port":   "444",
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

func testAccCheckListenerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		_, err := tfglobalaccelerator.FindListenerByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckListenerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_listener" {
				continue
			}

			_, err := tfglobalaccelerator.FindListenerByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Listener %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccListenerConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.arn
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 81
  }
}
`, rName)
}

func testAccListenerConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = %[1]q
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.arn
  client_affinity = "SOURCE_IP"
  protocol        = "UDP"

  port_range {
    from_port = 443
    to_port   = 444
  }
}
`, rName)
}
