// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchSchedulingPolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_scheduling_policy.test"
	dataSourceName := "data.aws_batch_scheduling_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSchedulingPolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "fair_share_policy.#", resourceName, "fair_share_policy.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fair_share_policy.0.compute_reservation", resourceName, "fair_share_policy.0.compute_reservation"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fair_share_policy.0.share_decay_seconds", resourceName, "fair_share_policy.0.share_decay_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fair_share_policy.0.share_distribution.#", resourceName, "fair_share_policy.0.share_distribution.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccSchedulingPolicyDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test" {
  name = %[1]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }
}
`, rName)
}

func testAccSchedulingPolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(testAccSchedulingPolicyDataSourceConfig(rName) + `
data "aws_batch_scheduling_policy" "test" {
  arn = aws_batch_scheduling_policy.test.arn
}
`)
}
