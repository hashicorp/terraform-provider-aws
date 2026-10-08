// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build acctest
// +build acctest

package neptune_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneClusterDataSource_basic(t *testing.T) {
	identifier := os.Getenv("TF_ACC_NEPTUNE_CLUSTER_ID")
	if identifier == "" {
		t.Skip("TF_ACC_NEPTUNE_CLUSTER_ID must be set to run this test")
	}

	ctx := acctest.Context(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.Neptune) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Neptune),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNeptuneClusterConfig(identifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "id"),
					resource.TestCheckResourceAttr("data.aws_neptune_cluster.test", "identifier", identifier),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "engine"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "engine_version"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "reader_endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "port"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "status"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "deletion_protection"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "iam_database_authentication_enabled"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "multi_az"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "storage_encrypted"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "availability_zones.#"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "members.#"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "vpc_security_group_ids.#"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "kms_key_id"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "preferred_backup_window"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "preferred_maintenance_window"),
					resource.TestCheckResourceAttrSet("data.aws_neptune_cluster.test", "resource_id"),
					resource.TestMatchResourceAttr("data.aws_neptune_cluster.test", "storage_type", regexp.MustCompile("^.*$")),
				),
			},
		},
	})
}

func testAccDataSourceNeptuneClusterConfig(identifier string) string {
	return `
data "aws_neptune_cluster" "test" {
  identifier = "` + identifier + `"
}
`
}
