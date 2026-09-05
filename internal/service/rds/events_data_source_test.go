// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSEventsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_events.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Event content and count are timing-dependent; only
					// assert the attribute exists and is well-formed.
					resource.TestCheckResourceAttrSet(dataSourceName, "events.#"),
				),
			},
		},
	})
}

func testAccEventsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), `
data "aws_rds_events" "test" {
  source_identifier = aws_db_instance.test.identifier
  source_type       = "db-instance"

  depends_on = [aws_db_instance.test]
}
`)
}
