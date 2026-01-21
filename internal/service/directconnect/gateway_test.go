// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DirectConnectGateway
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "directconnect", "dx-gateway/{id}"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, gateway_CheckID(&v)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
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
	var v awstypes.DirectConnectGateway
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectGateway_complex(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DirectConnectGateway
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
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
	var v awstypes.DirectConnectGateway
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccGatewayConfig_basic(rName2, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccDirectConnectGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DirectConnectGateway
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
				Config: testAccGatewayConfig_tags1(rName, rBgpAsn, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "directconnect", "dx-gateway/{id}"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, gateway_CheckID(&v)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGatewayConfig_tags2(rName, rBgpAsn, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "directconnect", "dx-gateway/{id}"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, gateway_CheckID(&v)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGatewayConfig_tags1(rName, rBgpAsn, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "directconnect", "dx-gateway/{id}"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, gateway_CheckID(&v)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGatewayConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARNFormat(ctx, resourceName, names.AttrARN, "directconnect", "dx-gateway/{id}"),
					resource.TestCheckResourceAttrWith(resourceName, names.AttrID, gateway_CheckID(&v)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccCheckGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_gateway" {
				continue
			}

			_, err := tfdirectconnect.FindGatewayByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckGatewayExists(ctx context.Context, n string, v *awstypes.DirectConnectGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		output, err := tfdirectconnect.FindGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func gateway_CheckID(v *awstypes.DirectConnectGateway) func(string) error {
	return func(s string) error {
		expected := aws.ToString(v.DirectConnectGatewayId)
		if s != expected {
			return fmt.Errorf("expected value %q for String check, got: %q", expected, s)
		}
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

func testAccGatewayConfig_tags1(rName string, rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1)
}

func testAccGatewayConfig_tags2(rName string, rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}
