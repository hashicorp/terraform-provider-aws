// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRegisteredDomain_tags(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_tags1(domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegisteredDomainConfig_tags2(domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_tags1(domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccRegisteredDomain_autoRenew(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_autoRenew(domainName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", acctest.CtFalse),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_autoRenew(domainName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccRegisteredDomain_contacts(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_contacts(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.address_line_1", "99 High Street"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.address_line_2", "Flat 1a"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.city", "Little Nowhere"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.country_code", "GB"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.email", "terraform-acctest+aws-route53domains-test1@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.fax", "+44.123456788"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.first_name", "Sys"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.last_name", "Admin"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.organization_name", "Support"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.phone_number", "+44.123456789"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.zip_code", "ST1 1AB"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.address_line_1", "1 Mawson Street"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.address_line_2", "Unit 2"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.city", "Mawson"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.contact_type", "PERSON"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.country_code", "AU"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.email", "terraform-acctest+aws-route53domains-test4@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.fax", "+61.412345678"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.first_name", "John"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.last_name", "Cleese"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.organization_name", ""),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.phone_number", "+61.412345679"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.state", "ACT"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.zip_code", "2606"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.address_line_1", "100 Main Street"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.city", "New York City"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.email", "terraform-acctest+aws-route53domains-test2@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.phone_number", "+1.2025551234"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.state", "NY"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.zip_code", "10001"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.address_line_1", "The Castle"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.city", "Prague"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.contact_type", "PERSON"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.country_code", "CZ"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.email", "terraform-acctest+aws-route53domains-test3@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.first_name", "Franz"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.last_name", "Kafka"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.phone_number", "+420.224372434"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.zip_code", "119 01"),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_contactsUpdated(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.address_line_1", "101 2nd St #700"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.city", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.email", "terraform-acctest+aws@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.fax", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.phone_number", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.state", "CA"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.zip_code", "94105"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.address_line_1", "101 2nd St #700"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.city", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.email", "terraform-acctest+aws@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.fax", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.phone_number", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.state", "CA"),
					resource.TestCheckResourceAttr(resourceName, "billing_contact.0.zip_code", "94105"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.address_line_1", "101 2nd St #700"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.city", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.email", "terraform-acctest+aws@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.fax", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.phone_number", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.state", "CA"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.zip_code", "94105"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.address_line_1", "101 2nd St #700"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.city", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.email", "terraform-acctest+aws@hashicorp.com"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.fax", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.phone_number", "+1.4155551234"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.state", "CA"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.zip_code", "94105"),
				),
			},
		},
	})
}

