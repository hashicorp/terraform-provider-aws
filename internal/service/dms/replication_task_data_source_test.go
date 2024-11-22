// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationTaskDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	dataSourceName := "data.aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_task_id", resourceName, "replication_task_id"),
				),
			},
		},
	})
}

func testAccReplicationTaskDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
}

data "aws_dms_replication_task" "test" {
  replication_task_id = aws_dms_replication_task.test.replication_task_id
}
`, rName))
}
