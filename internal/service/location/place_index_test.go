package location_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
)

func TestAccLocationPlaceIndex_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaceIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigPlaceIndex_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_time"),
					resource.TestCheckResourceAttr(resourceName, "data_source", "Here"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", locationservice.IntendedUseSingleUse),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "index_arn", "geo", fmt.Sprintf("place-index/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
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

func TestAccLocationPlaceIndex_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigPlaceIndex_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflocation.ResourcePlaceIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationPlaceIndex_dataSourceConfigurationIntendedUse(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckMapDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigPlaceIndex_dataSourceConfigurationIntendedUse(rName, locationservice.IntendedUseSingleUse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", locationservice.IntendedUseSingleUse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigPlaceIndex_dataSourceConfigurationIntendedUse(rName, locationservice.IntendedUseStorage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", locationservice.IntendedUseStorage),
				),
			},
		},
	})
}

func TestAccLocationPlaceIndex_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPlaceIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigPlaceIndex_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigPlaceIndex_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccCheckPlaceIndexDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_location_place_index" {
			continue
		}

		input := &locationservice.DescribePlaceIndexInput{
			IndexName: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribePlaceIndex(input)

		if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Location Service Place Index (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Location Service Place Index (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckPlaceIndexExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn

		input := &locationservice.DescribePlaceIndexInput{
			IndexName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribePlaceIndex(input)

		if err != nil {
			return fmt.Errorf("error getting Location Service Place Index (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccConfigPlaceIndex_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}
`, rName)
}

func testAccConfigPlaceIndex_dataSourceConfigurationIntendedUse(rName, intendedUse string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"

  data_source_configuration {
    intended_use = %[2]q
  }

  index_name = %[1]q
}
`, rName, intendedUse)
}

func testAccConfigPlaceIndex_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  description = %[2]q
  index_name  = %[1]q
}
`, rName, description)
}
