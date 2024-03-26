// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftDataSharesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_redshift_data_shares.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSharesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "data_shares.#"),
					resource.TestMatchTypeSetElemNestedAttrs(dataSourceName, "data_shares.*", map[string]*regexp.Regexp{
						"data_share_arn": regexache.MustCompile(`datashare:+.`),
						"producer_arn":   regexache.MustCompile(`namespace/+.`),
					}),
				),
			},
		},
	})
}

func testAccDataSharesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
  db_name        = "test"
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_redshiftdata_statement" "test_create" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = "CREATE DATASHARE tfacctest;"
}
`, rName),
		// Split this resource into a string literal so the terraform `format` function
		// interpolates properly.
		//
		// Grant statement must be run before the data share will be returned from the
		// DescribeDataShares API.
		`
resource "aws_redshiftdata_statement" "test_grant_usage" {
  depends_on     = [aws_redshiftdata_statement.test_create]
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
  database       = aws_redshiftserverless_namespace.test.db_name
  sql            = format("GRANT USAGE ON DATASHARE tfacctest TO ACCOUNT '%s';", data.aws_caller_identity.current.account_id)
}

data "aws_redshift_data_shares" "test" {
  depends_on = [aws_redshiftdata_statement.test_grant_usage]
}
`)
}
