// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccApplicationSignalsServiceLevelObjectiveDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_applicationsignals_service_level_objective.test"
	resourceName := "aws_applicationsignals_service_level_objective.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationSignalsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLevelObjectiveDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedTime, resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_time", resourceName, "last_updated_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "evaluation_type", resourceName, "evaluation_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedTime),
				),
			},
		},
	})
}

func testAccServiceLevelObjectiveDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_applicationsignals_service_level_objective" "test" {
  name        = %[1]q
  description = "Test SLO"

  goal {
    attainment_goal   = 99.0
    warning_threshold = 99.5

    interval {
      rolling_interval {
        duration      = 7
        duration_unit = "DAY"
      }
    }
  }

  sli {
    comparison_operator = "GreaterThanOrEqualTo"
    metric_threshold    = 0.95

    sli_metric {
      metric_type = "LATENCY"

      metric_data_queries {
        id = "m1"

        metric_stat {
          metric {
            namespace   = "AWS/ApplicationSignals"
            metric_name = "Latency"
          }
          period = 60
          stat   = "Average"
        }
      }
    }
  }
}

data "aws_applicationsignals_service_level_objective" "test" {
  id = aws_applicationsignals_service_level_objective.test.name
}
`, rName)
}
