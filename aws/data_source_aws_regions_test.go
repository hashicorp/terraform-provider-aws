package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsRegions_Basic(t *testing.T) {
	resourceName := "data.aws_regions.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfig_empty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegions_OptIn(t *testing.T) {
	resourceName := "data.aws_regions.opt_in_status"

	statusOptedIn := "opted-in"
	statusNotOptedIn := "not-opted-in"
	statusOptInNotRequired := "opt-in-not-required"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// This resource has to be at the very top of the test scenario due to bug in Terrafom Plugin SDK
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegionsFiltered(statusOptedIn),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
				),
			},
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegionsFiltered(statusOptInNotRequired),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
				),
			},
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegionsFiltered(statusNotOptedIn),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegions_AllRegions(t *testing.T) {
	resourceAllRegions := "data.aws_regions.all_regions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceAllRegions),
					resource.TestCheckResourceAttr(resourceAllRegions, "all_regions", "true"),
					resource.TestCheckNoResourceAttr(resourceAllRegions, "opt_in_status"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRegionsCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsRegionsConfig_empty() string {
	return `
data "aws_regions" "empty" {}
`
}

func testAccDataSourceAwsRegionsConfig_allRegions() string {
	return `
data "aws_regions" "all_regions" {
	all_regions = "true"
}
`
}

func testAccDataSourceAwsRegionsConfig_allRegionsFiltered(filter string) string {
	return fmt.Sprintf(`
data "aws_regions" "opt_in_status" {
	filter {
       name   = "opt-in-status"
       values = ["%s"]
    }
}
`, filter)
}
