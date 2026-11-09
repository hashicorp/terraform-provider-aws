// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53domains "github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestDelegationSignerDigest verifies RFC 4034 compliant digest generation
func TestDelegationSignerDigest(t *testing.T) {
	t.Parallel()

	t.Run("Valid RFC 4034 digest generation", func(t *testing.T) {
		t.Parallel()
		domainName := "example.com"
		var (
			flags, algorithm int32 = 257, 13
		)
		publicKey := "AwEAAazJ1O4G2F9O7lYvRjWp1A5MvXk1zY7B2C4D5E6F7G8H9I0J"

		digest := tfroute53domains.DelegationSignerDigest(domainName, flags, algorithm, publicKey)

		if len(digest) != 64 {
			t.Errorf("Expected 64-char hex digest, got %d chars: %s", len(digest), digest)
		}

		for _, ch := range digest {
			if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')) {
				t.Errorf("Digest contains non-hex character: %c", ch)
			}
		}
	})

	t.Run("Domain case insensitivity (RFC 4034)", func(t *testing.T) {
		t.Parallel()

		domainLower := "example.com"
		domainUpper := "EXAMPLE.COM"
		var (
			flags, algorithm int32 = 257, 13
		)
		publicKey := "AwEAAazJ1O4G2F9O7lYvRjWp1A5MvXk1zY7B2C4D5E6F7G8H9I0J"

		digestLower := tfroute53domains.DelegationSignerDigest(domainLower, flags, algorithm, publicKey)
		digestUpper := tfroute53domains.DelegationSignerDigest(domainUpper, flags, algorithm, publicKey)

		if digestLower != digestUpper {
			t.Errorf("Domain case sensitivity violation.\nLower: %s\nUpper: %s", digestLower, digestUpper)
		}
	})

	t.Run("Invalid Base64 public key graceful handling", func(t *testing.T) {
		t.Parallel()

		domainName := "example.com"
		var (
			flags, algorithm int32 = 257, 13
		)
		invalidKey := "not_valid_base64!!!"

		digest := tfroute53domains.DelegationSignerDigest(domainName, flags, algorithm, invalidKey)

		if len(digest) == 0 {
			t.Errorf("Digest should not be empty even with invalid key")
		}
	})
}

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
				ConfigDirectory: config.StaticDirectory("testdata/DelegationSignerRecord/basic/"),
				ConfigVariables: config.Variables{
					"domain_name": config.StringVariable(domainName),
					"public_key":  config.StringVariable(publicKey),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("dnssec_key_id"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/DelegationSignerRecord/basic/"),
				ConfigVariables: config.Variables{
					"domain_name": config.StringVariable(domainName),
					"public_key":  config.StringVariable(publicKey),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrDomainName,
				ImportStateIdFunc:                    testAccDelegationSignerAssociationImportStateIDFunc(ctx, resourceName),
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
				ConfigDirectory: config.StaticDirectory("testdata/DelegationSignerRecord/basic/"),
				ConfigVariables: config.Variables{
					"domain_name": config.StringVariable(domainName),
					"public_key":  config.StringVariable(publicKey),
				},
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

			_, err := tfroute53domains.FindDNSSECKeyByDomainNameAndKeyID(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"])

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

		_, err := tfroute53domains.FindDNSSECKeyByDomainNameAndKeyID(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"])

		return err
	}
}

func testAccDelegationSignerAssociationImportStateIDFunc(_ context.Context, n string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(n, ",", names.AttrDomainName, "signing_attributes.0.flags", "signing_attributes.0.algorithm", "signing_attributes.0.public_key")
}
