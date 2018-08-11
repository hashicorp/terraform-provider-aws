package aws

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDbEventCategories_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbEventCategoriesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDbEventCategoriesAttr("data.aws_db_event_categories.example"),
				),
			},
		},
	})
}

func testAccCheckAwsDbEventCategoriesAttr(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find DB Event Categories: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DB Event Categories resource ID not set.")
		}

		actual, err := testAccCheckAwsDbEventCategoriesBuild(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		expected := []string{
			"notification",
			"deletion",
			"failover",
			"maintenance",
			"availability",
			"read replica",
			"failure",
			"configuration change",
			"recovery",
			"low storage",
			"backup",
			"creation",
			"backtrack",
			"restoration",
		}

		sort.Strings(actual)
		sort.Strings(expected)
		if reflect.DeepEqual(expected, actual) != true {
			return fmt.Errorf("DB Event Categories not matched: expected %v, got %v", expected, actual)
		}

		return nil
	}
}

func testAccCheckAwsDbEventCategoriesBuild(attrs map[string]string) ([]string, error) {
	v, ok := attrs["event_categories.#"]
	if !ok {
		return nil, fmt.Errorf("DB Event Categories list is missing.")
	}

	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No DB Event Categories found.")
	}
	eventCategories := make([]string, qty)
	for n := range eventCategories {
		eventCategory, ok := attrs["event_categories."+strconv.Itoa(n)]
		if !ok {
			return nil, fmt.Errorf("DB Event Categories list is corrupted.")
		}
		eventCategories[n] = eventCategory
	}
	return eventCategories, nil

}

var testAccCheckAwsDbEventCategoriesConfig = `
data "aws_db_event_categories" "example" {}
`
