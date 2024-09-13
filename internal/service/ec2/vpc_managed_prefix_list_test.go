// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCManagedPrefixList_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_name(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "max_entries", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_updated(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_entries", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceManagedPrefixList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixList_AddressFamily_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_addressFamily(rName, "IPv6"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixList_Entry_cidr(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_entryCIDR1(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "1.0.0.0/8",
						names.AttrDescription: "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "2.0.0.0/8",
						names.AttrDescription: "Test2",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_entryCIDR2(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "1.0.0.0/8",
						names.AttrDescription: "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "3.0.0.0/8",
						names.AttrDescription: "Test3",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixList_Entry_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_entryDescription(rName, "description1"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "1.0.0.0/8",
						names.AttrDescription: "description1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "2.0.0.0/8",
						names.AttrDescription: "description1",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_entryDescription(rName, "description2"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "1.0.0.0/8",
						names.AttrDescription: "description2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":                "2.0.0.0/8",
						names.AttrDescription: "description2",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3), // description-only updates require two operations
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_updateEntryAndMaxEntry(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_entryMaxEntry(rName, 2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct2),
				),
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_entryMaxEntry(rName, 3),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_entryMaxEntry(rName, 1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_name(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_name(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccVPCManagedPrefixListConfig_name(rName2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedPrefixListDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCManagedPrefixListConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				Config: testAccVPCManagedPrefixListConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckManagedPrefixListDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_managed_prefix_list" {
				continue
			}

			_, err := tfec2.FindManagedPrefixListByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Managed Prefix List %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccManagedPrefixListExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindManagedPrefixListByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheckManagedPrefixList(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeManagedPrefixListsInput{}

	_, err := conn.DescribeManagedPrefixLists(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCManagedPrefixListConfig_addressFamily(rName string, addressFamily string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = %[2]q
  max_entries    = 1
  name           = %[1]q
}
`, rName, addressFamily)
}

func testAccVPCManagedPrefixListConfig_entryCIDR1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "2.0.0.0/8"
    description = "Test2"
  }
}
`, rName)
}

func testAccVPCManagedPrefixListConfig_entryCIDR2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "3.0.0.0/8"
    description = "Test3"
  }
}
`, rName)
}

func testAccVPCManagedPrefixListConfig_entryDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = %[2]q
  }

  entry {
    cidr        = "2.0.0.0/8"
    description = %[2]q
  }
}
`, rName, description)
}

func testAccVPCManagedPrefixListConfig_entryMaxEntry(rName string, maxEntryLength int) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = %[2]d
  name           = %[1]q

  dynamic entry {
    for_each = toset(slice(["1.0.0.0/8", "2.0.0.0/8", "3.0.0.0/8"], 0, %[2]d))

    content {
      cidr        = entry.key
      description = entry.key
    }
  }
}
`, rName, maxEntryLength)
}

func testAccVPCManagedPrefixListConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}
`, rName)
}

func testAccVPCManagedPrefixListConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 2
  name           = %[1]q
}
`, rName)
}

func testAccVPCManagedPrefixListConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCManagedPrefixListConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
