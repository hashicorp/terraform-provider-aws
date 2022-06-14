package route53domains_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53Domains_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"RegisteredDomain": {
			"tags":           testAccRegisteredDomain_tags,
			"autoRenew":      testAccRegisteredDomain_autoRenew,
			"contacts":       testAccRegisteredDomain_contacts,
			"contactPrivacy": testAccRegisteredDomain_contactPrivacy,
			"nameservers":    testAccRegisteredDomain_nameservers,
			"transferLock":   testAccRegisteredDomain_transferLock,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccPreCheck(t *testing.T) {
	acctest.PreCheckPartitionHasService(names.Route53DomainsEndpointID, t)

	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsConn

	input := &route53domains.ListDomainsInput{}

	_, err := conn.ListDomains(context.TODO(), input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRegisteredDomain_tags(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainConfigTags1(domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRegisteredDomainConfigTags2(domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRegisteredDomainConfigTags1(domainName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccRegisteredDomain_autoRenew(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainAutoRenewConfig(domainName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "false"),
				),
			},
			{
				Config: testAccRegisteredDomainAutoRenewConfig(domainName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_renew", "true"),
				),
			},
		},
	})
}

func testAccRegisteredDomain_contacts(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainContactsConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_contact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.address_line_1", "99 High Street"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.address_line_2", "Flat 1a"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.city", "Little Nowhere"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.contact_type", "ASSOCIATION"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.country_code", "GB"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.email", "test1@example.com"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.fax", "+44.123456788"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.first_name", "Sys"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.last_name", "Admin"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.organization_name", "Support"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.phone_number", "+44.123456789"),
					resource.TestCheckResourceAttr(resourceName, "admin_contact.0.zip_code", "ST1 1AB"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.address_line_1", "100 Main Street"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.city", "New York City"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.contact_type", "COMPANY"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.email", "test2@example.com"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.first_name", "Terraform"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.last_name", "Team"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.organization_name", "HashiCorp"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.phone_number", "+1.2025551234"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.state", "NY"),
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.0.zip_code", "10001"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.address_line_1", "The Castle"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.city", "Prague"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.contact_type", "PERSON"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.country_code", "CZ"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.email", "test3@example.com"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.first_name", "Franz"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.last_name", "Kafka"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.phone_number", "+420.224372434"),
					resource.TestCheckResourceAttr(resourceName, "tech_contact.0.zip_code", "119 01"),
				),
			},
			{
				Config: testAccRegisteredDomainContactsUpdatedConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_contact.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "registrant_contact.#", "1"),
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
					resource.TestCheckResourceAttr(resourceName, "tech_contact.#", "1"),
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
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainContactPrivacyConfig(domainName, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", "true"),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", "true"),
				),
			},
			{
				Config: testAccRegisteredDomainContactPrivacyConfig(domainName, false, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "admin_privacy", "false"),
					resource.TestCheckResourceAttr(resourceName, "registrant_privacy", "false"),
					resource.TestCheckResourceAttr(resourceName, "tech_privacy", "false"),
				),
			},
		},
	})
}

func testAccRegisteredDomain_nameservers(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainNameserversConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_server.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.name", fmt.Sprintf("ns1.%s", domainName)),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.glue_ips.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "name_server.0.glue_ips.*", "1.1.1.1"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.name", "ns2.worldnic.com"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.glue_ips.#", "0"),
				),
			},
			{
				Config: testAccRegisteredDomainNameserversUpdatedConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_server.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.name", "ns-195.awsdns-24.com"),
					resource.TestCheckResourceAttr(resourceName, "name_server.0.glue_ips.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.name", "ns-1632.awsdns-12.co.uk"),
					resource.TestCheckResourceAttr(resourceName, "name_server.1.glue_ips.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name_server.2.name", "ns-874.awsdns-45.net"),
					resource.TestCheckResourceAttr(resourceName, "name_server.2.glue_ips.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name_server.3.name", "ns-1118.awsdns-11.org"),
					resource.TestCheckResourceAttr(resourceName, "name_server.3.glue_ips.#", "0"),
				),
			},
		},
	})
}

func testAccRegisteredDomain_transferLock(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, names.Route53DomainsEndpointID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRegisteredDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRegisteredDomainTransferLockConfig(domainName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", "false"),
				),
			},
			{
				Config: testAccRegisteredDomainTransferLockConfig(domainName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "transfer_lock", "true"),
				),
			},
		},
	})
}

func testAccCheckRegisteredDomainDestroy(s *terraform.State) error {
	return nil
}

func testAccRegisteredDomainConfigTags1(domainName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, domainName, tagKey1, tagValue1)
}

func testAccRegisteredDomainConfigTags2(domainName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccRegisteredDomainAutoRenewConfig(domainName string, autoRenew bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q
  auto_renew  = %[2]t
}
`, domainName, autoRenew)
}

func testAccRegisteredDomainContactsConfig(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  admin_contact {
    address_line_1    = "99 High Street"
    address_line_2    = "Flat 1a"
    city              = "Little Nowhere"
    contact_type      = "ASSOCIATION"
    country_code      = "GB"
    email             = "test1@example.com"
    fax               = "+44.123456788"
    first_name        = "Sys"
    last_name         = "Admin"
    organization_name = "Support"
    phone_number      = "+44.123456789"
    zip_code          = "ST1 1AB"
  }

  registrant_contact {
    address_line_1    = "100 Main Street"
    city              = "New York City"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = "test2@example.com"
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
    email          = "test3@example.com"
    first_name     = "Franz"
    last_name      = "Kafka"
    phone_number   = "+420.224372434"
    zip_code       = "119 01"
  }
}
`, domainName)
}

func testAccRegisteredDomainContactsUpdatedConfig(domainName string) string {
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

func testAccRegisteredDomainContactPrivacyConfig(domainName string, adminPrivacy, registrantPrivacy, techPrivacy bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  admin_privacy      = %[2]t
  registrant_privacy = %[3]t
  tech_privacy       = %[4]t
}
`, domainName, adminPrivacy, registrantPrivacy, techPrivacy)
}

func testAccRegisteredDomainNameserversConfig(domainName string) string {
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

func testAccRegisteredDomainNameserversUpdatedConfig(domainName string) string {
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

func testAccRegisteredDomainTransferLockConfig(domainName string, transferLock bool) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name   = %[1]q
  transfer_lock = %[2]t
}
`, domainName, transferLock)
}
