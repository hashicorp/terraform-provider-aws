// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheUserGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var userGroup awstypes.UserGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_user_group.test"
	resourceName := "aws_elasticache_user_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserGroupExists(ctx, t, resourceName, &userGroup),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "elasticache", regexache.MustCompile(`usergroup:.+`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, "user_group_id"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey1, resourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsKey2, resourceName, acctest.CtTagsKey2),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_group_id", resourceName, "user_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_ids", resourceName, "user_ids"),
				),
			},
		},
	})
}

func testAccUserGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test1" {
  user_id       = "%[1]s-1"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user" "test2" {
  user_id       = "%[1]s-2"
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  user_group_id = %[1]q
  engine        = "REDIS"
  user_ids      = [aws_elasticache_user.test1.user_id]

  tags = {
    "key1" = "value1"
    "key2" = "value2"
  }
}

data "aws_elasticache_user_group" "test" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
}
`, rName)
}
