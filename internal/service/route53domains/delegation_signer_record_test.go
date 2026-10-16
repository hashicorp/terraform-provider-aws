// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53domains "github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegationSignerRecord_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_delegation_signer_record.test"
	publicKey := "M1jizbcdK5IBt7d4lRIiFTx+N06BJzaeScK215EOfa8efd2akkUWmxD5OhodZCIiRMvL5ZnStUP5UReIeUmBeA=="

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "dnssec_key_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccDelegationSignerAssociationImportStateIDFunc(ctx, resourceName),
			},
		},
	})
}

func testAccDelegationSignerRecord_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_delegation_signer_record.test"
	publicKey := "M1jizbcdK5IBt7d4lRIiFTx+N06BJzaeScK215EOfa8efd2akkUWmxD5OhodZCIiRMvL5ZnStUP5UReIeUmBeA=="

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfroute53domains.ResourceDelegationSignerRecord, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckDelegationSignerAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53DomainsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53domains_delegation_signer_record" {
				continue
			}

			_, err := tfroute53domains.FindDNSSECKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Domains Delegation Signer Record %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDelegationSignerAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53DomainsClient(ctx)

		_, err := tfroute53domains.FindDNSSECKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"])

		return err
	}
}

func testAccDelegationSignerAssociationImportStateIDFunc(_ context.Context, n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"]), nil
	}
}

func testAccDelegationSignerAssociationConfig_basic(domainName string, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_delegation_signer_record" "test" {
  domain_name = %[1]q

  signing_attributes {
    algorithm  = 13
    flags      = 257
    public_key = %[2]q
  }
}
`, domainName, publicKey)
}

// TestCalculateExpectedDigest verifies RFC 4034 compliant digest generation
func TestCalculateExpectedDigest(t *testing.T) {
	t.Run("Valid RFC 4034 digest generation", func(t *testing.T) {
		// Example domain and key configuration
		domainName := "example.com"
		flags := 257       // KSK/ZSK
		algorithm := 13    // ECDSAP256SHA256
		publicKey := "AwEAAazJ1O4G2F9O7lYvRjWp1A5MvXk1zY7B2C4D5E6F7G8H9I0J"

		digest := tfroute53domains.CalculateExpectedDigest(domainName, flags, algorithm, publicKey)

		// Verify it returns a valid hex string (64 chars for SHA-256)
		if len(digest) != 64 {
			t.Errorf("Expected 64-char hex digest, got %d chars: %s", len(digest), digest)
		}

		// Verify it's uppercase hex
		for _, ch := range digest {
			if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
				t.Errorf("Digest contains non-hex character: %c", ch)
			}
		}
	})

	t.Run("Domain case insensitivity (RFC 4034)", func(t *testing.T) {
		domainLower := "example.com"
		domainUpper := "EXAMPLE.COM"
		flags := 257
		algorithm := 13
		publicKey := "AwEAAazJ1O4G2F9O7lYvRjWp1A5MvXk1zY7B2C4D5E6F7G8H9I0J"

		digestLower := tfroute53domains.CalculateExpectedDigest(domainLower, flags, algorithm, publicKey)
		digestUpper := tfroute53domains.CalculateExpectedDigest(domainUpper, flags, algorithm, publicKey)

		if digestLower != digestUpper {
			t.Errorf("Domain case sensitivity violation.\nLower: %s\nUpper: %s", digestLower, digestUpper)
		}
	})

	t.Run("Invalid Base64 public key graceful handling", func(t *testing.T) {
		domainName := "example.com"
		flags := 257
		algorithm := 13
		invalidKey := "not_valid_base64!!!"

		// Should not crash, just use empty key bytes
		digest := tfroute53domains.CalculateExpectedDigest(domainName, flags, algorithm, invalidKey)

		if len(digest) == 0 {
			t.Errorf("Digest should not be empty even with invalid key")
		}
	})
}
