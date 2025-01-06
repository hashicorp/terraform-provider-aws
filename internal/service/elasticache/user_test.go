// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "username1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"no_password_required",
					"passwords",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_passwordAuthMode(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigWithPasswordAuthMode_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "username1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.passwords.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "authentication_mode.0.passwords.*", "aaaaaaaaaaaaaaaa"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.type", names.AttrPassword),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"authentication_mode.0.passwords.#",
					"authentication_mode.0.passwords.0",
					"no_password_required",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_iamAuthMode(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigWithIAMAuthMode_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.type", "iam"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"no_password_required",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_update(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
				),
			},
			{
				Config: testAccUserConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"no_password_required",
					"passwords",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_updateEngine(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigWithEngine(rName, "VALKEY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "valkey"),
				),
			},
			{
				Config: testAccUserConfigWithEngine(rName, "REDIS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"no_password_required",
					"passwords",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_updatePasswordAuthMode(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfigWithPasswordAuthMode_twoPasswords(rName, "aaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"authentication_mode.0.passwords",
					"no_password_required",
				},
			},
			{
				Config: testAccUserConfigWithPasswordAuthMode_onePassword(rName, "aaaaaaaaaaaaaaaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"authentication_mode.0.passwords",
					"no_password_required",
				},
			},
			{
				Config: testAccUserConfigWithPasswordAuthMode_twoPasswords(rName, "cccccccccccccccc", "dddddddddddddddd"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "authentication_mode.0.password_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"authentication_mode.0.passwords",
					"no_password_required",
				},
			},
		},
	})
}

func TestAccElastiCacheUser_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags(rName, "tagKey", "tagVal"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "username1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal"),
				),
			},
			{
				Config: testAccUserConfig_tags(rName, "tagKey", "tagVal2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "username1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal2"),
				),
			},
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrUserName, "username1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccElastiCacheUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/34002.
func TestAccElastiCacheUser_oobModify(t *testing.T) {
	ctx := acctest.Context(t)
	var user awstypes.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			// Start to update the user out-of-band.
			{
				Config: testAccUserConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserUpdateOOB(ctx, &user),
				),
				ExpectNonEmptyPlan: true,
			},
			// Update tags.
			{
				Config: testAccUserConfig_tags(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_user" {
				continue
			}

			_, err := tfelasticache.FindUserByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache User (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserExists(ctx context.Context, n string, v *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindUserByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserUpdateOOB(ctx context.Context, v *awstypes.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		_, err := conn.ModifyUser(ctx, &elasticache.ModifyUserInput{
			AccessString: aws.String("on ~* +@all"),
			UserId:       v.UserId,
		})

		return err
	}
}

func testAccUserConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
`, rName)
}

func testAccUserConfigWithPasswordAuthMode_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }
}
`, rName)
}

func testAccUserConfigWithIAMAuthMode_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = %[1]q
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"

  authentication_mode {
    type = "iam"
  }
}
`, rName)
}

func testAccUserConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~* +@all"
  engine        = "REDIS"
  passwords     = ["password234567891", "password345678912"]
}
`, rName)
}

func testAccUserConfigWithEngine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~* +@all"
  engine        = %[2]q
  passwords     = ["password123456789"]
}
`, rName, engine)
}

func testAccUserConfigWithPasswordAuthMode_twoPasswords(rName string, password1 string, password2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"

  authentication_mode {
    type      = "password"
    passwords = [%[2]q, %[3]q]
  }
}
`, rName, password1, password2)
}

func testAccUserConfigWithPasswordAuthMode_onePassword(rName string, password string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"

  authentication_mode {
    type      = "password"
    passwords = [%[2]q]
  }
}
`, rName, password)
}

func testAccUserConfig_tags(rName, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey, tagValue)
}
