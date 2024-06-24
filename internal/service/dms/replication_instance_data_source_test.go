// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationInstanceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	replicationInstanceClass := "dms.c4.large"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_instance.test"
	dataSourceName := "data.aws_dms_replication_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckReplicationInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationInstanceDataSourceConfig_basic(rName, replicationInstanceClass),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_instance_arn", resourceName, "replication_instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_instance_id", resourceName, "replication_instance_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_instance_class", resourceName, "replication_instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_instance_private_ips.#", resourceName, "replication_instance_private_ips.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_instance_public_ips.#", resourceName, "replication_instance_public_ips.#"),
				),
			},
		},
	})
}

func testAccReplicationInstanceDataSourceConfig_basic(rName, replicationInstanceClass string) string {
	return acctest.ConfigCompose(testAccReplicationInstanceConfig_replicationInstanceClass(rName, replicationInstanceClass), `
data "aws_dms_replication_instance" "test" {
  replication_instance_id = aws_dms_replication_instance.test.replication_instance_id
}
`)
}
