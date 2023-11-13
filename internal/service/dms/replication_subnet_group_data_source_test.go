// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDMSReplicationSubnetGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_subnet_group.test"
	dataSourceName := "data.aws_dms_replication_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSubnetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_subnet_group_id", resourceName, "replication_subnet_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_subnet_group_description", resourceName, "replication_subnet_group_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccReplicationSubnetGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccReplicationSubnetGroupConfig_basic(rName, "testing"), `
data "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`)
}
