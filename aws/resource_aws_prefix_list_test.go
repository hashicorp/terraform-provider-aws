package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsPrefixList_computePrefixListEntriesModification(t *testing.T) {
	type testEntry struct {
		CIDR        string
		Description string
	}

	tests := []struct {
		name            string
		oldEntries      []testEntry
		newEntries      []testEntry
		expectedAdds    []testEntry
		expectedRemoves []testEntry
	}{
		{
			name:            "add two",
			oldEntries:      []testEntry{},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			expectedAdds:    []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			expectedRemoves: []testEntry{},
		},
		{
			name:            "remove one",
			oldEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1"}},
			expectedAdds:    []testEntry{},
			expectedRemoves: []testEntry{{"2.3.4.5/32", "test2"}},
		},
		{
			name:            "modify description of one",
			oldEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2-1"}},
			expectedAdds:    []testEntry{{"2.3.4.5/32", "test2-1"}},
			expectedRemoves: []testEntry{},
		},
		{
			name:            "add third",
			oldEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}, {"3.4.5.6/32", "test3"}},
			expectedAdds:    []testEntry{{"3.4.5.6/32", "test3"}},
			expectedRemoves: []testEntry{},
		},
		{
			name:            "add and remove one",
			oldEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"3.4.5.6/32", "test3"}},
			expectedAdds:    []testEntry{{"3.4.5.6/32", "test3"}},
			expectedRemoves: []testEntry{{"2.3.4.5/32", "test2"}},
		},
		{
			name:            "add and remove one with description change",
			oldEntries:      []testEntry{{"1.2.3.4/32", "test1"}, {"2.3.4.5/32", "test2"}},
			newEntries:      []testEntry{{"1.2.3.4/32", "test1-1"}, {"3.4.5.6/32", "test3"}},
			expectedAdds:    []testEntry{{"1.2.3.4/32", "test1-1"}, {"3.4.5.6/32", "test3"}},
			expectedRemoves: []testEntry{{"2.3.4.5/32", "test2"}},
		},
		{
			name:            "basic test update",
			oldEntries:      []testEntry{{"1.0.0.0/8", "Test1"}},
			newEntries:      []testEntry{{"1.0.0.0/8", "Test1-1"}, {"2.2.0.0/16", "Test2"}},
			expectedAdds:    []testEntry{{"1.0.0.0/8", "Test1-1"}, {"2.2.0.0/16", "Test2"}},
			expectedRemoves: []testEntry{},
		},
	}

	for _, test := range tests {
		oldEntryList := []*ec2.PrefixListEntry(nil)
		for _, entry := range test.oldEntries {
			oldEntryList = append(oldEntryList, &ec2.PrefixListEntry{
				Cidr:        aws.String(entry.CIDR),
				Description: aws.String(entry.Description),
			})
		}

		newEntryList := []*ec2.AddPrefixListEntry(nil)
		for _, entry := range test.newEntries {
			newEntryList = append(newEntryList, &ec2.AddPrefixListEntry{
				Cidr:        aws.String(entry.CIDR),
				Description: aws.String(entry.Description),
			})
		}

		addList, removeList := computePrefixListEntriesModification(oldEntryList, newEntryList)

		if len(addList) != len(test.expectedAdds) {
			t.Errorf("expected %d adds, got %d", len(test.expectedAdds), len(addList))
		}

		for i, added := range addList {
			expected := test.expectedAdds[i]

			actualCidr := aws.StringValue(added.Cidr)
			expectedCidr := expected.CIDR
			if actualCidr != expectedCidr {
				t.Errorf("add[%d]: expected cidr %s, got %s", i, expectedCidr, actualCidr)
			}

			actualDesc := aws.StringValue(added.Description)
			expectedDesc := expected.Description
			if actualDesc != expectedDesc {
				t.Errorf("add[%d]: expected description '%s', got '%s'", i, expectedDesc, actualDesc)
			}
		}

		if len(removeList) != len(test.expectedRemoves) {
			t.Errorf("expected %d removes, got %d", len(test.expectedRemoves), len(removeList))
		}

		for i, removed := range removeList {
			expected := test.expectedRemoves[i]

			actualCidr := aws.StringValue(removed.Cidr)
			expectedCidr := expected.CIDR
			if actualCidr != expectedCidr {
				t.Errorf("add[%d]: expected cidr %s, got %s", i, expectedCidr, actualCidr)
			}
		}
	}
}

func testAccCheckAWSPrefixListDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_prefix_list" {
			continue
		}

		id := rs.Primary.ID

		switch _, ok, err := getManagedPrefixList(id, conn); {
		case err != nil:
			return err
		case ok:
			return fmt.Errorf("managed prefix list %s still exists", id)
		}
	}

	return nil
}

