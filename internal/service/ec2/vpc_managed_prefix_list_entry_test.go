package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCManagedPrefixListEntry_ipv4(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPrefixListEntryIPv4Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_ipv6(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPrefixListEntryIPv6Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "::/0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_expectInvalidTypeError(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccManagedPrefixListEntryExpectInvalidType(rName),
				ExpectError: regexp.MustCompile(`invalid CIDR address: ::/244`),
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_expectInvalidCIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccManagedPrefixListEntryInvalidIPv4CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccManagedPrefixListEntryInvalidIPv6CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_description(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPrefixListEntryDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccManagedPrefixListEntryImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_disappears(t *testing.T) {
	var entry ec2.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListEntryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccManagedPrefixListEntryIPv4Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(resourceName, &entry),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceManagedPrefixListEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckManagedPrefixListEntryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list_entry" {
			continue
		}

		plID, cidr, err := tfec2.ManagedPrefixListEntryParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindManagedPrefixListEntryByIDAndCIDR(conn, plID, cidr)

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

func testAccCheckManagedPrefixListEntryExists(n string, v *ec2.PrefixListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List Entry ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		plID, cidr, err := tfec2.ManagedPrefixListEntryParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfec2.FindManagedPrefixListEntryByIDAndCIDR(conn, plID, cidr)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccManagedPrefixListEntryImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
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

func testAccManagedPrefixListEntryIPv4Config(rName string) string {
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

func testAccManagedPrefixListEntryIPv6Config(rName string) string {
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

func testAccManagedPrefixListEntryDescriptionConfig(rName string) string {
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

func testAccManagedPrefixListEntryExpectInvalidType(rName string) string {
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

func testAccManagedPrefixListEntryInvalidIPv4CIDR(rName string) string {
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

func testAccManagedPrefixListEntryInvalidIPv6CIDR(rName string) string {
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
