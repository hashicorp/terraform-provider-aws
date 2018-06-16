package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLightsailDomainEntry_basic(t *testing.T) {
	var domainEntry lightsail.DomainEntry
	domainName := fmt.Sprintf("tf-lightsail-domain-%s.com", acctest.RandString(5))
	entryName := fmt.Sprintf("tf-lightsail-domain-entry-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLightsailDomainEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailDomainEntryConfig(domainName, entryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailDomainEntryExists("aws_lightsail_domain_entry.example", &domainEntry),
				),
			},
		},
	})
}

func testAccCheckAWSLightsailDomainEntryExists(n string, domainEntry *lightsail.DomainEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Lightsail Domain Entry ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		domainName := rs.Primary.Attributes["domain_name"]

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(domainName),
		})
		if err != nil {
			return err
		}

		if resp == nil || resp.Domain == nil {
			return fmt.Errorf("Domain (%s) not found", domainName)
		}

		for _, entry := range resp.Domain.DomainEntries {
			if *entry.Name == rs.Primary.ID {
				return nil
			}
		}
		return fmt.Errorf("No entry exists in domain %s", domainName)
	}
}

func testAccCheckAWSLightsailDomainEntryDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lightsail_domain_entry" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
		domainName := rs.Primary.Attributes["domain_name"]

		resp, err := conn.GetDomain(&lightsail.GetDomainInput{
			DomainName: aws.String(domainName),
		})

		if err == nil && resp.Domain != nil {
			for _, entry := range resp.Domain.DomainEntries {
				if *entry.Name == rs.Primary.ID {
					return fmt.Errorf("Lightsail domain entry %s still exists", rs.Primary.ID)
				}
			}
			return nil
		}

		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccAWSLightsailDomainEntryConfig(domainName, entryName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_lightsail_domain" "example" {
	domain_name = "%s"
}

resource "aws_lightsail_domain_entry" "example" {
	domain_name = "${aws_lightsail_domain.example.id}"
	domain_entry = {
		name = "%s.${aws_lightsail_domain.example.id}"
		target = "1.2.3.4"
		type = "A"
	}
}
`, domainName, entryName)
}
