// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testReplicationSetDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssmincidents_replication_set.test"
	resourceName := "aws_ssmincidents_replication_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMIncidentsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMIncidentsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationSetDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "created_by", dataSourceName, "created_by"),
					resource.TestCheckResourceAttrPair(resourceName, "deletion_protected", dataSourceName, "deletion_protected"),
					resource.TestCheckResourceAttrPair(resourceName, "last_modified_by", dataSourceName, "last_modified_by"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.a", dataSourceName, "tags.a"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.b", dataSourceName, "tags.b"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "region.0.name", dataSourceName, "region.0.name"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "region.0.kms_key_arn", dataSourceName, "region.0.kms_key_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "region.0.status", dataSourceName, "region.0.status"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "region.0.status_message", dataSourceName, "region.0.status_message"),

					acctest.MatchResourceAttrGlobalARN(dataSourceName, "arn", "ssm-incidents", regexp.MustCompile(`replication-set\/+.`)),
				),
			},
		},
	})
}

func testAccReplicationSetDataSourceConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }

  tags = {
    a = "tag1"
    b = ""
  }
}

data "aws_ssmincidents_replication_set" "test" {
  depends_on = [aws_ssmincidents_replication_set.test]
}


`, acctest.Region())
}
