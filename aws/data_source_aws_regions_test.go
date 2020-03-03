package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsRegionsBasic(t *testing.T) {
	resourceName := "data.aws_regions.empty"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
					resource.TestCheckNoResourceAttr(resourceName, "opt_in_status"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegionsOptIn(t *testing.T) {
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
				Config:      testAccDataSourceAwsRegionsConfigOptIn("invalid-opt-in-status"),
				ExpectError: regexp.MustCompile(`expected opt_in_status to be one of .*, got invalid-opt-in-status`),
			},
			{
				Config: testAccDataSourceAwsRegionsConfigOptIn(statusOptedIn),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", statusOptedIn),
				),
			},
			{
				Config: testAccDataSourceAwsRegionsConfigOptIn(statusOptInNotRequired),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", statusOptInNotRequired),
				),
			},
			{
				Config: testAccDataSourceAwsRegionsConfigOptIn(statusNotOptedIn),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRegionsCheck(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "all_regions"),
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", statusNotOptedIn),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRegionsAllRegions(t *testing.T) {
	resourceAllRegions := "data.aws_regions.all_regions"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRegionsConfigAllRegions(),
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

func testAccDataSourceAwsRegionsConfigEmpty() string {
	return `
data "aws_regions" "empty" {}
`
}

func testAccDataSourceAwsRegionsConfigAllRegions() string {
	return `
data "aws_regions" "all_regions" {
	all_regions = "true"
}
`
}

func testAccDataSourceAwsRegionsConfigOptIn(optInStatus string) string {
	return fmt.Sprintf(`
data "aws_regions" "opt_in_status" {
	opt_in_status = "%s"
}
`, optInStatus)
}
