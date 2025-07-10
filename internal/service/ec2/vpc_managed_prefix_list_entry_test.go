// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCManagedPrefixListEntry_ipv4(t *testing.T) {
	ctx := acctest.Context(t)
	var entry awstypes.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListEntryConfig_ipv4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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

func TestAccVPCManagedPrefixListEntry_ipv4Multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var entry awstypes.PrefixListEntry
	resourceName1 := "aws_ec2_managed_prefix_list_entry.test1"
	resourceName2 := "aws_ec2_managed_prefix_list_entry.test2"
	resourceName3 := "aws_ec2_managed_prefix_list_entry.test3"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListEntryConfig_ipv4Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName1, &entry),
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName2, &entry),
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName3, &entry),
					resource.TestCheckResourceAttr(resourceName1, "cidr", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(resourceName2, "cidr", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName3, "cidr", "10.0.2.0/24"),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var entry awstypes.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListEntryConfig_ipv6(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cidr", "::/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCManagedPrefixListEntryConfig_expectInvalidType(rName),
				ExpectError: regexache.MustCompile(`invalid CIDR address: ::/244`),
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_expectInvalidCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCManagedPrefixListEntryConfig_invalidIPv4CIDR(rName),
				ExpectError: regexache.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccVPCManagedPrefixListEntryConfig_invalidIPv6CIDR(rName),
				ExpectError: regexache.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

func TestAccVPCManagedPrefixListEntry_description(t *testing.T) {
	ctx := acctest.Context(t)
	var entry awstypes.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListEntryConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName, &entry),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
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
	ctx := acctest.Context(t)
	var entry awstypes.PrefixListEntry
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListEntryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListEntryConfig_ipv4(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedPrefixListEntryExists(ctx, resourceName, &entry),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceManagedPrefixListEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckManagedPrefixListEntryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_managed_prefix_list_entry" {
				continue
			}

			plID, cidr, err := tfec2.ManagedPrefixListEntryParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfec2.FindManagedPrefixListEntryByIDAndCIDR(ctx, conn, plID, cidr)

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
}

func testAccCheckManagedPrefixListEntryExists(ctx context.Context, n string, v *awstypes.PrefixListEntry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List Entry ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		plID, cidr, err := tfec2.ManagedPrefixListEntryParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfec2.FindManagedPrefixListEntryByIDAndCIDR(ctx, conn, plID, cidr)

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

		return tfec2.ManagedPrefixListEntryCreateResourceID(plID, cidr), nil
	}
}

func testAccVPCManagedPrefixListEntryConfig_ipv4(rName string) string {
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

func testAccVPCManagedPrefixListEntryConfig_ipv4Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test1" {
  cidr           = "10.0.0.0/24"
  description    = "description 1"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}

resource "aws_ec2_managed_prefix_list_entry" "test2" {
  cidr           = "10.0.1.0/24"
  description    = "description 2"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}

resource "aws_ec2_managed_prefix_list_entry" "test3" {
  cidr           = "10.0.2.0/24"
  description    = "description 3"
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
}
`, rName)
}

func testAccVPCManagedPrefixListEntryConfig_ipv6(rName string) string {
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

func testAccVPCManagedPrefixListEntryConfig_description(rName string) string {
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

func testAccVPCManagedPrefixListEntryConfig_expectInvalidType(rName string) string {
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

func testAccVPCManagedPrefixListEntryConfig_invalidIPv4CIDR(rName string) string {
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

func testAccVPCManagedPrefixListEntryConfig_invalidIPv6CIDR(rName string) string {
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
