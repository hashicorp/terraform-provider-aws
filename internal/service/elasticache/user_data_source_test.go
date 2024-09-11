// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheUserDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_elasticache_user.test-basic"
	dataSourceName := "data.aws_elasticache_user.test-basic"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(dataSourceName, "access_string", resourceName, "access_string"),
				),
			},
		},
	})
}

// Basic Resource
func testAccUserDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-basic" {
  user_id              = %[1]q
  user_name            = %[1]q
  access_string        = "on ~* +@all"
  engine               = "REDIS"
  no_password_required = true
}

data "aws_elasticache_user" "test-basic" {
  user_id = aws_elasticache_user.test-basic.user_id
}
`, rName)
}
