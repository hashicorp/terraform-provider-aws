// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheServiceUpdatesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_elasticache_service_updates.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdatesDataSourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_updates"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_updates"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.ServiceUpdateStatusAvailable),
						}),
						// There are currently no cancelled service updates
						// knownvalue.ObjectPartial(map[string]knownvalue.Check{
						// 	names.AttrStatus: tfknownvalue.StringExact(awstypes.ServiceUpdateStatusCancelled),
						// }),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.ServiceUpdateStatusExpired),
						}),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccElastiCacheServiceUpdatesDataSource_bySingleStatus(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_elasticache_service_updates.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdatesDataSourceConfig_bySingleStatus(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_updates"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrStatus: tfknownvalue.StringExact(awstypes.ServiceUpdateStatusAvailable),
						}),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.ServiceUpdateStatusAvailable),
					})),
				},
			},
		},
	})
}

func testAccServiceUpdatesDataSourceConfig_basic() string {
	return `
data "aws_elasticache_service_updates" "test" {}
`
}

func testAccServiceUpdatesDataSourceConfig_bySingleStatus() string {
	return `
data "aws_elasticache_service_updates" "test" {
  status = ["available"]
}
`
}
