// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheUserGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheUserGroupAssociation_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationConfig_preUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				Config: testAccUserGroupAssociationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", fmt.Sprintf("%s-3", rName)),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func TestAccElastiCacheUserGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceUserGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheUserGroupAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName1 := "aws_elasticache_user_group_association.test1"
	resourceName2 := "aws_elasticache_user_group_association.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationConfig_preMultiple(rName),
			},
			{
				Config: testAccUserGroupAssociationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(ctx, resourceName1),
					testAccCheckUserGroupAssociationExists(ctx, resourceName2),
				),
			},
		},
	})
}

func testAccCheckUserGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_user_group_association" {
				continue
			}

			err := tfelasticache.FindUserGroupAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["user_group_id"], rs.Primary.Attributes["user_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache User Group Association (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserGroupAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		err := tfelasticache.FindUserGroupAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["user_group_id"], rs.Primary.Attributes["user_id"])

		return err
	}
}

func testAccUserGroupAssociationConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
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

  lifecycle {
    ignore_changes = [user_ids]
  }
}
`, rName))
}

func testAccUserGroupAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccUserGroupAssociationConfig_base(rName), `
resource "aws_elasticache_user_group_association" "test" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test2.user_id
}
`)
}

func testAccUserGroupAssociationConfig_preUpdate(rName string) string {
	return acctest.ConfigCompose(testAccUserGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_elasticache_user" "test3" {
  user_id       = "%[1]s-3"
  user_name     = "username2"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group_association" "test" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test2.user_id
}
`, rName))
}

func testAccUserGroupAssociationConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccUserGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_elasticache_user" "test3" {
  user_id       = "%[1]s-3"
  user_name     = "username2"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group_association" "test" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test3.user_id
}
`, rName))
}

func testAccUserGroupAssociationConfig_preMultiple(rName string) string {
	return acctest.ConfigCompose(testAccUserGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_elasticache_user" "test3" {
  user_id       = "%[1]s-3"
  user_name     = "username2"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
`, rName))
}

func testAccUserGroupAssociationConfig_multiple(rName string) string {
	return acctest.ConfigCompose(testAccUserGroupAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_elasticache_user" "test3" {
  user_id       = "%[1]s-3"
  user_name     = "username2"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group_association" "test1" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test2.user_id
}

resource "aws_elasticache_user_group_association" "test2" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test3.user_id
}
`, rName))
}
