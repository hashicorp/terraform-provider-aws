// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccHoursOfOperation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HoursOfOperation
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	originalDescription := "original description"
	updatedDescription := "updated description"

	resourceName := "aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHoursOfOperationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationConfig_basic(rName, rName2, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/operating-hours/{hours_of_operation_id}"),
					resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           "1",
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         "1",
						"start_time.0.hours":   "8",
						"start_time.0.minutes": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttrSet(resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "time_zone", "EST"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHoursOfOperationConfig_basic(rName, rName2, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "connect", "instance/{instance_id}/operating-hours/{hours_of_operation_id}"),
					resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           "1",
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         "1",
						"start_time.0.hours":   "8",
						"start_time.0.minutes": "0",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttrSet(resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "time_zone", "EST"),
				),
			},
		},
	})
}

func testAccHoursOfOperation_updateConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HoursOfOperation
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	description := "example description"

	resourceName := "aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHoursOfOperationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationConfig_basic(rName, rName2, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           "1",
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         "1",
						"start_time.0.hours":   "8",
						"start_time.0.minutes": "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHoursOfOperationConfig_multipleConfig(rName, rName2, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "config.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           "1",
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         "1",
						"start_time.0.hours":   "8",
						"start_time.0.minutes": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "TUESDAY",
						"end_time.#":           "1",
						"end_time.0.hours":     "21",
						"end_time.0.minutes":   "0",
						"start_time.#":         "1",
						"start_time.0.hours":   "9",
						"start_time.0.minutes": "0",
					}),
				),
			},
		},
	})
}

func testAccHoursOfOperation_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HoursOfOperation
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHoursOfOperationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationConfig_basic(rName, rName2, names.AttrTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHoursOfOperationConfig_tags(rName, rName2, names.AttrTags),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccHoursOfOperationConfig_tagsUpdated(rName, rName2, names.AttrTags),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccHoursOfOperation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HoursOfOperation
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHoursOfOperationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationConfig_basic(rName, rName2, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceHoursOfOperation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHoursOfOperationExists(ctx context.Context, n string, v *awstypes.HoursOfOperation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		output, err := tfconnect.FindHoursOfOperationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["hours_of_operation_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckHoursOfOperationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_hours_of_operation" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

			_, err := tfconnect.FindHoursOfOperationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["hours_of_operation_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Hours Of Operation %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccHoursOfOperationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccHoursOfOperationConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccHoursOfOperationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }

  tags = {
    "Name" = "Test Hours of Operation"
  }
}
`, rName2, label))
}

func testAccHoursOfOperationConfig_multipleConfig(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccHoursOfOperationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }

  config {
    day = "TUESDAY"

    end_time {
      hours   = 21
      minutes = 0
    }

    start_time {
      hours   = 9
      minutes = 0
    }
  }

  tags = {
    "Name" = "Test Hours of Operation"
  }
}
`, rName2, label))
}

func testAccHoursOfOperationConfig_tags(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccHoursOfOperationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }

  tags = {
    "Name" = "Test Hours of Operation"
    "Key2" = "Value2a"
  }
}
`, rName2, label))
}

func testAccHoursOfOperationConfig_tagsUpdated(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccHoursOfOperationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q
  time_zone   = "EST"

  config {
    day = "MONDAY"

    end_time {
      hours   = 23
      minutes = 8
    }

    start_time {
      hours   = 8
      minutes = 0
    }
  }

  tags = {
    "Name" = "Test Hours of Operation"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName2, label))
}
