package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsEc2ManagedPrefixListEntry_basic(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	entry := ec2.PrefixListEntry{}

	checkAttributes := func(*terraform.State) error {
		if actual := aws.StringValue(entry.Cidr); actual != "1.0.0.0/8" {
			return fmt.Errorf("bad cidr: %s", actual)
		}

		if actual := aws.StringValue(entry.Description); actual != "Create" {
			return fmt.Errorf("bad description: %s", actual)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_basic_create,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					checkAttributes,
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "Create"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_basic_update,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "Update"),
				),
			},
		},
	})
}

const testAccAwsEc2ManagedPrefixListEntryConfig_basic_create = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = "Create"
}
`

const testAccAwsEc2ManagedPrefixListEntryConfig_basic_update = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = "Update"
}
`

func testAccAwsEc2ManagedPrefixListEntryExists(
	name string,
	out *ec2.PrefixListEntry,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		switch {
		case !ok:
			return fmt.Errorf("resource %s not found", name)
		case rs.Primary.ID == "":
			return fmt.Errorf("resource %s has not set its id", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		ss := strings.Split(rs.Primary.ID, "_")
		prefixListId, cidrBlock := ss[0], ss[1]

		entry, ok, err := getManagedPrefixListEntryByCIDR(prefixListId, conn, 0, cidrBlock)
		switch {
		case err != nil:
			return err
		case !ok:
			return fmt.Errorf("resource %s (%s) has not been created", name, prefixListId)
		}

		if out != nil {
			*out = *entry
		}

		return nil
	}
}

func TestAccAwsEc2ManagedPrefixListEntry_disappears(t *testing.T) {
	prefixListResourceName := "aws_ec2_managed_prefix_list.test"
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	pl := ec2.ManagedPrefixList{}
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_disappears,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					testAccAwsEc2ManagedPrefixListExists(prefixListResourceName, &pl, nil),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixListEntry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccAwsEc2ManagedPrefixListEntryConfig_disappears = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
}
`

func TestAccAwsEc2ManagedPrefixListEntry_prefixListDisappears(t *testing.T) {
	prefixListResourceName := "aws_ec2_managed_prefix_list.test"
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	pl := ec2.ManagedPrefixList{}
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_disappears,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					testAccAwsEc2ManagedPrefixListExists(prefixListResourceName, &pl, nil),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixList(), prefixListResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_alreadyExists(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_alreadyExists,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
				),
				ExpectError: regexp.MustCompile(`an entry for this cidr block already exists`),
			},
		},
	})
}

const testAccAwsEc2ManagedPrefixListEntryConfig_alreadyExists = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr_block = "1.0.0.0/8"
  }
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = "Test"
}
`

func TestAccAwsEc2ManagedPrefixListEntry_description(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list_entry.test"
	entry := ec2.PrefixListEntry{}

	checkDescription := func(expect string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if actual := aws.StringValue(entry.Description); actual != expect {
				return fmt.Errorf("bad description: %s", actual)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_description_none,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					checkDescription("Test1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_description_some,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					checkDescription("Test2"),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_description_empty,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					checkDescription(""),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_description_null,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry),
					checkDescription(""),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

const testAccAwsEc2ManagedPrefixListEntryConfig_description_none = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = "Test1"
}
`

const testAccAwsEc2ManagedPrefixListEntryConfig_description_some = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = "Test2"
}
`

const testAccAwsEc2ManagedPrefixListEntryConfig_description_empty = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
  description    = ""
}
`

const testAccAwsEc2ManagedPrefixListEntryConfig_description_null = `
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list_entry" "test" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "1.0.0.0/8"
}
`

func TestAccAwsEc2ManagedPrefixListEntry_exceedLimit(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list_entry.test_1"
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_exceedLimit(2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListEntryExists(resourceName, &entry)),
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_exceedLimit(3),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile(`You've reached the maximum number of entries for the prefix list.`),
			},
		},
	})
}

func testAccAwsEc2ManagedPrefixListEntryConfig_exceedLimit(count int) string {
	entries := ``
	for i := 0; i < count; i++ {
		entries += fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list_entry" "test_%[1]d" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "%[1]d.0.0.0/8"
  description    = "Test_%[1]d"
}
`,
			i+1)
	}

	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 2
}

%[1]s
`,
		entries)
}

func testAccAwsEc2ManagedPrefixListSortEntries(list []*ec2.PrefixListEntry) {
	sort.Slice(list, func(i, j int) bool {
		return aws.StringValue(list[i].Cidr) < aws.StringValue(list[j].Cidr)
	})
}

func TestAccAwsEc2ManagedPrefixListEntry_concurrentModification(t *testing.T) {
	prefixListResourceName := "aws_ec2_managed_prefix_list.test"
	pl, entries := ec2.ManagedPrefixList{}, []*ec2.PrefixListEntry(nil)

	checkAllEntriesExist := func(prefix string, count int) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if len(entries) != count {
				return fmt.Errorf("expected %d entries", count)
			}

			expectEntries := make([]*ec2.PrefixListEntry, 0, count)
			for i := 0; i < count; i++ {
				expectEntries = append(expectEntries, &ec2.PrefixListEntry{
					Cidr:        aws.String(fmt.Sprintf("%d.0.0.0/8", i+1)),
					Description: aws.String(fmt.Sprintf("%s%d", prefix, i+1))})
			}
			testAccAwsEc2ManagedPrefixListSortEntries(expectEntries)

			testAccAwsEc2ManagedPrefixListSortEntries(entries)

			if !reflect.DeepEqual(expectEntries, entries) {
				return fmt.Errorf("expected entries %#v, got %#v", expectEntries, entries)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_concurrentModification("Step0_", 20),
				ResourceName: prefixListResourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(prefixListResourceName, &pl, &entries),
					checkAllEntriesExist("Step0_", 20)),
			},
			{
				// update the first 10 and drop the last 10
				Config:       testAccAwsEc2ManagedPrefixListEntryConfig_concurrentModification("Step1_", 10),
				ResourceName: prefixListResourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(prefixListResourceName, &pl, &entries),
					checkAllEntriesExist("Step1_", 10)),
			},
		},
	})
}

func testAccAwsEc2ManagedPrefixListEntryConfig_concurrentModification(prefix string, count int) string {
	entries := ``
	for i := 0; i < count; i++ {
		entries += fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list_entry" "test_%[1]d" {
  prefix_list_id = aws_ec2_managed_prefix_list.test.id
  cidr_block     = "%[1]d.0.0.0/8"
  description    = "%[2]s%[1]d"
}
`,
			i+1,
			prefix)
	}

	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = "tf-test-acc"
  address_family = "IPv4"
  max_entries    = 20
}

%[1]s
`,
		entries)
}
