package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strconv"
	"testing"
	"time"
)

func TestAccAWSSpotPriceHistoryDataSource_product_description_filter_windows(t *testing.T) {
	resourceName := "data.aws_spot_price_history.product_windows"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotPriceHistoryDataSourceProductFilterWindowsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotPriceHistoryFilterResponses("product_description", resourceName),
				),
			},
		},
	})
}

/* not working even though documented as valid parameter
func TestAccAWSSpotPriceHistoryDataSource_product_description_filter_windows_vpc(t *testing.T) {
	resourceName := "data.aws_spot_price_history.product_windows_vpc"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotPriceHistoryDataSourceProductFilterWindowsVPCConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotPriceHistoryFilterResponses("product_description", resourceName),
				),
			},
		},
	})
}
*/

func TestAccAWSSpotPriceHistoryDataSource_instance_type_filter(t *testing.T) {
	resourceName := "data.aws_spot_price_history.instance_type_m1_large"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotPriceHistoryDataSourceInstanceTypeFilterConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotPriceHistoryFilterResponses("instance_type", resourceName),
					testAccCheckElementOrder(resourceName),
				),
			},
		},
	})
}

func testAccCheckElementOrder(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find spot price history resource: %s", n)
		}

		err := testAccCheckLatestPriceIsNewerThanTheFirstPreviousPrice(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckLatestPriceIsNewerThanTheFirstPreviousPrice(attrs map[string]string) error {
	var previousTime time.Time
	var latestTime time.Time
	var err error
	l, ok := attrs["latest.timestamp"]
	if !ok {
		return fmt.Errorf("Spot price history latest is missing, this is probably a bug.")
	}
	latestTime, err = time.Parse(time.RFC3339, l)
	if err != nil {
		return err
	}

	c, ok := attrs["previous.#"]
	if !ok {
		return fmt.Errorf("Spot price history list is missing, this is probably a bug.")
	}
	qty, err := strconv.Atoi(c)
	if err != nil {
		return err
	}
	if qty < 1 {
		return fmt.Errorf("No spot price history found, this is probably a bug.")
	}

	i := make([]string, qty)
	for n := range i {
		if t, ok := attrs["previous."+strconv.Itoa(n)+".timestamp"]; ok {
			previousTime, err = time.Parse(time.RFC3339, t)
			if err != nil {
				return err
			}
		}

		if latestTime.Sub(previousTime) < 0 {
			return fmt.Errorf("Spot price history order is not correct. This is definitely bug.")
		}
	}

	return nil
}

func testAccCheckSpotPriceHistoryFilterResponses(filter_name, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find spot price history resource: %s", n)
		}

		err := testAccCheckResponseContainsTheSameValueForFilters(filter_name, rs.Primary.Attributes)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckResponseContainsTheSameValueForFilters(filter_name string, attrs map[string]string) error {
	c, ok := attrs["latest."+filter_name]
	if !ok {
		return fmt.Errorf("Latest spot price is missing, this is probably a bug.")
	}
	v, ok := attrs["previous.#"]
	if !ok {
		return fmt.Errorf("Spot price history list is missing, this is probably a bug.")
	}
	qty, err := strconv.Atoi(v)
	if err != nil {
		return err
	}
	if qty < 1 {
		return fmt.Errorf("No spot price history found, this is probably a bug.")
	}
	i := make([]string, qty)
	for n := range i {
		t, ok := attrs["previous."+strconv.Itoa(n)+"."+filter_name]
		if ok && c != "" && t != c {
			return fmt.Errorf("Spot price history list contains different values when filtering for %s. This is definitely bug.", filter_name)
		}
		c = t
	}

	return nil
}

func testAccAWSSpotPriceHistoryDataSourceProductFilterWindowsConfig() string {
	now := time.Now()
	return fmt.Sprintf(`
data "aws_spot_price_history" "product_windows" {
	start_time = "%s"
	end_time = "%s"
	filter {
		name = "product-description"
		values = ["Windows"]
	}
}`, now.Add(time.Duration(-1)*time.Hour).Format(time.RFC3339), now.Add(time.Duration(-30)*time.Minute).Format(time.RFC3339))
}

/*
func testAccAWSSpotPriceHistoryDataSourceProductFilterWindowsVPCConfig() string {
	now := time.Now()
	return fmt.Sprintf(`
data "aws_spot_price_history" "product_windows_vpc" {
	start_time = "%s"
	end_time = "%s"
	filter {
		name = "product-description"
		values = ["Windows (Amazon VPC)"]
	}
}`, now.Add(time.Duration(-1)*time.Hour).Format(time.RFC3339), now.Add(time.Duration(-30)*time.Minute).Format(time.RFC3339))
}
*/

// specific instance type
func testAccAWSSpotPriceHistoryDataSourceInstanceTypeFilterConfig() string {
	now := time.Now()
	return fmt.Sprintf(`
data "aws_spot_price_history" "instance_type_m1_large" {
	start_time = "%s"
	end_time = "%s"
	filter {
    	name = "instance-type"
    	values = ["m1.large"]
  	}
}`, now.Add(time.Duration(-1)*time.Hour).Format(time.RFC3339), now.Add(time.Duration(-30)*time.Minute).Format(time.RFC3339))
}
