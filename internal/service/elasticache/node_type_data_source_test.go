// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheNodeTypeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elasticache_node_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNodeTypeDataSourceConfig_basic("cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cache_node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(dataSourceName, "ec2_instance_type", "t3.medium"),
					resource.TestCheckResourceAttr(dataSourceName, "memory_size", "4096"),
					resource.TestCheckResourceAttr(dataSourceName, "default_vcpus", "2"),
					resource.TestCheckResourceAttrSet(dataSourceName, "burstable_performance_supported"),
					resource.TestCheckResourceAttrSet(dataSourceName, "current_generation"),
				),
			},
		},
	})
}

func TestAccElastiCacheNodeTypeDataSource_invalidType(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccNodeTypeDataSourceConfig_basic("cache.unknown.type"),
				ExpectError: regexp.MustCompile(`No EC2 instance type found|corresponding EC2 instance type`),
			},
		},
	})
}

func testAccNodeTypeDataSourceConfig_basic(cacheNodeType string) string {
	return fmt.Sprintf(`
data "aws_elasticache_node_type" "test" {
  cache_node_type = %[1]q
}
`, cacheNodeType)
}
