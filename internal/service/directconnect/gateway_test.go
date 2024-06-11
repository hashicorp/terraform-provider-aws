// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
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
	ctx := acctest.Context(t)
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdirectconnect.ResourceGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectGateway_complex(t *testing.T) {
	ctx := acctest.Context(t)
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_associationMultiVPNSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerAccountID),
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

func TestAccDirectConnectGateway_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v directconnect.Gateway
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName1, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccGatewayConfig_basic(rName2, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func testAccCheckGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_gateway" {
				continue
			}

			_, err := tfdirectconnect.FindGatewayByID(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckGatewayExists(ctx context.Context, name string, v *directconnect.Gateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

		output, err := tfdirectconnect.FindGatewayByID(ctx, conn, rs.Primary.ID)

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
