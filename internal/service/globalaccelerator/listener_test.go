package globalaccelerator_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlobalAcceleratorListener_basic(t *testing.T) {
	resourceName := "aws_globalaccelerator_listener.example"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TCP"),
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
	resourceName := "aws_globalaccelerator_listener.example"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfglobalaccelerator.ResourceListener(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorListener_update(t *testing.T) {
	resourceName := "aws_globalaccelerator_listener.example"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerConfig_basic(rInt),
			},
			{
				Config: testAccListenerConfig_update(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "client_affinity", "SOURCE_IP"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "UDP"),
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

func testAccCheckListenerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		_, err := tfglobalaccelerator.FindListenerByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckListenerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_globalaccelerator_listener" {
			continue
		}

		_, err := tfglobalaccelerator.FindListenerByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Global Accelerator Accelerator %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccListenerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  protocol        = "TCP"

  port_range {
    from_port = 80
    to_port   = 81
  }
}
`, rInt)
}

func testAccListenerConfig_update(rInt int) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "example" {
  name            = "tf-%d"
  ip_address_type = "IPV4"
  enabled         = false
}

resource "aws_globalaccelerator_listener" "example" {
  accelerator_arn = aws_globalaccelerator_accelerator.example.id
  client_affinity = "SOURCE_IP"
  protocol        = "UDP"

  port_range {
    from_port = 443
    to_port   = 444
  }
}
`, rInt)
}
