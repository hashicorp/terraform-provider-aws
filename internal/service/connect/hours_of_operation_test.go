// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccHoursOfOperation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeHoursOfOperationOutput
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           acctest.Ct1,
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         acctest.Ct1,
						"start_time.0.hours":   "8",
						"start_time.0.minutes": acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
					resource.TestCheckResourceAttrSet(resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           acctest.Ct1,
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         acctest.Ct1,
						"start_time.0.hours":   "8",
						"start_time.0.minutes": acctest.Ct0,
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
					resource.TestCheckResourceAttrSet(resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "time_zone", "EST"),
				),
			},
		},
	})
}

func testAccHoursOfOperation_updateConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeHoursOfOperationOutput
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
					resource.TestCheckResourceAttr(resourceName, "config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           acctest.Ct1,
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         acctest.Ct1,
						"start_time.0.hours":   "8",
						"start_time.0.minutes": acctest.Ct0,
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
					resource.TestCheckResourceAttr(resourceName, "config.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "MONDAY",
						"end_time.#":           acctest.Ct1,
						"end_time.0.hours":     "23",
						"end_time.0.minutes":   "8",
						"start_time.#":         acctest.Ct1,
						"start_time.0.hours":   "8",
						"start_time.0.minutes": acctest.Ct0,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "config.*", map[string]string{
						"day":                  "TUESDAY",
						"end_time.#":           acctest.Ct1,
						"end_time.0.hours":     "21",
						"end_time.0.minutes":   acctest.Ct0,
						"start_time.#":         acctest.Ct1,
						"start_time.0.hours":   "9",
						"start_time.0.minutes": acctest.Ct0,
					}),
				),
			},
		},
	})
}

func testAccHoursOfOperation_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeHoursOfOperationOutput
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test Hours of Operation"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccHoursOfOperationConfig_tagsUpdated(rName, rName2, names.AttrTags),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHoursOfOperationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
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
	var v connect.DescribeHoursOfOperationOutput
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

func testAccCheckHoursOfOperationExists(ctx context.Context, resourceName string, function *connect.DescribeHoursOfOperationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Hours of Operation not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Hours of Operation ID not set")
		}
		instanceID, hoursOfOperationID, err := tfconnect.HoursOfOperationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeHoursOfOperationInput{
			HoursOfOperationId: aws.String(hoursOfOperationID),
			InstanceId:         aws.String(instanceID),
		}

		getFunction, err := conn.DescribeHoursOfOperationWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckHoursOfOperationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_hours_of_operation" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, hoursOfOperationID, err := tfconnect.HoursOfOperationParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeHoursOfOperationInput{
				HoursOfOperationId: aws.String(hoursOfOperationID),
				InstanceId:         aws.String(instanceID),
			}

			_, err = conn.DescribeHoursOfOperationWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}
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
