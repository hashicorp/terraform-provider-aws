package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccAWSCloudSearchDomain_basic(t *testing.T) {
	var domains cloudsearch.DescribeDomainsOutput
	resourceName := "aws_cloudsearch_domain.test"
	rString := acctest.RandString(8)
	domainName := fmt.Sprintf("tf-acc-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudSearchDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudSearchDomainConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudSearchDomainExists(resourceName, &domains),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "search.m1.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       false,
				ImportStateVerify: false,
			},
		},
	})
}

func testAccCheckAWSCloudSearchDomainExists(n string, domains *cloudsearch.DescribeDomainsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudsearchconn

		domainlist := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		resp, err := conn.DescribeDomains(&domainlist)
		if err != nil {
			return err
		}

		*domains = *resp

		return nil
	}
}

func testAccCheckAWSCloudSearchDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudsearch_domain" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudsearchconn

		domainlist := cloudsearch.DescribeDomainsInput{
			DomainNames: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		resp, err := conn.DescribeDomains(&domainlist)
		if err != nil {
			return err
		}
		domain := resp.DomainStatusList[0]

		if domain.Deleted != nil || *domain.Deleted != true {
			return fmt.Errorf("Expected Cloudsearch domain to be deleted/destroyed, %s found with Deleted != true", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSCloudSearchDomainConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_cloudsearch_domain" "test" {
	domain_name   = "%s"
	wait_for_endpoints = "false"
	indexes {
		name            = "headline"
		type            = "text"
		search          = true
		return          = true
		sort            = true
		highlight       = false
		analysis_scheme = "_en_default_"
		}
	
	access_policy = <<EOF
	{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"AWS": [
			"*"
			]
		},
		"Action": [
			"cloudsearch:*"
		]
		}
	]
	}
	EOF
	}
`, domainName)
}
