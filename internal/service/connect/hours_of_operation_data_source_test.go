// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccHoursOfOperationDataSource_hoursOfOperationID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_hours_of_operation.test"
	datasourceName := "data.aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationDataSourceConfig_id(rName, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hours_of_operation_id", resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "time_zone", resourceName, "time_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "config.#", resourceName, "config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccHoursOfOperationDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_hours_of_operation.test"
	datasourceName := "data.aws_connect_hours_of_operation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHoursOfOperationDataSourceConfig_name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hours_of_operation_id", resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "time_zone", resourceName, "time_zone"),
					resource.TestCheckResourceAttrPair(datasourceName, "config.#", resourceName, "config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccHoursOfOperationBaseDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Hours of Operation Description"
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
	`, rName, rName2)
}

func testAccHoursOfOperationDataSourceConfig_id(rName, rName2 string) string {
	return fmt.Sprintf(testAccHoursOfOperationBaseDataSourceConfig(rName, rName2) + `
data "aws_connect_hours_of_operation" "test" {
  instance_id           = aws_connect_instance.test.id
  hours_of_operation_id = aws_connect_hours_of_operation.test.hours_of_operation_id
}
`)
}

func testAccHoursOfOperationDataSourceConfig_name(rName, rName2 string) string {
	return fmt.Sprintf(testAccHoursOfOperationBaseDataSourceConfig(rName, rName2) + `
data "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_hours_of_operation.test.name
}
`)
}
