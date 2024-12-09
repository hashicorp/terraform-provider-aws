// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53domains "github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDelegationSignerRecord_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_delegation_signer_record.test"
	publicKey := "M1jizbcdK5IBt7d4lRIiFTx+N06BJzaeScK215EOfa8efd2akkUWmxD5OhodZCIiRMvL5ZnStUP5UReIeUmBeA=="

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53domains.ResourceDelegationSignerRecord, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDelegationSignerAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53domains_ds_association" {
				continue
			}

			_, err := tfroute53domains.FindDNSSECKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["dnssec_key_id"])

			if tfresource.NotFound(err) {
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

func testAccCheckDelegationSignerAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

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
