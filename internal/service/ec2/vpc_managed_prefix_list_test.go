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

func TestAccVPCManagedPrefixList_basic(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_name(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "2"),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_disappears(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceManagedPrefixList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCManagedPrefixList_AddressFamily_ipv6(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_addressFamily(rName, "IPv6"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
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
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_entryCIDR1(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "Test2",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "3.0.0.0/8",
						"description": "Test3",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_entryDescription(rName, "description1"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "description1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "description1",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "description2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "description2",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "3"), // description-only updates require two operations
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_name(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccVPCManagedPrefixListConfig_name(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccVPCManagedPrefixList_tags(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckManagedPrefixList(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCManagedPrefixListConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCManagedPrefixListConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccVPCManagedPrefixListConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func testAccCheckManagedPrefixListDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list" {
			continue
		}

		_, err := tfec2.FindManagedPrefixListByID(conn, rs.Primary.ID)

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

func testAccManagedPrefixListExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := tfec2.FindManagedPrefixListByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccPreCheckManagedPrefixList(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeManagedPrefixListsInput{}

	_, err := conn.DescribeManagedPrefixLists(input)

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
