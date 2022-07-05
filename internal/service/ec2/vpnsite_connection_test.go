package ec2_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type TunnelOptions struct {
	psk                        string
	tunnelCidr                 string
	dpdTimeoutAction           string
	dpdTimeoutSeconds          int
	ikeVersions                string
	phase1DhGroupNumbers       string
	phase1EncryptionAlgorithms string
	phase1IntegrityAlgorithms  string
	phase1LifetimeSeconds      int
	phase2DhGroupNumbers       string
	phase2EncryptionAlgorithms string
	phase2IntegrityAlgorithms  string
	phase2LifetimeSeconds      int
	rekeyFuzzPercentage        int
	rekeyMarginTimeSeconds     int
	replayWindowSize           int
	startupAction              string
}

func TestXmlConfigToTunnelInfo(t *testing.T) {
	testCases := []struct {
		Name                  string
		XML                   string
		Tunnel1PreSharedKey   string
		Tunnel1InsideCidr     string
		Tunnel1InsideIpv6Cidr string
		ExpectError           bool
		ExpectTunnelInfo      tfec2.TunnelInfo
	}{
		{
			Name: "outside address sort",
			XML:  testAccVPNTunnelInfoXML,
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "1.1.1.1",
				Tunnel1BGPASN:           "1111",
				Tunnel1BGPHoldTime:      31,
				Tunnel1CgwInsideAddress: "169.254.11.1",
				Tunnel1PreSharedKey:     "FIRST_KEY",
				Tunnel1VgwInsideAddress: "168.254.11.2",
				Tunnel2Address:          "2.2.2.2",
				Tunnel2BGPASN:           "2222",
				Tunnel2BGPHoldTime:      32,
				Tunnel2CgwInsideAddress: "169.254.12.1",
				Tunnel2PreSharedKey:     "SECOND_KEY",
				Tunnel2VgwInsideAddress: "169.254.12.2",
			},
		},
		{
			Name:                "Tunnel1PreSharedKey",
			XML:                 testAccVPNTunnelInfoXML,
			Tunnel1PreSharedKey: "SECOND_KEY",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		{
			Name:              "Tunnel1InsideCidr",
			XML:               testAccVPNTunnelInfoXML,
			Tunnel1InsideCidr: "169.254.12.0/30",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		// IPv6 logic is equivalent to IPv4, so we can reuse configuration, expected, etc.
		{
			Name:                  "Tunnel1InsideIpv6Cidr",
			XML:                   testAccVPNTunnelInfoXML,
			Tunnel1InsideIpv6Cidr: "169.254.12.1",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			tunnelInfo, err := tfec2.CustomerGatewayConfigurationToTunnelInfo(testCase.XML, testCase.Tunnel1PreSharedKey, testCase.Tunnel1InsideCidr, testCase.Tunnel1InsideIpv6Cidr)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error, got none")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("expected no error, got: %s", err)
			}

			if actual, expected := *tunnelInfo, testCase.ExpectTunnelInfo; !reflect.DeepEqual(actual, expected) { // nosemgrep: prefer-aws-go-sdk-pointer-conversion-assignment
				t.Errorf("expected tfec2.TunnelInfo:\n%+v\n\ngot:\n%+v\n\n", expected, actual)
			}
		})
	}
}

func TestAccSiteVPNConnection_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpn-connection/vpn-.+`)),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_attachment_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "customer_gateway_configuration"),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
					resource.TestCheckResourceAttr(resourceName, "local_ipv4_network_cidr", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "local_ipv6_network_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "outside_ip_address_type", "PublicIpv4"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv4_network_cidr", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv6_network_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "routes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_ike_versions"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_inside_cidr"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_preshared_key"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", ""),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_ike_versions"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_inside_cidr"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_lifetime_seconds", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms"),
					resource.TestCheckNoResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_preshared_key"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", ""),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_vgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel_inside_ip_version", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "vgw_telemetry.#", "2"),
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

func TestAccSiteVPNConnection_transitGatewayID(t *testing.T) {
	var vpn ec2.VpnConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_transitGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestMatchResourceAttr(resourceName, "transit_gateway_attachment_id", regexp.MustCompile(`tgw-attach-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
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

