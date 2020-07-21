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
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsPrefixListEntry_basic(t *testing.T) {
	resourceName := "aws_prefix_list_entry.test"
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
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_basic_create,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
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
				Config:       testAccAwsPrefixListEntryConfig_basic_update,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
					resource.TestCheckResourceAttr(resourceName, "cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "Update"),
				),
			},
		},
	})
}

const testAccAwsPrefixListEntryConfig_basic_create = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
    description    = "Create"
}
`

const testAccAwsPrefixListEntryConfig_basic_update = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
    description    = "Update"
}
`

func testAccAwsPrefixListEntryExists(
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

func TestAccAwsPrefixListEntry_disappears(t *testing.T) {
	prefixListResourceName := "aws_prefix_list.test"
	resourceName := "aws_prefix_list_entry.test"
	pl := ec2.ManagedPrefixList{}
	entry := ec2.PrefixListEntry{}

	checkDisappears := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := ec2.ModifyManagedPrefixListInput{
			PrefixListId:   pl.PrefixListId,
			CurrentVersion: pl.Version,
			RemoveEntries: []*ec2.RemovePrefixListEntry{
				{
					Cidr: entry.Cidr,
				},
			},
		}

		_, err := conn.ModifyManagedPrefixList(&input)
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_disappears,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
					testAccAwsPrefixListExists(prefixListResourceName, &pl, nil),
					checkDisappears,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccAwsPrefixListEntryConfig_disappears = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
}
`

func TestAccAwsPrefixListEntry_prefixListDisappears(t *testing.T) {
	prefixListResourceName := "aws_prefix_list.test"
	resourceName := "aws_prefix_list_entry.test"
	pl := ec2.ManagedPrefixList{}
	entry := ec2.PrefixListEntry{}

	checkDisappears := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := ec2.DeleteManagedPrefixListInput{
			PrefixListId: pl.PrefixListId,
		}

		_, err := conn.DeleteManagedPrefixList(&input)
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_disappears,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
					testAccAwsPrefixListExists(prefixListResourceName, &pl, nil),
					checkDisappears,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsPrefixListEntry_alreadyExists(t *testing.T) {
	resourceName := "aws_prefix_list_entry.test"
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_alreadyExists,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
				),
				ExpectError: regexp.MustCompile(`an entry for this cidr block already exists`),
			},
		},
	})
}

const testAccAwsPrefixListEntryConfig_alreadyExists = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5

	entry {
		cidr_block = "1.0.0.0/8"
	}
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
	description    = "Test"
}
`

func TestAccAwsPrefixListEntry_description(t *testing.T) {
	resourceName := "aws_prefix_list_entry.test"
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
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_description_none,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
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
				Config:       testAccAwsPrefixListEntryConfig_description_some,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
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
				Config:       testAccAwsPrefixListEntryConfig_description_empty,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
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
				Config:       testAccAwsPrefixListEntryConfig_description_null,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry),
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

const testAccAwsPrefixListEntryConfig_description_none = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
	description    = "Test1"
}
`

const testAccAwsPrefixListEntryConfig_description_some = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
	description    = "Test2"
}
`

const testAccAwsPrefixListEntryConfig_description_empty = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
	description    = ""
}
`

const testAccAwsPrefixListEntryConfig_description_null = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}

resource "aws_prefix_list_entry" "test" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "1.0.0.0/8"
}
`

func TestAccAwsPrefixListEntry_exceedLimit(t *testing.T) {
	resourceName := "aws_prefix_list_entry.test_1"
	entry := ec2.PrefixListEntry{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_exceedLimit(2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListEntryExists(resourceName, &entry)),
			},
			{
				Config:       testAccAwsPrefixListEntryConfig_exceedLimit(3),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile(`You've reached the maximum number of entries for the prefix list.`),
			},
		},
	})
}

func testAccAwsPrefixListEntryConfig_exceedLimit(count int) string {
	entries := ``
	for i := 0; i < count; i++ {
		entries += fmt.Sprintf(`
resource "aws_prefix_list_entry" "test_%[1]d" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "%[1]d.0.0.0/8"
    description    = "Test_%[1]d"
}
`,
			i+1)
	}

	return fmt.Sprintf(`
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 2
}

%[1]s
`,
		entries)
}

func testAccAwsPrefixListSortEntries(list []*ec2.PrefixListEntry) {
	sort.Slice(list, func(i, j int) bool {
		return aws.StringValue(list[i].Cidr) < aws.StringValue(list[j].Cidr)
	})
}

func TestAccAwsPrefixListEntry_concurrentModification(t *testing.T) {
	prefixListResourceName := "aws_prefix_list.test"
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
			testAccAwsPrefixListSortEntries(expectEntries)

			testAccAwsPrefixListSortEntries(entries)

			if !reflect.DeepEqual(expectEntries, entries) {
				return fmt.Errorf("expected entries %#v, got %#v", expectEntries, entries)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListEntryConfig_concurrentModification("Step0_", 20),
				ResourceName: prefixListResourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(prefixListResourceName, &pl, &entries),
					checkAllEntriesExist("Step0_", 20)),
			},
			{
				// update the first 10 and drop the last 10
				Config:       testAccAwsPrefixListEntryConfig_concurrentModification("Step1_", 10),
				ResourceName: prefixListResourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(prefixListResourceName, &pl, &entries),
					checkAllEntriesExist("Step1_", 10)),
			},
		},
	})
}

func testAccAwsPrefixListEntryConfig_concurrentModification(prefix string, count int) string {
	entries := ``
	for i := 0; i < count; i++ {
		entries += fmt.Sprintf(`
resource "aws_prefix_list_entry" "test_%[1]d" {
	prefix_list_id = aws_prefix_list.test.id
	cidr_block     = "%[1]d.0.0.0/8"
    description    = "%[2]s%[1]d"
}
`,
			i+1,
			prefix)
	}

	return fmt.Sprintf(`
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 20
}

%[1]s
`,
		entries)
}
