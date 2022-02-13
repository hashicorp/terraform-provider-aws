package route53domains_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRoute53Domains_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"RegisteredDomain": {
			"tags":        testAccRoute53DomainsRegisteredDomain_tags,
			"nameservers": testAccRoute53DomainsRegisteredDomain_nameservers,
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

func testAccPreCheckRoute53Domains(t *testing.T) {
	acctest.PreCheckPartitionHasService(route53domains.EndpointsID, t)

	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsConn

	input := &route53domains.ListDomainsInput{}

	_, err := conn.ListDomains(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRoute53DomainsRegisteredDomain_tags(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckRoute53Domains(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53domains.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegisteredDomainDestroy,
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

func testAccRoute53DomainsRegisteredDomain_nameservers(t *testing.T) {
	key := "ROUTE53DOMAINS_DOMAIN_NAME"
	domainName := os.Getenv(key)
	if domainName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resourceName := "aws_route53domains_registered_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckRoute53Domains(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53domains.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRegisteredDomainDestroy,
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

func testAccRegisteredDomainNameserversConfig(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_registered_domain" "test" {
  domain_name = %[1]q

  name_server {
    name     = "ns1.%[1]s"
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