func testAccCheckAwsPrefixListVersion(
	prefixList *ec2.ManagedPrefixList,
	version int64,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if actual := aws.Int64Value(prefixList.Version); actual != version {
			return fmt.Errorf("expected prefix list version %d, got %d", version, actual)
		}

		return nil
	}
}

func TestAccAwsPrefixList_basic(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	pl, entries := ec2.ManagedPrefixList{}, []*ec2.PrefixListEntry(nil)

	checkAttributes := func(*terraform.State) error {
		if actual := aws.StringValue(pl.AddressFamily); actual != "IPv4" {
			return fmt.Errorf("bad address family: %s", actual)
		}

		if actual := aws.Int64Value(pl.MaxEntries); actual != 5 {
			return fmt.Errorf("bad max entries: %d", actual)
		}

		if actual := aws.StringValue(pl.OwnerId); actual != testAccGetAccountID() {
			return fmt.Errorf("bad owner id: %s", actual)
		}

		if actual := aws.StringValue(pl.PrefixListName); actual != "tf-test-basic-create" {
			return fmt.Errorf("bad name: %s", actual)
		}

		sort.Slice(pl.Tags, func(i, j int) bool {
			return aws.StringValue(pl.Tags[i].Key) < aws.StringValue(pl.Tags[j].Key)
		})

		expectTags := []*ec2.Tag{
			{Key: aws.String("Key1"), Value: aws.String("Value1")},
			{Key: aws.String("Key2"), Value: aws.String("Value2")},
		}

		if !reflect.DeepEqual(expectTags, pl.Tags) {
			return fmt.Errorf("expected tags %#v, got %#v", expectTags, pl.Tags)
		}

		sort.Slice(entries, func(i, j int) bool {
			return aws.StringValue(entries[i].Cidr) < aws.StringValue(entries[j].Cidr)
		})

		expectEntries := []*ec2.PrefixListEntry{
			{Cidr: aws.String("1.0.0.0/8"), Description: aws.String("Test1")},
			{Cidr: aws.String("2.0.0.0/8"), Description: aws.String("Test2")},
		}

		if !reflect.DeepEqual(expectEntries, entries) {
			return fmt.Errorf("expected entries %#v, got %#v", expectEntries, entries)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListConfig_basic_create,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, &entries),
					checkAttributes,
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-basic-create"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "5"),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "entry.3370291439.cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.3370291439.description", "Test1"),
					resource.TestCheckResourceAttr(resourceName, "entry.3776037899.cidr_block", "2.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.3776037899.description", "Test2"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccCheckAwsPrefixListVersion(&pl, 1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_basic_update,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, &entries),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-basic-update"),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "entry.3370291439.cidr_block", "1.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.3370291439.description", "Test1"),
					resource.TestCheckResourceAttr(resourceName, "entry.4190046295.cidr_block", "3.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.4190046295.description", "Test3"),
					testAccCheckAwsPrefixListVersion(&pl, 2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

const testAccAwsPrefixListConfig_basic_create = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-basic-create"
	address_family = "IPv4"
	max_entries    = 5

	entry {
		cidr_block  = "1.0.0.0/8"
		description = "Test1"
	}

	entry {
		cidr_block  = "2.0.0.0/8"
		description = "Test2"
	}

	tags = {
		Key1 = "Value1"
		Key2 = "Value2"
	}
}
`

const testAccAwsPrefixListConfig_basic_update = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-basic-update"
	address_family = "IPv4"
	max_entries    = 5

	entry {
		cidr_block  = "1.0.0.0/8"
		description = "Test1"
	}

	entry {
		cidr_block  = "3.0.0.0/8"
		description = "Test3"
	}

	tags = {
		Key1 = "Value1"
		Key3 = "Value3"
	}
}
`

func testAccAwsPrefixListExists(
	name string,
	out *ec2.ManagedPrefixList,
	entries *[]*ec2.PrefixListEntry,
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
		id := rs.Primary.ID

		pl, ok, err := getManagedPrefixList(id, conn)
		switch {
		case err != nil:
			return err
		case !ok:
			return fmt.Errorf("resource %s (%s) has not been created", name, id)
		}

		if out != nil {
			*out = *pl
		}

		if entries != nil {
			entries1, err := getPrefixListEntries(id, conn, *pl.Version)
			if err != nil {
				return err
			}

			*entries = entries1
		}

		return nil
	}
}

func TestAccAwsPrefixList_disappears(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	pl := ec2.ManagedPrefixList{}

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
				Config:       testAccAwsPrefixListConfig_disappears,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkDisappears,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccAwsPrefixListConfig_disappears = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-disappears"
	address_family = "IPv4"
	max_entries    = 2

	entry {
		cidr_block = "1.0.0.0/8"
	}
}
`

func TestAccAwsPrefixList_name(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	pl := ec2.ManagedPrefixList{}

	checkName := func(name string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if actual := aws.StringValue(pl.PrefixListName); actual != name {
				return fmt.Errorf("expected name %s, got %s", name, actual)
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
				Config:       testAccAwsPrefixListConfig_name_create,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-name-create"),
					checkName("tf-test-name-create"),
					testAccCheckAwsPrefixListVersion(&pl, 1),
				),
			},
			{
				Config:       testAccAwsPrefixListConfig_name_update,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-test-name-update"),
					checkName("tf-test-name-update"),
					testAccCheckAwsPrefixListVersion(&pl, 1),
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

const testAccAwsPrefixListConfig_name_create = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-name-create"
	address_family = "IPv4"
	max_entries    = 5
}
`

const testAccAwsPrefixListConfig_name_update = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-name-update"
	address_family = "IPv4"
	max_entries    = 5
}
`

func TestAccAwsPrefixList_tags(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	pl := ec2.ManagedPrefixList{}

	checkTags := func(m map[string]string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			sort.Slice(pl.Tags, func(i, j int) bool {
				return aws.StringValue(pl.Tags[i].Key) < aws.StringValue(pl.Tags[j].Key)
			})

			expectTags := []*ec2.Tag(nil)

			if m != nil {
				for k, v := range m {
					expectTags = append(expectTags, &ec2.Tag{
						Key:   aws.String(k),
						Value: aws.String(v),
					})
				}

				sort.Slice(expectTags, func(i, j int) bool {
					return aws.StringValue(expectTags[i].Key) < aws.StringValue(expectTags[j].Key)
				})
			}

			if !reflect.DeepEqual(expectTags, pl.Tags) {
				return fmt.Errorf("expected tags %#v, got %#v", expectTags, pl.Tags)
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
				Config:       testAccAwsPrefixListConfig_tags_none,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkTags(nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_tags_addSome,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkTags(map[string]string{"Key1": "Value1", "Key2": "Value2", "Key3": "Value3"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_tags_dropOrModifySome,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkTags(map[string]string{"Key2": "Value2-1", "Key3": "Value3"}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2-1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_tags_empty,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkTags(nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_tags_none,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &pl, nil),
					checkTags(nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

const testAccAwsPrefixListConfig_tags_none = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
}
`

const testAccAwsPrefixListConfig_tags_addSome = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5

	tags = {
		Key1 = "Value1"
		Key2 = "Value2"
		Key3 = "Value3"
	}
}
`

const testAccAwsPrefixListConfig_tags_dropOrModifySome = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5

	tags = {
		Key2 = "Value2-1"
		Key3 = "Value3"
	}
}
`

