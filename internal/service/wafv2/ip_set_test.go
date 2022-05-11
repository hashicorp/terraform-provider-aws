package wafv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
)

func TestAccWAFV2IPSet_basic(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
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
				Config: testAccIPSetUpdateConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
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
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_disappears(t *testing.T) {
	var r wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &r),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafv2.ResourceIPSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2IPSet_ipv6(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetIPv6Config(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv6),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "1234:5678:9abc:6811:0000:0000:0000:0000/64"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "2001:db8::/32"),
					resource.TestCheckTypeSetElemAttr(resourceName, "addresses.*", "1111:0000:0000:0000:0000:0000:0000:0111/128"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_minimal(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetMinimalConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2IPSet_changeNameForceNew(t *testing.T) {
	var before, after wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	ipSetNewName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "2"),
				),
			},
			{
				Config: testAccIPSetConfig(ipSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
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

func TestAccWAFV2IPSet_tags(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetOneTagConfig(ipSetName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
			{
				Config: testAccIPSetTwoTagsConfig(ipSetName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccIPSetOneTagConfig(ipSetName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func TestAccWAFV2IPSet_large(t *testing.T) {
	var v wafv2.IPSet
	ipSetName := fmt.Sprintf("ip-set-%s", sdkacctest.RandString(5))
	resourceName := "aws_wafv2_ip_set.ip_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSetLargeConfig(ipSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPSetExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/ipset/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "description", ipSetName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "ip_address_version", wafv2.IPAddressVersionIpv4),
					resource.TestCheckResourceAttr(resourceName, "addresses.#", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccIPSetImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccCheckIPSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_ip_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetIPSet(
			&wafv2.GetIPSetInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp == nil || resp.IPSet == nil {
				return fmt.Errorf("Error getting WAFv2 IPSet")
			}
			if aws.StringValue(resp.IPSet.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 IPSet %s still exists", rs.Primary.ID)
			}
			return nil
		}

		// Return nil if the IPSet is already destroyed
		if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckIPSetExists(n string, v *wafv2.IPSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 IPSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetIPSet(&wafv2.GetIPSetInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.IPSet == nil {
			return fmt.Errorf("Error getting WAFv2 IPSet")
		}

		if aws.StringValue(resp.IPSet.Id) == rs.Primary.ID {
			*v = *resp.IPSet
			return nil
		}

		return fmt.Errorf("WAFv2 IPSet (%s) not found", rs.Primary.ID)
	}
}

func testAccIPSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
`, name, name)
}

func testAccIPSetUpdateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "Updated"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32", "3.3.3.3/32"]
}
`, name)
}

func testAccIPSetIPv6Config(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV6"
  addresses = [
    "1111:0000:0000:0000:0000:0000:0000:0111/128",
    "1234:5678:9abc:6811:0000:0000:0000:0000/64",
    "2001:db8::/32"
  ]
}
`, name, name)
}

func testAccIPSetMinimalConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
}
`, name)
}

func testAccIPSetOneTagConfig(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    "%s" = "%s"
  }
}
`, name, name, tagKey, tagValue)
}

func testAccIPSetTwoTagsConfig(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccIPSetLargeConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "ip_set" {
  name               = "%s"
  description        = "%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses = [
    "1.1.1.50/32", "1.1.1.73/32", "1.1.1.15/32", "2.2.2.30/32", "1.1.1.38/32",
    "2.2.2.53/32", "1.1.1.21/32", "2.2.2.24/32", "1.1.1.44/32", "1.1.1.1/32",
    "1.1.1.67/32", "2.2.2.76/32", "2.2.2.99/32", "1.1.1.26/32", "2.2.2.93/32",
    "2.2.2.64/32", "1.1.1.32/32", "2.2.2.12/32", "2.2.2.47/32", "1.1.1.91/32",
    "1.1.1.78/32", "2.2.2.82/32", "2.2.2.58/32", "1.1.1.85/32", "2.2.2.4/32",
    "2.2.2.65/32", "2.2.2.23/32", "2.2.2.17/32", "2.2.2.42/32", "1.1.1.56/32",
    "1.1.1.79/32", "2.2.2.81/32", "2.2.2.36/32", "2.2.2.59/32", "2.2.2.9/32",
    "1.1.1.7/32", "1.1.1.84/32", "1.1.1.51/32", "2.2.2.70/32", "2.2.2.87/32",
    "1.1.1.39/32", "1.1.1.90/32", "2.2.2.31/32", "1.1.1.62/32", "1.1.1.14/32",
    "1.1.1.20/32", "2.2.2.25/32", "1.1.1.45/32", "1.1.1.2/32", "2.2.2.98/32"
  ]
}
`, name, name)
}

func testAccIPSetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
