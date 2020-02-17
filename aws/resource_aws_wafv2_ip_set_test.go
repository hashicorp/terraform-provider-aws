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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSWafv2IPSetImportStateIdFunc(resourceName),
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

func TestAccAwsWafv2IPSet_changeNameForceNew(t *testing.T) {
	var before, after wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", acctest.RandString(5))
	ipSetNewName := fmt.Sprintf("ip-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2IPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2IPSetConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
				),
			},
			{
				Config: testAccAwsWafv2IPSetConfig(ipSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetNewName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
				),
			},
		},
	})
}

func TestAccAwsWafv2IPSet_tags(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafv2IPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2IPSetConfigOneTag(ipSetName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAWSWafv2IPSetImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2IPSetConfigTwoTags(ipSetName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2IPSetConfigOneTag(ipSetName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafv2IPSetExists("aws_wafv2_ip_set.ip_set", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
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

func testAccAwsWafv2IPSetConfigOneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    %q = %q
  }
}
`, name, name, tagKey, tagValue)
}

func testAccAwsWafv2IPSetConfigTwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name = "%s"
  description = "%s"
  scope = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSWafv2IPSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
