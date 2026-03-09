// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationRouteCalculator_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_route_calculator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteCalculatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteCalculatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "calculator_arn", "geo", fmt.Sprintf("route-calculator/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "calculator_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, "data_source", "Here"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccLocationRouteCalculator_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_route_calculator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteCalculatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteCalculatorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflocation.ResourceRouteCalculator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationRouteCalculator_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_route_calculator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteCalculatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteCalculatorConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteCalculatorConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationRouteCalculator_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_location_route_calculator.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteCalculatorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteCalculatorConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRouteCalculatorConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRouteCalculatorConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckRouteCalculatorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_route_calculator" {
				continue
			}

			input := &location.DescribeRouteCalculatorInput{
				CalculatorName: aws.String(rs.Primary.ID),
			}

			_, err := conn.DescribeRouteCalculator(ctx, input)
			if err != nil {
				if errs.IsA[*awstypes.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			return fmt.Errorf("Expected Location Service Route Calculator to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRouteCalculatorExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Location Service Route Calculator is set")
		}

		conn := acctest.ProviderMeta(ctx, t).LocationClient(ctx)
		_, err := conn.DescribeRouteCalculator(ctx, &location.DescribeRouteCalculatorInput{
			CalculatorName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error describing Location Service Route Calculator: %s", err.Error())
		}

		return nil
	}
}

func testAccRouteCalculatorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_route_calculator" "test" {
  calculator_name = %[1]q
  data_source     = "Here"
}
`, rName)
}

func testAccRouteCalculatorConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_route_calculator" "test" {
  calculator_name = %[1]q
  data_source     = "Here"
  description     = %[2]q
}
`, rName, description)
}

func testAccRouteCalculatorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_route_calculator" "test" {
  calculator_name = %[1]q
  data_source     = "Here"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRouteCalculatorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_route_calculator" "test" {
  calculator_name = %[1]q
  data_source     = "Here"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