func TestAccSiteVPNConnection_tunnel1InsideCIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_tunnel1InsideCIDR(rName, rBgpAsn, "169.254.8.0/30", "169.254.9.0/30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccSiteVPNConnection_tunnel1InsideIPv6CIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_tunnel1InsideIPv6CIDR(rName, rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccSiteVPNConnection_tunnel1PreSharedKey(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_tunnel1PresharedKey(rName, rBgpAsn, "tunnel1presharedkey", "tunnel2presharedkey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "tunnel1presharedkey"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "tunnel2presharedkey"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccSiteVPNConnection_tunnelOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	badCidrRangeErr := regexp.MustCompile(`expected \w+ to not be any of \[[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/30\s?]+\]`)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	tunnel1 := TunnelOptions{
		psk:                        "12345678",
		tunnelCidr:                 "169.254.8.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	tunnel2 := TunnelOptions{
		psk:                        "abcdefgh",
		tunnelCidr:                 "169.254.9.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "not-a-cidr"),
				ExpectError: regexp.MustCompile(`invalid CIDR address: not-a-cidr`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.254.0/31"),
				ExpectError: regexp.MustCompile(`expected "\w+" to contain a network Value with between 30 and 30 significant bits`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "172.16.0.0/30"),
				ExpectError: regexp.MustCompile(`must be within 169.254.0.0/16`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.0.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.1.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.2.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.3.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.4.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.5.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "12345678", "169.254.169.252/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "1234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, sdkacctest.RandStringFromCharSet(65, sdkacctest.CharSetAlpha), "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "01234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`cannot start with zero character`),
			},
			{
				Config:      testAccSiteVPNConnectionConfig_singleTunnelOptions(rName, rBgpAsn, "1234567!", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`can only contain alphanumeric, period and underscore characters`),
			},
			{
				Config: testAccSiteVPNConnectionConfig_tunnelOptions(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),

					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),

					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

// TestAccSiteVPNConnection_tunnelOptionsLesser tests less algorithms such as those supported in GovCloud.
func TestAccSiteVPNConnection_tunnelOptionsLesser(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2, vpn3, vpn4, vpn5 ec2.VpnConnection

	tunnel1 := TunnelOptions{
		psk:                        "12345678",
		tunnelCidr:                 "169.254.8.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "14, 15, 16, 17, 18, 19, 20, 21",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES128-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	tunnel2 := TunnelOptions{
		psk:                        "abcdefgh",
		tunnelCidr:                 "169.254.9.0/30",
		dpdTimeoutAction:           "none",
		dpdTimeoutSeconds:          45,
		ikeVersions:                "\"ikev2\"",
		phase1DhGroupNumbers:       "18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\"",
		phase1IntegrityAlgorithms:  "\"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      1800,
		phase2DhGroupNumbers:       "15, 16, 17, 18, 19, 20, 21, 22",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      1200,
		rekeyFuzzPercentage:        90,
		rekeyMarginTimeSeconds:     360,
		replayWindowSize:           512,
		startupAction:              "start",
	}

	// inside_cidr can't be updated in-place.
	tunnel1Updated := tunnel2
	tunnel1Updated.tunnelCidr = tunnel1.tunnelCidr

	tunnel2Updated := tunnel1
	tunnel2Updated.tunnelCidr = tunnel2.tunnelCidr

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_tunnelOptions(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", "clear"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_ike_versions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "14"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "28800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "540"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", "add"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "45"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_ike_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_lifetime_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "1200"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "360"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", "start"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_vgw_inside_address"),
				),
			},
			// Update just tunnel1.
			{
				Config: testAccSiteVPNConnectionConfig_tunnelOptions(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1Updated, tunnel2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "45"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_ike_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "1200"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "360"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", "start"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "45"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_ike_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_lifetime_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "1200"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "360"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", "start"),
				),
			},
			// Update just tunnel2.
			{
				Config: testAccSiteVPNConnectionConfig_tunnelOptions(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1Updated, tunnel2Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn3),
					testAccCheckVPNConnectionNotRecreated(&vpn2, &vpn3),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "45"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_ike_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "1200"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "360"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", "start"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", "clear"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_ike_versions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_ike_versions.*", "ikev1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "14"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_lifetime_seconds", "28800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "12345678"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "540"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", "add"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_vgw_inside_address"),
				),
			},
			// Update tunnel1 and tunnel2.
			{
				Config: testAccSiteVPNConnectionConfig_tunnelOptions(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn4),
					testAccCheckVPNConnectionNotRecreated(&vpn3, &vpn4),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", "clear"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_ike_versions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "14"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "28800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers.#", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "5"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel1_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "540"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", "add"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "45"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_ike_versions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_ike_versions.*", "ikev2"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers.#", "7"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "22"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "23"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_dh_group_numbers.*", "24"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_encryption_algorithms.*", "AES256"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase1_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_lifetime_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers.#", "8"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "15"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "17"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "18"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "19"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "20"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "21"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_dh_group_numbers.*", "22"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES128-GCM-16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_encryption_algorithms.*", "AES256-GCM-16"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-256"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-384"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tunnel2_phase2_integrity_algorithms.*", "SHA2-512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "1200"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "90"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "360"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "512"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", "start"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_vgw_inside_address"),
				),
			},
			// Test resetting to defaults.
			// [local|remote]_ipv[4|6]_network_cidr, tunnel[1|2]_inside_[ipv6_]cidr and tunnel[1|2]_preshared_key are Computed so no diffs will be detected.
			{
				Config: testAccSiteVPNConnectionConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn5),
					testAccCheckVPNConnectionNotRecreated(&vpn4, &vpn5),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_action", "clear"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_dpd_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_ike_versions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_dh_group_numbers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_encryption_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_integrity_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase1_lifetime_seconds", "28800"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_dh_group_numbers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_encryption_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_integrity_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_phase2_lifetime_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_fuzz_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_rekey_margin_time_seconds", "540"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_replay_window_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_startup_action", "add"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel1_vgw_inside_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_address"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_bgp_asn"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_bgp_holdtime", "30"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_cgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_action", "clear"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_dpd_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_ike_versions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", ""),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_dh_group_numbers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_encryption_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase1_integrity_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_dh_group_numbers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_encryption_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_integrity_algorithms.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_phase2_lifetime_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_fuzz_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_rekey_margin_time_seconds", "540"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_replay_window_size", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_startup_action", "add"),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel2_vgw_inside_address"),
					resource.TestCheckResourceAttr(resourceName, "tunnel_inside_ip_version", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "vgw_telemetry.#", "2"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSiteVPNConnection_staticRoutes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_staticRoutes(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "true"),
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

func TestAccSiteVPNConnection_outsideAddressTypePrivate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_outsideAddressTypePrivate(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "outside_ip_address_type", "PrivateIpv4"),
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

func TestAccSiteVPNConnection_outsideAddressTypePublic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_outsideAddressTypePublic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "outside_ip_address_type", "PublicIpv4"),
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

