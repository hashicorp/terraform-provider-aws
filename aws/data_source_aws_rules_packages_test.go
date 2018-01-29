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

func TestAccAWSRulesPackages_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsRulesPackagesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRulesPackagesMeta("data.aws_rules_packages.aws_rules_packages"),
				),
			},
		},
	})
}

func testAccCheckAwsRulesPackagesMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Tules Packages resource: %s", n)
		}

		fmt.Printf("%s", rs)

		if rs.Primary.ID == "" {
			return fmt.Errorf("Rules Packages resource ID not set.")
		}

		actual, err := testAccCheckAwsRulesPackagesBuildAvailable(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		expected := actual
		sort.Strings(expected)
		if reflect.DeepEqual(expected, actual) != true {
			return fmt.Errorf("Rules Packages not sorted - expected %v, got %v", expected, actual)
		}
		return nil
	}
}

func testAccCheckAwsRulesPackagesBuildAvailable(attrs map[string]string) ([]string, error) {
	v, ok := attrs["arns.#"]
	if !ok {
		return nil, fmt.Errorf("Available Rules Packages list is missing.")
	}
	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No Rules Packages found in region, this is probably a bug.")
	}
	packages := make([]string, qty)
	for n := range packages {
		zone, ok := attrs["arns."+strconv.Itoa(n)]
		if !ok {
			return nil, fmt.Errorf("Rules Packages list corrupt, this is definitely a bug.")
		}
		packages[n] = zone
	}
	return packages, nil
}

const testAccCheckAwsRulesPackagesConfig = `
data "aws_rules_packages" "rules_packages" { }
`
