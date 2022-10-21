package networkmanager_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/networkmanager"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkmanager "github.com/hashicorp/terraform-provider-aws/internal/service/networkmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNetworkManagerVpnAttachment_basic(t *testing.T) {
	var v networkmanager.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_vpn_attachment.test"
	vpnResourceName := "aws_vpn_connection.test"
	coreNetwork := "awscc_networkmanager_core_network.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bgpASN := 65000
	vpnIP := "172.0.0.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVpnAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnAttachmentConfig_basic(rName, bgpASN, vpnIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpnAttachmentExists(resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "networkmanager", regexp.MustCompile(`attachment/.+`)),
					resource.TestCheckResourceAttr(resourceName, "attachment_policy_rule_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachment_type", "SITE_TO_SITE_VPN"),
					resource.TestCheckResourceAttrPair(resourceName, "core_network_arn", coreNetwork, "core_network_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "core_network_id"),
					resource.TestCheckResourceAttr(resourceName, "edge_location", acctest.Region()),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", vpnResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "segment_name", "shared"),
					resource.TestCheckResourceAttrSet(resourceName, "state"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_arn", vpnResourceName, "arn"),
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

func TestAccNetworkManagerVpnAttachment_disappears(t *testing.T) {
	var v networkmanager.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_vpn_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bgpASN := 65001
	vpnIP := "172.0.0.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVpnAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnAttachmentConfig_basic(rName, bgpASN, vpnIP),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnAttachmentExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkmanager.ResourceVpnAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkManagerVpnAttachment_tags(t *testing.T) {
	var v networkmanager.SiteToSiteVpnAttachment
	resourceName := "aws_networkmanager_vpn_attachment.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"awscc": {
			Source:            "hashicorp/awscc",
			VersionConstraint: "0.29.0",
		},
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bgpASN := 65002
	vpnIP := "172.0.0.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, networkmanager.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckVpnAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnAttachmentConfig_tags1(rName, vpnIP, "segment", "shared", bgpASN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
				),
			},
			{
				Config: testAccVpnAttachmentConfig_tags2(rName, vpnIP, "segment", "shared", "Name", "test", bgpASN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
				),
			},
			{
				Config: testAccVpnAttachmentConfig_tags1(rName, vpnIP, "segment", "shared", bgpASN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnAttachmentExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.segment", "shared"),
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

func testAccCheckVpnAttachmentExists(n string, v *networkmanager.SiteToSiteVpnAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Manager VPN Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

		output, err := tfnetworkmanager.FindVpnAttachmentByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVpnAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_vpn_attachment" {
			continue
		}

		_, err := tfnetworkmanager.FindVpnAttachmentByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Network Manager VPN Attachment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVpnAttachmentConfig_base(rName string, bgpASN int, vpnIP string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = %[3]q
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "test" {
    customer_gateway_id = aws_customer_gateway.test.id
    type                = "ipsec.1"

    tags = {
      Name = %[1]q
    }
  }

resource "aws_networkmanager_global_network" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "awscc_networkmanager_core_network" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  policy_document   = jsonencode(jsondecode(data.aws_networkmanager_core_network_policy_document.test.json))
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    vpn_ecmp_support = false
    asn_ranges       = ["64512-64555"]
    edge_locations {
      location = data.aws_region.current.name
      asn      = 64512
    }
  }

  segments {
    name                          = "shared"
    description                   = "SegmentForSharedServices"
    require_attachment_acceptance = true
  }

  segment_actions {
    action     = "share"
    mode       = "attachment-route"
    segment    = "shared"
      share_with = ["*"]
  }

  attachment_policies {
    rule_number     = 1
    condition_logic = "or"

      conditions {
        type     = "tag-value"
        operator = "equals"
        key      = "segment"
        value    = "shared"
      }

      action {
        association_method = "constant"
        segment            = "shared"
      }
  }
 }
`, rName, bgpASN, vpnIP))
}

func testAccVpnAttachmentConfig_basic(rName string, bgpASN int, vpnIP string) string {
	return acctest.ConfigCompose(testAccVpnAttachmentConfig_base(rName, bgpASN, vpnIP), `
resource "aws_networkmanager_vpn_attachment" "test" {
  core_network_id = awscc_networkmanager_core_network.test.id
  vpn_arn         = aws_vpn_connection.test.arn
	tags = {
		segment = "shared"
	}
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpn_attachment.test.id
  attachment_type = aws_networkmanager_vpn_attachment.test.attachment_type
}
`)
}

func testAccVpnAttachmentConfig_tags1(rName, vpnIP, tagKey1, tagValue1 string, bgpASN int) string {
	return acctest.ConfigCompose(testAccVpnAttachmentConfig_base(rName, bgpASN, vpnIP), fmt.Sprintf(`
resource "aws_networkmanager_vpn_attachment" "test" {
  core_network_id = awscc_networkmanager_core_network.test.id
  vpn_arn         = aws_vpn_connection.test.arn

  tags = {
    %[1]q = %[2]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpn_attachment.test.id
  attachment_type = aws_networkmanager_vpn_attachment.test.attachment_type
}
`, tagKey1, tagValue1))
}

func testAccVpnAttachmentConfig_tags2(rName, vpnIP, tagKey1, tagValue1, tagKey2, tagValue2 string, bgpASN int) string {
	return acctest.ConfigCompose(testAccVpnAttachmentConfig_base(rName, bgpASN, vpnIP), fmt.Sprintf(`
resource "aws_networkmanager_vpn_attachment" "test" {
  core_network_id = awscc_networkmanager_core_network.test.id
  vpn_arn         = aws_vpn_connection.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}

resource "aws_networkmanager_attachment_accepter" "test" {
  attachment_id   = aws_networkmanager_vpn_attachment.test.id
  attachment_type = aws_networkmanager_vpn_attachment.test.attachment_type
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