func TestAccSiteVPNConnection_enableAcceleration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_enableAcceleration(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "true"),
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

func TestAccSiteVPNConnection_ipv6(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_ipv6(rName, rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d201/128", "fd00:2001:db8:2:2d1:81ff:fe41:d202/128", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
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

func TestAccSiteVPNConnection_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_tags1(rName, rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_tags2(rName, rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSiteVPNConnectionConfig_tags1(rName, rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_specifyIPv4(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_localRemoteIPv4CIDRs(rName, rBgpAsn, "10.111.0.0/16", "10.222.33.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "local_ipv4_network_cidr", "10.111.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv4_network_cidr", "10.222.33.0/24"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_localRemoteIPv4CIDRs(rName, rBgpAsn, "10.112.0.0/16", "10.222.32.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "local_ipv4_network_cidr", "10.112.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv4_network_cidr", "10.222.32.0/24"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_specifyIPv6(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_ipv6(rName, rBgpAsn, "1111:2222:3333:4444::/64", "5555:6666:7777::/48", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "local_ipv6_network_cidr", "1111:2222:3333:4444::/64"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv6_network_cidr", "5555:6666:7777::/48"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_basic(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPNConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSiteVPNConnection_updateCustomerGatewayID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn1 := sdkacctest.RandIntRange(64512, 65534)
	rBgpAsn2 := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2 ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_customerGatewayID(rName, rBgpAsn1, rBgpAsn2),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttrPair(resourceName, "customer_gateway_id", "aws_customer_gateway.test1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_customerGatewayIDUpdated(rName, rBgpAsn1, rBgpAsn2),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttrPair(resourceName, "customer_gateway_id", "aws_customer_gateway.test2", "id"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_updateVPNGatewayID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2 ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_vpnGatewayID(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", "aws_vpn_gateway.test1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_vpnGatewayIDUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", "aws_vpn_gateway.test2", "id"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_updateTransitGatewayID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2 ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayID(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", "aws_ec2_transit_gateway.test1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayIDUpdated(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttrSet(resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", "aws_ec2_transit_gateway.test2", "id"),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_vpnGatewayIDToTransitGatewayID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2 ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayIDOrVPNGatewayID(rName, rBgpAsn, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", "aws_vpn_gateway.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayIDOrVPNGatewayID(rName, rBgpAsn, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", "aws_ec2_transit_gateway.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpn_gateway_id", ""),
				),
			},
		},
	})
}

func TestAccSiteVPNConnection_transitGatewayIDToVPNGatewayID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn1, vpn2 ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayIDOrVPNGatewayID(rName, rBgpAsn, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn1),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", "aws_ec2_transit_gateway.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpn_gateway_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNConnectionConfig_transitGatewayIDOrVPNGatewayID(rName, rBgpAsn, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn2),
					testAccCheckVPNConnectionNotRecreated(&vpn1, &vpn2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpn_gateway_id", "aws_vpn_gateway.test", "id"),
				),
			},
		},
	})
}

func testAccVPNConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_connection" {
			continue
		}

		_, err := tfec2.FindVPNConnectionByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 VPN Connection %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccVPNConnectionExists(n string, v *ec2.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 VPN Connection ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindVPNConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVPNConnectionNotRecreated(before, after *ec2.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.VpnConnectionId), aws.StringValue(after.VpnConnectionId); before != after {
			return fmt.Errorf("Expected EC2 VPN Connection IDs not to change, but got before: %s, after: %s", before, after)
		}

		return nil
	}
}

