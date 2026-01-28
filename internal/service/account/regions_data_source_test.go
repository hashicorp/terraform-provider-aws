// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package account_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	enabledByDefault = "ENABLED_BY_DEFAULT"
	disabled         = "DISABLED"
)

func TestAccAccountRegionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_account_regions.test"

	enabledByDefaultRegions := []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ca-central-1", "eu-central-1", "eu-north-1", "eu-west-1", "eu-west-2", "eu-west-3", "sa-east-1"}

	var testChecks []resource.TestCheckFunc
	for _, region := range enabledByDefaultRegions {
		testChecks = append(testChecks, resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
			"region_name":       region,
			"region_opt_status": enabledByDefault,
		}))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: `data "aws_account_regions" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					testChecks...,
				),
			},
			{
				Config: testAccAccountRegionsDataSourceConfig_disabled(),
				Check: resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "regions.*", map[string]string{
					"region_opt_status": disabled,
				}),
			},
		},
	})
}

func testAccAccountRegionsDataSourceConfig_disabled() string {
	return fmt.Sprintf(`
data "aws_account_regions" "test" {
  region_opt_status_contains = ["%s"]
}
`, disabled)
}