func testAccRegisteredDomain_contactPrivacy(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_contactPrivacy(domainName, true, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "billing_privacy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", acctest.CtTrue),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_contactPrivacy(domainName, false, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "billing_privacy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccRegisteredDomain_nameservers(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_nameservers(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_server.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.name", fmt.Sprintf("ns1.%s", domainName)),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.glue_ips.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "name_server.0.glue_ips.*", "1.1.1.1"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.name", "ns2.worldnic.com"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.glue_ips.#", acctest.Ct0),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_nameserversUpdated(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_server.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.name", "ns-195.awsdns-24.com"),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.glue_ips.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.name", "ns-1632.awsdns-12.co.uk"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.glue_ips.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "name_server.2.name", "ns-874.awsdns-45.net"),
					resource.TestCheckResourceAttr(resourceName, "name_server.2.glue_ips.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "name_server.3.name", "ns-1118.awsdns-11.org"),
					resource.TestCheckResourceAttr(resourceName, "name_server.3.glue_ips.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccRegisteredDomain_transferLock(t *testing.T) {
	ctx := acctest.Context(t)
	domainName := acctest.SkipIfEnvVarNotSet(t, "ROUTE53DOMAINS_DOMAIN_NAME")
	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfig_transferLock(domainName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", acctest.CtFalse),
				),
			},
			{
				Config: testAccRegisteredDomainConfig_transferLock(domainName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccRegisteredDomainConfig_tags1(domainName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, domainName, tagKey1, tagValue1)
}

func testAccRegisteredDomainConfig_tags2(domainName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, domainName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRegisteredDomainConfig_autoRenew(domainName string, autoRenew bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q
  auto_renew  = %[2]t
}
`, domainName, autoRenew)
}

func testAccRegisteredDomainConfig_contacts(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  admin_contact {
    address_line_1    = "99 High Street"
    address_line_2    = "Flat 1a"
    city              = "Little Nowhere"
    contact_type      = "COMPANY"
    country_code      = "GB"
    email             = "terraform-acctest+aws-route53domains-test1@hashicorp.com"
    fax               = "+44.123456788"
    first_name        = "Sys"
    last_name         = "Admin"
    organization_name = "Support"
    phone_number      = "+44.123456789"
    zip_code          = "ST1 1AB"
  }

  billing_contact {
    address_line_1 = "1 Mawson Street"
    address_line_2 = "Unit 2"
    city           = "Mawson"
    contact_type   = "PERSON"
    country_code   = "AU"
    email          = "terraform-acctest+aws-route53domains-test4@hashicorp.com"
    fax            = "+61.412345678"
    first_name     = "John"
    last_name      = "Cleese"
    phone_number   = "+61.412345679"
    state          = "ACT"
    zip_code       = "2606"
  }

  registrant_contact {
    address_line_1    = "100 Main Street"
    city              = "New York City"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "terraform-acctest+aws-route53domains-test2@hashicorp.com"
    first_name        = "Terraform" # Changing owner's first or last name is a change of ownership.
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.2025551234"
    state             = "NY"
    zip_code          = "10001"
  }

  tech_contact {
    address_line_1 = "The Castle"
    city           = "Prague"
    contact_type   = "PERSON"
    country_code   = "CZ"
    email          = "terraform-acctest+aws-route53domains-test3@hashicorp.com"
    first_name     = "Franz"
    last_name      = "Kafka"
    phone_number   = "+420.224372434"
    zip_code       = "119 01"
  }
}
`, domainName)
}

func testAccRegisteredDomainConfig_contactsUpdated(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  admin_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "terraform-acctest+aws@hashicorp.com"
    fax               = "+1.4155551234"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551234"
    state             = "CA"
    zip_code          = "94105"
  }

  billing_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "terraform-acctest+aws@hashicorp.com"
    fax               = "+1.4155551234"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551234"
    state             = "CA"
    zip_code          = "94105"
  }

  registrant_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "terraform-acctest+aws@hashicorp.com"
    fax               = "+1.4155551234"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551234"
    state             = "CA"
    zip_code          = "94105"
  }

  tech_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "terraform-acctest+aws@hashicorp.com"
    fax               = "+1.4155551234"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551234"
    state             = "CA"
    zip_code          = "94105"
  }
}
`, domainName)
}

func testAccRegisteredDomainConfig_contactPrivacy(domainName string, adminPrivacy, billingPrivacy, registrantPrivacy, techPrivacy bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  admin_privacy      = %[2]t
  billing_privacy    = %[3]t
  registrant_privacy = %[4]t
  tech_privacy       = %[5]t
}
`, domainName, adminPrivacy, billingPrivacy, registrantPrivacy, techPrivacy)
}

func testAccRegisteredDomainConfig_nameservers(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  name_server {
    name = "ns1.%[1]s"

    # Glue records are only applicable when the name server is a sub-domain.
    glue_ips = ["1.1.1.1"]
  }

  name_server {
    name = "ns2.worldnic.com"
  }
}
`, domainName)
}

func testAccRegisteredDomainConfig_nameserversUpdated(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  name_server {
    name = "ns-195.awsdns-24.com"
  }

  name_server {
    name = "ns-1632.awsdns-12.co.uk"
  }

  name_server {
    name = "ns-874.awsdns-45.net"
  }

  name_server {
    name = "ns-1118.awsdns-11.org"
  }
}
`, domainName)
}

func testAccRegisteredDomainConfig_transferLock(domainName string, transferLock bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name   = %[1]q
  transfer_lock = %[2]t
}
`, domainName, transferLock)
}
