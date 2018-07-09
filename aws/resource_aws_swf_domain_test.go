package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsSwfDomain_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSwfDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsSwfDomainConfig, "test_swf_domain_"),
				Check:  testAccCheckAwsSwfDomainExists("aws_swf_domain.test"),
			},
		},
	})
}

func testAccCheckAwsSwfDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).swfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_swf_domain" {
			continue
		}

		name := rs.Primary.ID
		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return err
		}

		if *resp.DomainInfo.Status != "DEPRECATED" {
			return fmt.Errorf(`SWF Domain %s status is %s instead of "DEPRECATED". Failing!`, name, *resp.DomainInfo.Status)
		}
	}

	return nil
}

func testAccCheckAwsSwfDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SWF Domain not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SWF Domain name not set")
		}

		name := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).swfconn

		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return fmt.Errorf("SWF Domain %s not found in AWS", name)
		}

		if *resp.DomainInfo.Status != "REGISTERED" {
			return fmt.Errorf(`SWF Domain %s status is %s instead of "REGISTERED". Failing!`, name, *resp.DomainInfo.Status)
		}
		return nil
	}
}

const testAccAwsSwfDomainConfig = `
resource "aws_swf_domain" "test" {
	name_prefix                                 = "%s"
	workflow_execution_retention_period_in_days = 1
}
`
