// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/location"
	awstypes "github.com/aws/aws-sdk-go-v2/service/location/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflocation "github.com/hashicorp/terraform-provider-aws/internal/service/location"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationPlaceIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, "data_source", "Here"),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", "SingleUse"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "index_arn", "geo", fmt.Sprintf("place-index/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "index_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflocation.ResourcePlaceIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLocationPlaceIndex_dataSourceConfigurationIntendedUse(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexConfig_configurationIntendedUse(rName, "SingleUse"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", "SingleUse"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlaceIndexConfig_configurationIntendedUse(rName, "Storage"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "data_source_configuration.0.intended_use", "Storage"),
				),
			},
		},
	})
}

func TestAccLocationPlaceIndex_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlaceIndexConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccLocationPlaceIndex_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPlaceIndexConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPlaceIndexConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPlaceIndexExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckPlaceIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_place_index" {
				continue
			}

			input := &location.DescribePlaceIndexInput{
				IndexName: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribePlaceIndex(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
}

func testAccCheckPlaceIndexExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationClient(ctx)

		input := &location.DescribePlaceIndexInput{
			IndexName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribePlaceIndex(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Location Service Place Index (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPlaceIndexConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}
`, rName)
}

func testAccPlaceIndexConfig_configurationIntendedUse(rName, intendedUse string) string {
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

func testAccPlaceIndexConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  description = %[2]q
  index_name  = %[1]q
}
`, rName, description)
}

func testAccPlaceIndexConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPlaceIndexConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