func testAccSiteVPNConnectionConfig_basic(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_customerGatewayID(rName string, rBgpAsn1, rBgpAsn2 int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test1" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test2" {
  bgp_asn    = %[3]d
  ip_address = "178.0.0.16"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test1.id
  type                = "ipsec.1"
}
`, rName, rBgpAsn1, rBgpAsn2)
}

func testAccSiteVPNConnectionConfig_customerGatewayIDUpdated(rName string, rBgpAsn1, rBgpAsn2 int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test1" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test2" {
  bgp_asn    = %[3]d
  ip_address = "178.0.0.16"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test2.id
  type                = "ipsec.1"
}
`, rName, rBgpAsn1, rBgpAsn2)
}

func testAccSiteVPNConnectionConfig_vpnGatewayID(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test1.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_vpnGatewayIDUpdated(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test2.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_outsideAddressTypePrivate(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "64521"
}

resource "aws_ec2_transit_gateway" "test" {
  amazon_side_asn = "64522"
  description     = %[1]q
  transit_gateway_cidr_blocks = [
    "10.0.0.0/24",
  ]
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = 64523
  ip_address = "10.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association" "test" {
  dx_gateway_id         = aws_dx_gateway.test.id
  associated_gateway_id = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.0.0.0/8",
  ]
}

data "aws_ec2_transit_gateway_dx_gateway_attachment" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  dx_gateway_id      = aws_dx_gateway.test.id

  depends_on = [
    aws_dx_gateway_association.test
  ]
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id                     = aws_customer_gateway.test.id
  outside_ip_address_type                 = "PrivateIpv4"
  transit_gateway_id                      = aws_ec2_transit_gateway.test.id
  transport_transit_gateway_attachment_id = data.aws_ec2_transit_gateway_dx_gateway_attachment.test.id
  type                                    = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_outsideAddressTypePublic(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id     = aws_customer_gateway.test.id
  outside_ip_address_type = "PublicIpv4"
  type                    = "ipsec.1"
  vpn_gateway_id          = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_staticRoutes(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_enableAcceleration(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = true

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_ipv6(rName string, rBgpAsn int, localIpv6NetworkCidr string, remoteIpv6NetworkCidr string, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = false

  local_ipv6_network_cidr  = %[3]q
  remote_ipv6_network_cidr = %[4]q
  tunnel_inside_ip_version = "ipv6"

  tunnel1_inside_ipv6_cidr = %[5]q
  tunnel2_inside_ipv6_cidr = %[6]q

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, localIpv6NetworkCidr, remoteIpv6NetworkCidr, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccSiteVPNConnectionConfig_singleTunnelOptions(rName string, rBgpAsn int, psk string, tunnelCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false

  tunnel1_inside_cidr   = %[3]q
  tunnel1_preshared_key = %[4]q

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, tunnelCidr, psk)
}

func testAccSiteVPNConnectionConfig_transitGateway(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_transitGatewayID(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test1" {
  description = "%[1]s-1"
}

resource "aws_ec2_transit_gateway" "test2" {
  description = "%[1]s-2"
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test1.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_transitGatewayIDUpdated(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test1" {
  description = "%[1]s-1"
}

resource "aws_ec2_transit_gateway" "test2" {
  description = "%[1]s-2"
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test2.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNConnectionConfig_tunnel1InsideCIDR(rName string, rBgpAsn int, tunnel1InsideCidr string, tunnel2InsideCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  tunnel1_inside_cidr = %[3]q
  tunnel2_inside_cidr = %[4]q
  type                = "ipsec.1"
  vpn_gateway_id      = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, tunnel1InsideCidr, tunnel2InsideCidr)
}

func testAccSiteVPNConnectionConfig_tunnel1InsideIPv6CIDR(rName string, rBgpAsn int, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id      = aws_customer_gateway.test.id
  transit_gateway_id       = aws_ec2_transit_gateway.test.id
  tunnel_inside_ip_version = "ipv6"
  tunnel1_inside_ipv6_cidr = %[3]q
  tunnel2_inside_ipv6_cidr = %[4]q
  type                     = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccSiteVPNConnectionConfig_tunnel1PresharedKey(rName string, rBgpAsn int, tunnel1PresharedKey string, tunnel2PresharedKey string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id   = aws_customer_gateway.test.id
  tunnel1_preshared_key = %[3]q
  tunnel2_preshared_key = %[4]q
  type                  = "ipsec.1"
  vpn_gateway_id        = aws_vpn_gateway.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, tunnel1PresharedKey, tunnel2PresharedKey)
}

func testAccSiteVPNConnectionConfig_tunnelOptions(
	rName string,
	rBgpAsn int,
	localIpv4NetworkCidr string,
	remoteIpv4NetworkCidr string,
	tunnel1 TunnelOptions,
	tunnel2 TunnelOptions,
) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"

  local_ipv4_network_cidr  = %[3]q
  remote_ipv4_network_cidr = %[4]q

  tunnel1_inside_cidr                  = %[5]q
  tunnel1_preshared_key                = %[6]q
  tunnel1_dpd_timeout_action           = %[7]q
  tunnel1_dpd_timeout_seconds          = %[8]d
  tunnel1_ike_versions                 = [%[9]s]
  tunnel1_phase1_dh_group_numbers      = [%[10]s]
  tunnel1_phase1_encryption_algorithms = [%[11]s]
  tunnel1_phase1_integrity_algorithms  = [%[12]s]
  tunnel1_phase1_lifetime_seconds      = %[13]d
  tunnel1_phase2_dh_group_numbers      = [%[14]s]
  tunnel1_phase2_encryption_algorithms = [%[15]s]
  tunnel1_phase2_integrity_algorithms  = [%[16]s]
  tunnel1_phase2_lifetime_seconds      = %[17]d
  tunnel1_rekey_fuzz_percentage        = %[18]d
  tunnel1_rekey_margin_time_seconds    = %[19]d
  tunnel1_replay_window_size           = %[20]d
  tunnel1_startup_action               = %[21]q

  tunnel2_inside_cidr                  = %[22]q
  tunnel2_preshared_key                = %[23]q
  tunnel2_dpd_timeout_action           = %[24]q
  tunnel2_dpd_timeout_seconds          = %[25]d
  tunnel2_ike_versions                 = [%[26]s]
  tunnel2_phase1_dh_group_numbers      = [%[27]s]
  tunnel2_phase1_encryption_algorithms = [%[28]s]
  tunnel2_phase1_integrity_algorithms  = [%[29]s]
  tunnel2_phase1_lifetime_seconds      = %[30]d
  tunnel2_phase2_dh_group_numbers      = [%[31]s]
  tunnel2_phase2_encryption_algorithms = [%[32]s]
  tunnel2_phase2_integrity_algorithms  = [%[33]s]
  tunnel2_phase2_lifetime_seconds      = %[34]d
  tunnel2_rekey_fuzz_percentage        = %[35]d
  tunnel2_rekey_margin_time_seconds    = %[36]d
  tunnel2_replay_window_size           = %[37]d
  tunnel2_startup_action               = %[38]q

  tags = {
    Name = %[1]q
  }
}
`,
		rName,
		rBgpAsn,
		localIpv4NetworkCidr,
		remoteIpv4NetworkCidr,
		tunnel1.tunnelCidr,
		tunnel1.psk,
		tunnel1.dpdTimeoutAction,
		tunnel1.dpdTimeoutSeconds,
		tunnel1.ikeVersions,
		tunnel1.phase1DhGroupNumbers,
		tunnel1.phase1EncryptionAlgorithms,
		tunnel1.phase1IntegrityAlgorithms,
		tunnel1.phase1LifetimeSeconds,
		tunnel1.phase2DhGroupNumbers,
		tunnel1.phase2EncryptionAlgorithms,
		tunnel1.phase2IntegrityAlgorithms,
		tunnel1.phase2LifetimeSeconds,
		tunnel1.rekeyFuzzPercentage,
		tunnel1.rekeyMarginTimeSeconds,
		tunnel1.replayWindowSize,
		tunnel1.startupAction,
		tunnel2.tunnelCidr,
		tunnel2.psk,
		tunnel2.dpdTimeoutAction,
		tunnel2.dpdTimeoutSeconds,
		tunnel2.ikeVersions,
		tunnel2.phase1DhGroupNumbers,
		tunnel2.phase1EncryptionAlgorithms,
		tunnel2.phase1IntegrityAlgorithms,
		tunnel2.phase1LifetimeSeconds,
		tunnel2.phase2DhGroupNumbers,
		tunnel2.phase2EncryptionAlgorithms,
		tunnel2.phase2IntegrityAlgorithms,
		tunnel2.phase2LifetimeSeconds,
		tunnel2.rekeyFuzzPercentage,
		tunnel2.rekeyMarginTimeSeconds,
		tunnel2.replayWindowSize,
		tunnel2.startupAction)
}

func testAccSiteVPNConnectionConfig_tags1(rName string, rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1)
}

func testAccSiteVPNConnectionConfig_tags2(rName string, rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSiteVPNConnectionConfig_localRemoteIPv4CIDRs(rName string, rBgpAsn int, localIpv4Cidr string, remoteIpv4Cidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false

  local_ipv4_network_cidr  = %[3]q
  remote_ipv4_network_cidr = %[4]q

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, localIpv4Cidr, remoteIpv4Cidr)
}

func testAccSiteVPNConnectionConfig_transitGatewayIDOrVPNGatewayID(rName string, rBgpAsn int, useTransitGateway bool) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = %[3]t ? aws_ec2_transit_gateway.test.id : null
  vpn_gateway_id      = %[3]t ? null : aws_vpn_gateway.test.id
  type                = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, useTransitGateway)
}

// Test our VPN tunnel config XML parsing
const testAccVPNTunnelInfoXML = `
<vpn_connection id="vpn-abc123">
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>22.22.22.22</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>2.2.2.2</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>2222</asn>
        <hold_time>32</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>SECOND_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>11.11.11.11</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.11.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>1.1.1.1</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>168.254.11.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>1111</asn>
        <hold_time>31</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>FIRST_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
</vpn_connection>
`
