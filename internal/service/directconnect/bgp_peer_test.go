// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectBGPPeer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	vifID := acctest.SkipIfEnvVarNotSet(t, "DX_VIRTUAL_INTERFACE_ID")
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	resourceName := "aws_dx_bgp_peer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBGPPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBGPPeerConfig_basic(vifID, bgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBGPPeerExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "ipv6"),
				),
			},
		},
	})
}

func TestAccDirectConnectBGPPeer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	vifID := acctest.SkipIfEnvVarNotSet(t, "DX_VIRTUAL_INTERFACE_ID")
	bgpAsn := acctest.RandIntRange(t, 64512, 65534)
	resourceName := "aws_dx_bgp_peer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBGPPeerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBGPPeerConfig_basic(vifID, bgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBGPPeerExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceBGPPeer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBGPPeerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_bgp_peer" {
				continue
			}

			_, err := tfdirectconnect.FindBGPPeerByThreePartKey(ctx, conn, rs.Primary.Attributes["virtual_interface_id"], awstypes.AddressFamily(rs.Primary.Attributes["address_family"]), flex.StringValueToInt32Value(rs.Primary.Attributes["bgp_asn"]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect BGP Peer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBGPPeerExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		_, err := tfdirectconnect.FindBGPPeerByThreePartKey(ctx, conn, rs.Primary.Attributes["virtual_interface_id"], awstypes.AddressFamily(rs.Primary.Attributes["address_family"]), flex.StringValueToInt32Value(rs.Primary.Attributes["bgp_asn"]))

		return err
	}
}

func testAccBGPPeerConfig_basic(vifID string, bgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_bgp_peer" "test" {
  virtual_interface_id = %[1]q

  address_family = "ipv6"
  bgp_asn        = %[2]d
}
`, vifID, bgpAsn)
}
