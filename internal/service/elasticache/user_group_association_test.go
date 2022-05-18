package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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

func TestAccElastiCacheUserGroupAssociation_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(resourceName),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationPreUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				Config: testAccUserGroupAssociationUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_id", fmt.Sprintf("%s-3", rName)),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func TestAccElastiCacheUserGroupAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_user_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupAssociationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserGroupAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceUserGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserGroupAssociationDestroy(s *terraform.State) error {
	return testAccCheckUserGroupAssociationDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckUserGroupAssociationDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_user_group_association" {
			continue
		}

		groupID, userID, err := tfelasticache.UserGroupAssociationParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("parsing User Group Association ID (%s): %w", rs.Primary.ID, err)
		}

		output, err := tfelasticache.FindUserGroupByID(conn, groupID)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserGroupNotFoundFault) {
				return nil
			}
		}

		gotUserID := ""
		for _, v := range output.UserIds {
			if aws.StringValue(v) == userID {
				gotUserID = aws.StringValue(v)
				break
			}
		}

		if gotUserID != "" {
			return fmt.Errorf("ElastiCache User Group Association (%s) found, should be deleted", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccCheckUserGroupAssociationExists(n string) resource.TestCheckFunc {
	return testAccCheckUserGroupAssociationExistsWithProvider(n, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckUserGroupAssociationExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache User Group Association ID is set")
		}

		groupID, userID, err := tfelasticache.UserGroupAssociationParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("parsing User Group Association ID (%s): %w", rs.Primary.ID, err)
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn
		output, err := tfelasticache.FindUserGroupByID(conn, groupID)
		if err != nil {
			return fmt.Errorf("ElastiCache User Group Association (%s) not found: %w", rs.Primary.ID, err)
		}

		gotUserID := ""
		for _, v := range output.UserIds {
			if aws.StringValue(v) == userID {
				gotUserID = aws.StringValue(v)
				break
			}
		}

		if gotUserID == "" {
			return fmt.Errorf("ElastiCache User Group Association (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUserGroupAssociationBaseConfig(rName string) string {
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

func testAccUserGroupAssociationBasicConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccUserGroupAssociationBaseConfig(rName),
		`
resource "aws_elasticache_user_group_association" "test" {
  user_group_id = aws_elasticache_user_group.test.user_group_id
  user_id       = aws_elasticache_user.test2.user_id
}
`)
}

func testAccUserGroupAssociationPreUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccUserGroupAssociationBaseConfig(rName),
		fmt.Sprintf(`
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

func testAccUserGroupAssociationUpdateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccUserGroupAssociationBaseConfig(rName),
		fmt.Sprintf(`
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
