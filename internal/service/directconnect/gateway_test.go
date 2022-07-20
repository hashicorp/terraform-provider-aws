package directconnect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDirectConnectGateway_basic(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
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

func TestAccDirectConnectGateway_disappears(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfdirectconnect.ResourceGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectGateway_complex(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_associationMultiVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
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

func testAccCheckGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway" {
			continue
		}

		_, err := tfdirectconnect.FindGatewayByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Gateway %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckGatewayExists(name string, v *directconnect.Gateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		output, err := tfdirectconnect.FindGatewayByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccGatewayConfig_basic(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}
`, rName, rBgpAsn)
}
