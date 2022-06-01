package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
)

func TestAccElastiCacheUserGroup_basic(t *testing.T) {
	var userGroup elasticache.UserGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
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

func TestAccElastiCacheUserGroup_update(t *testing.T) {
	var userGroup elasticache.UserGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
				),
			},
			{
				Config: testAccUserGroupMultipleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "user_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
				),
			},
			{
				Config: testAccUserGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "user_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
				),
			},
		},
	})
}

func TestAccElastiCacheUserGroup_tags(t *testing.T) {
	var userGroup elasticache.UserGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccUserGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccUserGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccElastiCacheUserGroup_disappears(t *testing.T) {
	var userGroup elasticache.UserGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupExists(resourceName, &userGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceUserGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserGroupDestroy(s *terraform.State) error {
	return testAccCheckUserGroupDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckUserGroupDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_user_group" {
			continue
		}

		_, err := tfelasticache.FindUserGroupByID(conn, rs.Primary.ID)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckUserGroupExists(n string, v *elasticache.UserGroup) resource.TestCheckFunc {
	return testAccCheckUserGroupExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckUserGroupExistsWithProvider(n string, v *elasticache.UserGroup, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache User Group ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn
		resp, err := tfelasticache.FindUserGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ElastiCache User Group (%s) not found: %w", rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccUserGroupBasicConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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
}
`, rName))
}

func testAccUserGroupMultipleConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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
  user_ids      = [aws_elasticache_user.test1.user_id, aws_elasticache_user.test2.user_id]
}
`, rName))
}

func testAccUserGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_user" "test1" {
  user_id       = "%[1]s-1"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
resource "aws_elasticache_user_group" "test" {
  user_group_id = %[1]q
  engine        = "REDIS"
  user_ids      = [aws_elasticache_user.test1.user_id]
  tags = {
    %[2]s = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccUserGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_user" "test1" {
  user_id       = "%[1]s-1"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
resource "aws_elasticache_user_group" "test" {
  user_group_id = %[1]q
  engine        = "REDIS"
  user_ids      = [aws_elasticache_user.test1.user_id]
  tags = {
    %[2]s = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
