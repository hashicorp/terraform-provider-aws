package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAwsEc2ManagedPrefixListEntry_ipv4(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_ipv6(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "::/0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_expectInvalidTypeError(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryExpectInvalidType(rName),
				ExpectError: regexp.MustCompile(`invalid CIDR address: ::/244`),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_expectInvalidCIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryInvalidIPv4CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccAwsEc2ManagedPrefixListEntryInvalidIPv6CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_description(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_disappears(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2ManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSEc2ManagedPrefixListEntryExists(resourceName, &entry),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixListEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEc2ManagedPrefixListEntryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list_entry" {
			continue
		}

		plID, cidr, err := tfec2.ManagedPrefixListEntryParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.ManagedPrefixListEntryByIDAndCIDR(conn, plID, cidr)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Managed Prefix List Entry %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSEc2ManagedPrefixListEntryExists(n string, v *ec2.PrefixListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List Entry ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		plID, cidr, err := tfec2.ManagedPrefixListEntryParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := finder.ManagedPrefixListEntryByIDAndCIDR(conn, plID, cidr)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAwsEc2ManagedPrefixListEntryImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		plID := rs.Primary.Attributes["prefix_list_id"]
		cidr := rs.Primary.Attributes["cidr"]

		return tfec2.ManagedPrefixListEntryCreateID(plID, cidr), nil
	}
}

func testAccAwsEc2ManagedPrefixListEntryIpv4Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "10.0.0.0/8"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryIpv6Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv6"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "::/0"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "10.0.0.0/8"
  description    = %[1]q
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryExpectInvalidType(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "::/244"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryInvalidIPv4CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "1.2.3.4/33"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListEntryInvalidIPv6CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv6"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  cidr           = "::/244"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}
