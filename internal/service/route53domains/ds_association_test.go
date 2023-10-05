// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfroute53domains "github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
)

func TestAccRoute53DomainsDelegationSignerAssocation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_route53domains_ds_association.test"

	domainEnvironmentVariableName := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(domainEnvironmentVariableName)
	signingAlgorithmType := "13"
	flag := "257"
	publicKey := "M1jizbcdK5IBt7d4lRIiFTx+N06BJzaeScK215EOfa8efd2akkUWmxD5OhodZCIiRMvL5ZnStUP5UReIeUmBeA=="

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53DomainsEndpointID)
			testDelegationSignerAssociationAccPreCheck(ctx, domainEnvironmentVariableName, domainName, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, signingAlgorithmType, flag, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_type", signingAlgorithmType),
					resource.TestCheckResourceAttr(resourceName, "flag", flag),
					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDelegationSignerAssociationImportStateIDFunc(ctx, resourceName),
				ImportStateVerifyIdentifierAttribute: "dnssec_key_id",
			},
		},
	})
}

func TestAccRoute53DomainsDelegationSignerAssocation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_route53domains_ds_association.test"

	domainEnvironmentVariableName := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(domainEnvironmentVariableName)
	signingAlgorithmType := "13"
	flag := "257"
	publicKey := "M1jizbcdK5IBt7d4lRIiFTx+N06BJzaeScK215EOfa8efd2akkUWmxD5OhodZCIiRMvL5ZnStUP5UReIeUmBeA=="

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53DomainsEndpointID)
			testDelegationSignerAssociationAccPreCheck(ctx, domainEnvironmentVariableName, domainName, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDelegationSignerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDelegationSignerAssociationConfig_basic(domainName, signingAlgorithmType, flag, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDelegationSignerAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53domains.ResourceDelegationSignerAssociation, resourceName),
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

			domainName := rs.Primary.Attributes["domain_name"]
			dnssecKeyId := rs.Primary.Attributes["dnssec_key_id"]

			input := &route53domains.GetDomainDetailInput{
				DomainName: &domainName,
			}

			out, err := conn.GetDomainDetail(ctx, input)

			if err != nil {
				return create.Error(names.Route53Domains, create.ErrActionCheckingDestroyed, tfroute53domains.ResNameDelegationSignerAssociation, rs.Primary.ID, err)
			}

			dnssecKey := tfroute53domains.GetDnssecKeyWithId(out.DnssecKeys, dnssecKeyId)
			if dnssecKey == nil {
				return nil
			}

			return create.Error(names.Route53Domains, create.ErrActionCheckingDestroyed, tfroute53domains.ResNameDelegationSignerAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckDelegationSignerAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(names.Route53Domains, create.ErrActionCheckingExistence, tfroute53domains.ResNameDelegationSignerAssociation, resourceName, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Route53Domains, create.ErrActionCheckingExistence, tfroute53domains.ResNameDelegationSignerAssociation, resourceName, errors.New("not set"))
		}

		domainName := rs.Primary.Attributes["domain_name"]
		dnssecKeyId := rs.Primary.Attributes["dnssec_key_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

		input := &route53domains.GetDomainDetailInput{
			DomainName: &domainName,
		}

		out, err := conn.GetDomainDetail(ctx, input)

		if err != nil {
			return create.Error(names.Route53Domains, create.ErrActionCheckingExistence, tfroute53domains.ResNameDelegationSignerAssociation, rs.Primary.ID, err)
		}

		dnssecKey := tfroute53domains.GetDnssecKeyWithId(out.DnssecKeys, dnssecKeyId)

		if dnssecKey == nil {
			return create.Error(names.Route53Domains, create.ErrActionCheckingExistence, tfroute53domains.ResNameDelegationSignerAssociation, resourceName, errors.New("not found"))
		}

		return nil
	}
}

func testDelegationSignerAssociationAccPreCheck(ctx context.Context, domainEnvironmentVariableName string, domainName string, t *testing.T) {
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", domainEnvironmentVariableName)
	}
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

	input := &route53domains.GetDomainDetailInput{
		DomainName: &domainName,
	}
	_, err := conn.GetDomainDetail(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDelegationSignerAssociationImportStateIDFunc(ctx context.Context, resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", create.Error(names.Route53Domains, create.ErrActionImporting, tfroute53domains.ResNameDelegationSignerAssociation, resourceName, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return "", create.Error(names.Route53Domains, create.ErrActionImporting, tfroute53domains.ResNameDelegationSignerAssociation, resourceName, errors.New("not set"))
		}

		domainName := rs.Primary.Attributes["domain_name"]
		dnssecKeyId := rs.Primary.Attributes["dnssec_key_id"]

		return fmt.Sprintf("%s:%s", domainName, dnssecKeyId), nil
	}
}

func testAccDelegationSignerAssociationConfig_basic(domainName string, signingAlgorithmType string, flag string, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_ds_association" "test" {
  domain_name            = %[1]q
  signing_algorithm_type = %[2]q
  flag                   = %[3]q
  public_key             = %[4]q
}
`, domainName, signingAlgorithmType, flag, publicKey)
}
