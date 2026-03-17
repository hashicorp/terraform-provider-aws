// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallTLSInspectionConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_basic(rName, commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/.+$`)),
					resource.TestCheckNoResourceAttr(resourceName, "certificate_authority"),
					resource.TestCheckResourceAttr(resourceName, "certificates.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.*", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"tls_inspection_configuration", "update_token"},
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_basic(rName, commonName.String(), certificateDomainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceTLSInspectionConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_tags1(rName, commonName.String(), certificateDomainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"tls_inspection_configuration", "update_token"},
			},
			{
				Config: testAccTLSInspectionConfigurationConfig_tags2(rName, commonName.String(), certificateDomainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTLSInspectionConfigurationConfig_tags1(rName, commonName.String(), certificateDomainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName, commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/.+$`)),
					resource.TestCheckNoResourceAttr(resourceName, "certificate_authority"),
					resource.TestCheckResourceAttr(resourceName, "certificates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "encryption_configuration.0.key_id", kmsKeyResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.*", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65534"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.0.resource_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"tls_inspection_configuration", "update_token"},
			},
		},
	})
}

func TestAccNetworkFirewallTLSInspectionConfiguration_checkCertificateRevocationStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeTLSInspectionConfigurationOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	resourceName := "aws_networkfirewall_tls_inspection_configuration.test"
	testExternalProviders := map[string]resource.ExternalProvider{
		"tls": {
			Source:            "hashicorp/tls",
			VersionConstraint: "4.0.5",
		},
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders:        testExternalProviders,
		CheckDestroy:             testAccCheckTLSInspectionConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTLSInspectionConfigurationConfig_checkCertificateRevocationStatus(rName, commonName.String(), certificateDomainName, "REJECT", "PASS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/.+$`)),
					resource.TestCheckNoResourceAttr(resourceName, "certificates"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "REJECT"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.*", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65534"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"tls_inspection_configuration", "update_token"},
			},
			{
				Config: testAccTLSInspectionConfigurationConfig_checkCertificateRevocationStatus(rName, commonName.String(), certificateDomainName, "DROP", "PASS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTLSInspectionConfigurationExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "network-firewall", regexache.MustCompile(`tls-configuration/.+$`)),
					resource.TestCheckNoResourceAttr(resourceName, "certificates"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.key_id", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "AWS_OWNED_KMS_KEY"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "number_of_associations"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.certificate_authority_arn"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.revoked_status_action", "DROP"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.check_certificate_revocation_status.0.unknown_status_action", "PASS"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination.0.address_definition", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.destination_ports.0.to_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.protocols.*", "6"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source.0.address_definition", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.from_port", "1024"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.scope.0.source_ports.0.to_port", "65534"),
					resource.TestCheckResourceAttr(resourceName, "tls_inspection_configuration.0.server_certificate_configuration.0.server_certificate.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_inspection_configuration_id"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
				),
			},
		},
	})
}

func testAccCheckTLSInspectionConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_tls_inspection_configuration" {
				continue
			}

			_, err := tfnetworkfirewall.FindTLSInspectionConfigurationByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckTLSInspectionConfigurationExists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeTLSInspectionConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindTLSInspectionConfigurationByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, certificateDomainName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[2]q
    }
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}

resource "aws_acm_certificate" "test" {
  domain_name               = %[3]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}
`, rName, commonName, certificateDomainName)
}

func testAccTLSInspectionConfigurationConfig_certificateCheckCertificateRevocationStatus(commonName, certificateDomainName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  private_key_pem = tls_private_key.test.private_key_pem

  subject {
    common_name = %[1]q
  }

  is_ca_certificate    = true
  set_subject_key_id   = true
  set_authority_key_id = true

  validity_period_hours = 9000

  allowed_uses = [
    "cert_signing",
    "crl_signing",
    "digital_signature"
  ]
}

resource "aws_acm_certificate" "test" {
  private_key      = tls_private_key.test.private_key_pem
  certificate_body = tls_self_signed_cert.test.cert_pem
}
`, commonName, certificateDomainName)
}

func testAccTLSInspectionConfigurationConfig_basic(rName, commonName, certificateDomainName string) string {
	return acctest.ConfigCompose(
		testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, certificateDomainName),
		fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q

  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = aws_acm_certificate.test.arn
      }
      scope {
        protocols = [6]
        destination {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }
}
`, rName))
}

func testAccTLSInspectionConfigurationConfig_tags1(rName, commonName, certificateDomainName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, certificateDomainName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q

  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = aws_acm_certificate.test.arn
      }
      scope {
        protocols = [6]
        destination {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccTLSInspectionConfigurationConfig_tags2(rName, commonName, certificateDomainName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, certificateDomainName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name = %[1]q

  tls_inspection_configuration {
    server_certificate_configuration {
      server_certificate {
        resource_arn = aws_acm_certificate.test.arn
      }
      scope {
        protocols = [6]
        destination {
          address_definition = "0.0.0.0/0"
        }
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTLSInspectionConfigurationConfig_encryptionConfiguration(rName, commonName, certificateDomainName string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_certificateBase(rName, commonName, certificateDomainName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

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
        resource_arn = aws_acm_certificate.test.arn
      }
      scope {
        protocols = [6]

        destination {
          address_definition = "0.0.0.0/0"
        }
        destination_ports {
          from_port = 443
          to_port   = 8080
        }

        source {
          address_definition = "10.0.0.0/8"
        }
        source_ports {
          from_port = 1024
          to_port   = 65534
        }
      }
    }
  }
}
`, rName))
}

func testAccTLSInspectionConfigurationConfig_checkCertificateRevocationStatus(rName, commonName, certificateDomainName, revokedStatusAction, unknownStatusAction string) string {
	return acctest.ConfigCompose(testAccTLSInspectionConfigurationConfig_certificateCheckCertificateRevocationStatus(commonName, certificateDomainName), fmt.Sprintf(`
resource "aws_networkfirewall_tls_inspection_configuration" "test" {
  name        = %[1]q
  description = "test"

  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }

  tls_inspection_configuration {
    server_certificate_configuration {
      certificate_authority_arn = aws_acm_certificate.test.arn
      check_certificate_revocation_status {
        revoked_status_action = %[2]q
        unknown_status_action = %[3]q
      }
      scope {
        protocols = [6]

        destination {
          address_definition = "0.0.0.0/0"
        }
        destination_ports {
          from_port = 443
          to_port   = 8080
        }

        source {
          address_definition = "10.0.0.0/8"
        }
        source_ports {
          from_port = 1024
          to_port   = 65534
        }
      }
    }
  }
}
`, rName, revokedStatusAction, unknownStatusAction))
}
