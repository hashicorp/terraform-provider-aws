// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSourceClusterInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_cluster_instance.test"
	resourceName := "aws_rds_cluster_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceDataSourceConfig_complete(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_minor_version_upgrade", resourceName, "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ca_cert_identifier", resourceName, "ca_cert_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_identifier", resourceName, "cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "copy_tags_to_snapshot", resourceName, "copy_tags_to_snapshot"),
					resource.TestCheckResourceAttrPair(dataSourceName, "custom_iam_instance_profile", resourceName, "custom_iam_instance_profile"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_parameter_group_name", resourceName, "db_parameter_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_subnet_group_name", resourceName, "db_subnet_group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dbi_resource_id", resourceName, "dbi_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint", resourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_class", resourceName, "instance_class"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "monitoring_interval", resourceName, "monitoring_interval"),
					resource.TestCheckResourceAttrPair(dataSourceName, "monitoring_role_arn", resourceName, "monitoring_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_insights_enabled", resourceName, "performance_insights_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_insights_kms_key_id", resourceName, "performance_insights_kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_insights_retention_period", resourceName, "performance_insights_retention_period"),
					resource.TestCheckResourceAttrPair(dataSourceName, "preferred_backup_window", resourceName, "preferred_backup_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, "preferred_maintenance_window", resourceName, "preferred_maintenance_window"),
					resource.TestCheckResourceAttrPair(dataSourceName, "promotion_tier", resourceName, "promotion_tier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "publicly_accessible", resourceName, "publicly_accessible"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "writer", resourceName, "writer"),
				),
			},
		},
	})
}

func testAccClusterInstanceDataSourceConfig_complete(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterInstanceConfig_basic(rName),
		fmt.Sprintf(`
data "aws_rds_cluster_instance" "test" {
  identifier = aws_rds_cluster_instance.test.identifier
}
`))
}
