package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAwsWafv2IPSet_basic(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2IPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2IPSetConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
				),
			},
			{
				Config: testAccAwsWafv2IPSetConfigUpdate(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
				),
			},
		},
	})
}

func TestAccAwsWafv2IPSet_minimal(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2IPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2IPSetConfigMinimal(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafv2IPSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_ip_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetIPSet(
			&wafv2.GetIPSetInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if *resp.IPSet.Id == rs.Primary.ID {
				return fmt.Errorf("WAFV2 IPSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the IPSet is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafv2IPSetExists(n string, v *wafv2.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFV2 IPSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetIPSet(&wafv2.GetIPSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if *resp.IPSet.Id == rs.Primary.ID {
			*v = *resp.IPSet
			return nil
		}

		return fmt.Errorf("WAFV2 IPSet (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2IPSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = ["1.2.3.4/32", "5.6.7.8/32"]
}
`, name, name)
}

func testAccAwsWafv2IPSetConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name = "%s"
  description = "Updated"
  scope = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = ["1.1.1.1/32", "2.2.2.2/32", "3.3.3.3/32"]
}
`, name)
}

func testAccAwsWafv2IPSetConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name = "%s"
  scope = "REGIONAL"
  ip_address_version = "IPV4"
}
`, name)
}
