package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccCheckAwsEc2ManagedPrefixListDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list" {
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

func testAccCheckAwsEc2ManagedPrefixListVersion(
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

func TestAccAwsEc2ManagedPrefixList_basic(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	pl, entries := ec2.ManagedPrefixList{}, []*ec2.PrefixListEntry(nil)
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_basic_create(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, &entries),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "5"),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr_block":  "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr_block":  "2.0.0.0/8",
						"description": "Test2",
					}),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccCheckAwsEc2ManagedPrefixListVersion(&pl, 1),
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
				Config:       testAccAwsEc2ManagedPrefixListConfig_basic_update(rName2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, &entries),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr_block":  "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr_block":  "3.0.0.0/8",
						"description": "Test3",
					}),
					testAccCheckAwsEc2ManagedPrefixListVersion(&pl, 2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
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

func testAccAwsEc2ManagedPrefixListConfig_basic_create(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
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
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_basic_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
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
`, rName)
}

func testAccAwsEc2ManagedPrefixListExists(
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

func TestAccAwsEc2ManagedPrefixList_disappears(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	pl := ec2.ManagedPrefixList{}
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_disappears(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsEc2ManagedPrefixListConfig_disappears(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 2

  entry {
    cidr_block = "1.0.0.0/8"
  }
}
`, rName)
}

func TestAccAwsEc2ManagedPrefixList_name(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	pl := ec2.ManagedPrefixList{}
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_name_create(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					testAccCheckAwsEc2ManagedPrefixListVersion(&pl, 1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_name_update(rName2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					testAccCheckAwsEc2ManagedPrefixListVersion(&pl, 1),
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

func testAccAwsEc2ManagedPrefixListConfig_name_create(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_name_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}
`, rName)
}

func TestAccAwsEc2ManagedPrefixList_tags(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	pl := ec2.ManagedPrefixList{}
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_tags_none(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_tags_addSome(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
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
				Config:       testAccAwsEc2ManagedPrefixListConfig_tags_dropOrModifySome(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
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
				Config:       testAccAwsEc2ManagedPrefixListConfig_tags_empty(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_tags_none(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &pl, nil),
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

func testAccAwsEc2ManagedPrefixListConfig_tags_none(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_tags_addSome(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_tags_dropOrModifySome(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    Key2 = "Value2-1"
    Key3 = "Value3"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_tags_empty(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5
  tags           = {}
}
`, rName)
}

func TestAccAwsEc2ManagedPrefixList_exceedLimit(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	prefixList := ec2.ManagedPrefixList{}
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_exceedLimit(rName, 2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName, &prefixList, nil),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
				),
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_exceedLimit(rName, 3),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile(`You've reached the maximum number of entries for the prefix list.`),
			},
		},
	})
}

func testAccAwsEc2ManagedPrefixListConfig_exceedLimit(rName string, count int) string {
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
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[2]q
  address_family = "IPv4"
  max_entries    = 2
%[1]s
}
`, entries, rName)
}
