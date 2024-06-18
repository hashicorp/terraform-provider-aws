// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NOTE: acceptance tests require environment variable ACM_CERTIFICATE_ARN
// to be set and the ACM certificate to be validated during testing.

func TestAccNetworkFirewallTLSInspectionConfiguration_combinedIngressEgressBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}
	caCertificateArn := os.Getenv("ACM_CA_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", caCertificateArn)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_combinedIngressEgress(rName, certificateArn, caCertificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						names.AttrKeyID: "AWS_OWNED_KMS_KEY",
						names.AttrType:  "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn", caCertificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTLSInspectionConfigurationConfig_combinedIngressEgressUpdate(rName, certificateArn, caCertificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						names.AttrKeyID: "AWS_OWNED_KMS_KEY",
						names.AttrType:  "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65534"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn", caCertificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "REJECT"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
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

func TestAccNetworkFirewallTLSInspectionConfiguration_egressBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	ca := os.Getenv("ACM_CA_CERTIFICATE_ARN")
	if ca == "" {
		t.Skipf("Environment variable %s is not set, skipping test", ca)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_egressBasic(rName, ca),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						names.AttrKeyID: "AWS_OWNED_KMS_KEY",
						names.AttrType:  "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn", ca),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65535"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
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

func TestAccNetworkFirewallTLSInspectionConfiguration_egressWithEncryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	ca := os.Getenv("ACM_CA_CERTIFICATE_ARN")
	if ca == "" {
		t.Skipf("Environment variable %s is not set, skipping test", ca)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	keyName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_egressWithEncryptionConfiguration(rName, ca),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "encryption_configuration.0.key_id", keyName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn", ca),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65535"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
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

func TestAccNetworkFirewallTLSInspectionConfiguration_ingressBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressBasic(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "encryption_configuration.*", map[string]string{
						names.AttrKeyID: "AWS_OWNED_KMS_KEY",
						names.AttrType:  "AWS_OWNED_KMS_KEY",
					}),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "certificates.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
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

func TestAccNetworkFirewallTLSInspectionConfiguration_ingressWithEncryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	keyName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressWithEncryptionConfiguration(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "encryption_configuration.0.key_id", keyName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn", certificateArn),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.0", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", acctest.Ct0),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/+.`)),
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

func TestAccNetworkFirewallTLSInspectionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	certificateArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skipf("Environment variable %s is not set, skipping test", certificateArn)
	}

	var tlsinspectionconfiguration networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_ingressBasic(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, resourceName, &tlsinspectionconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceTLSInspectionConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTLSInspectionConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_tls_inspection_configuration" {
				continue
			}

			_, err := tfnetworkfirewall.FindTLSInspectionConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall TLS Inspection Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTLSInspectionConfigurationExists(ctx context.Context, n string, v *networkfirewall.DescribeTLSInspectionConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn(ctx)

		output, err := tfnetworkfirewall.FindTLSInspectionConfigurationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTLSInspectionConfigurationConfig_combinedIngressEgress(rName, certificateARN, ca string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn = %[3]q
      check_certificate_revocation_status {
        revoked_status_action = "REJECT"
        unknown_status_action = "PASS"
      }
      server_certificate {
        resource_arn = %[2]q
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 443
        }
        destination {
          address_definition = "0.0.0.0/0"
        }
        source_ports {
          from_port = 0
          to_port   = 65535
        }
        source {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, certificateARN, ca)
}

func testAccTLSInspectionConfigurationConfig_combinedIngressEgressUpdate(rName, certificateARN, ca string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn = %[3]q
      check_certificate_revocation_status {
        revoked_status_action = "PASS"
        unknown_status_action = "REJECT"
      }
      server_certificate {
        resource_arn = %[2]q
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 8080
        }
        destination {
          address_definition = "10.0.0.0/8"
        }
        source_ports {
          from_port = 1024
          to_port   = 65534
        }
        source {
          address_definition = "10.0.0.0/8"
        }
      }
    }
  }
}
`, rName, certificateARN, ca)
}

func testAccTLSInspectionConfigurationConfig_ingressBasic(rName, certificateARN string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = %[2]q
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 443
        }
        destination {
          address_definition = "0.0.0.0/0"
        }
        source_ports {
          from_port = 0
          to_port   = 65535
        }
        source {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, certificateARN)
}

func testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccTLSInspectionConfigurationConfig_ingressWithEncryptionConfiguration(rName, certificateARN string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = %[2]q
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 443
        }
        destination {
          address_definition = "0.0.0.0/0"
        }
        source_ports {
          from_port = 0
          to_port   = 65535
        }
        source {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, certificateARN))
}

func testAccTLSInspectionConfigurationConfig_egressBasic(rName, ca string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn = %[2]q
      check_certificate_revocation_status {
        revoked_status_action = "REJECT"
        unknown_status_action = "PASS"
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 443
        }
        destination {
          address_definition = "0.0.0.0/0"
        }
        source_ports {
          from_port = 0
          to_port   = 65535
        }
        source {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, ca)
}

func testAccTLSInspectionConfigurationConfig_egressWithEncryptionConfiguration(rName, ca string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"
  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }
  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn = %[2]q
      check_certificate_revocation_status {
        revoked_status_action = "REJECT"
        unknown_status_action = "PASS"
      }
      scope {
        protocols = [6]
        destination_ports {
          from_port = 443
          to_port   = 443
        }
        destination {
          address_definition = "0.0.0.0/0"
        }
        source_ports {
          from_port = 0
          to_port   = 65535
        }
        source {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName, ca))
}
