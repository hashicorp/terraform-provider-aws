// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rdsdata_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSDataQueryAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSDataServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryActionConfig_basic(rName),
			},
		},
	})
}

func testAccQueryActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccQueryDataSourceConfig_base(rName), `
resource "aws_rdsdata_query" "test" {
  depends_on   = [aws_rds_cluster_instance.test]
  resource_arn = aws_rds_cluster.test.arn
  secret_arn   = aws_secretsmanager_secret_version.test.arn
  sql          = "SELECT 1"
}
`)
}
