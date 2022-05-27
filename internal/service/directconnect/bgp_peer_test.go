package directconnect_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDirectConnectBGPPeer_basic(t *testing.T) {
	key := "DX_VIRTUAL_INTERFACE_ID"
	vifId := os.Getenv(key)
	if vifId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	bgpAsn := sdkacctest.RandIntRange(64512, 65534)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBGPPeerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBGPPeerConfig_basic(vifId, bgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBGPPeerExists("aws_dx_bgp_peer.foo"),
					resource.TestCheckResourceAttr("aws_dx_bgp_peer.foo", "address_family", "ipv6"),
				),
			},
		},
	})
}

func testAccCheckBGPPeerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_bgp_peer" {
			continue
		}
		input := &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.Attributes["virtual_interface_id"]),
		}

		resp, err := conn.DescribeVirtualInterfaces(input)
		if err != nil {
			return err
		}
		for _, peer := range resp.VirtualInterfaces[0].BgpPeers {
			if aws.StringValue(peer.AddressFamily) == rs.Primary.Attributes["address_family"] &&
				strconv.Itoa(int(aws.Int64Value(peer.Asn))) == rs.Primary.Attributes["bgp_asn"] &&
				aws.StringValue(peer.BgpPeerState) != directconnect.BGPPeerStateDeleted {
				return fmt.Errorf("[DESTROY ERROR] Dx BGP peer (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckBGPPeerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccBGPPeerConfig_basic(vifId string, bgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_bgp_peer" "foo" {
  virtual_interface_id = "%s"

  address_family = "ipv6"
  bgp_asn        = %d
}
`, vifId, bgpAsn)
}
