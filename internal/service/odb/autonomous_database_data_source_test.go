// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBAutonomousDatabaseDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_odb_autonomous_database.test"
	dataSourceName := "data.aws_odb_autonomous_database.test"
	displayName := acctest.RandomWithPrefix(t, "tf-odb-adbs")
	dbName := "TFADB" + acctest.RandStringFromCharSet(t, 10, "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	config := acctest.ConfigCompose(
		testAccAutonomousDatabaseConfigBasic(displayName, dbName, 2, "AL32UTF8", "test"),
		`
data "aws_odb_autonomous_database" "test" {
  id = aws_odb_autonomous_database.test.id
}
`,
	)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabasePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutonomousDatabaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "data_storage_size_in_tbs", dataSourceName, "data_storage_size_in_tbs"),
					resource.TestCheckResourceAttrPair(resourceName, "db_name", dataSourceName, "db_name"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Environment", "test"),
				),
			},
		},
	})
}

func TestAccODBAutonomousDatabaseDataSource_notFound(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAutonomousDatabaseServicePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(`data "aws_odb_autonomous_database" "test" { id = %q }`, "adb-does-not-exist"),
				ExpectError: regexache.MustCompile("reading Autonomous Database Data Source"),
			},
		},
	})
}