const testAccAwsPrefixListConfig_tags_empty = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 5
	
	tags = {}
}
`

func TestAccAwsPrefixList_entryConfigMode(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	prefixList := ec2.ManagedPrefixList{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListConfig_entryConfigMode_blocks,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &prefixList, nil),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsPrefixListConfig_entryConfigMode_noBlocks,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &prefixList, nil),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
				),
			},
			{
				Config:       testAccAwsPrefixListConfig_entryConfigMode_zeroed,
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &prefixList, nil),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "0"),
				),
			},
		},
	})
}

const testAccAwsPrefixListConfig_entryConfigMode_blocks = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	max_entries    = 5
	address_family = "IPv4"

	entry {
		cidr_block  = "1.0.0.0/8"
		description = "Entry1"
	}

	entry {
		cidr_block  = "2.0.0.0/8"
		description = "Entry2"
	}
}
`

const testAccAwsPrefixListConfig_entryConfigMode_noBlocks = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	max_entries    = 5
	address_family = "IPv4"
}
`

const testAccAwsPrefixListConfig_entryConfigMode_zeroed = `
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	max_entries    = 5
	address_family = "IPv4"
	entry          = []
}
`

func TestAccAwsPrefixList_exceedLimit(t *testing.T) {
	resourceName := "aws_prefix_list.test"
	prefixList := ec2.ManagedPrefixList{}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsPrefixListConfig_exceedLimit(2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsPrefixListExists(resourceName, &prefixList, nil),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
				),
			},
			{
				Config:       testAccAwsPrefixListConfig_exceedLimit(3),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile(`You've reached the maximum number of entries for the prefix list.`),
			},
		},
	})
}

func testAccAwsPrefixListConfig_exceedLimit(count int) string {
	entries := ``
	for i := 0; i < count; i++ {
		entries += fmt.Sprintf(`
	entry {
		cidr_block  = "%[1]d.0.0.0/8"
		description = "Test_%[1]d"
	}
`, i+1)
	}

	return fmt.Sprintf(`
resource "aws_prefix_list" "test" {
	name           = "tf-test-acc"
	address_family = "IPv4"
	max_entries    = 2
%[1]s
}
`,
		entries)
}
