package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSElasticacheUser_basic(t *testing.T) {
	var user elasticache.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "username1"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
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

func TestAccAWSElasticacheUser_update(t *testing.T) {
	var user elasticache.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
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

func TestAccAWSElasticacheUser_tags(t *testing.T) {
	var user elasticache.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigTags(rName, "tagKey", "tagVal"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "username1"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal"),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigTags(rName, "tagKey", "tagVal2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "username1"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "tagVal2"),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "no_password_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "username1"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUser_disappears(t *testing.T) {
	var user elasticache.User
	rName := sdkacctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_elasticache_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(resourceName, &user),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSElasticacheUserDestroy(s *terraform.State) error {
	return testAccCheckAWSElasticacheUserDestroyWithProvider(s, acctest.Provider)
}

func testAccCheckAWSElasticacheUserDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_user" {
			continue
		}

		user, err := finder.ElastiCacheUserById(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeUserNotFoundFault) || tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if user != nil {
			return fmt.Errorf("Elasticache User (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSElasticacheUserExists(n string, v *elasticache.User) resource.TestCheckFunc {
	return testAccCheckAWSElasticacheUserExistsWithProvider(n, v, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckAWSElasticacheUserExistsWithProvider(n string, v *elasticache.User, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache User ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*conns.AWSClient).ElastiCacheConn
		resp, err := finder.ElastiCacheUserById(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ElastiCache User (%s) not found: %w", rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccAWSElasticacheUserConfigBasic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}
`, rName))
}

func testAccAWSElasticacheUserConfigUpdate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[1]q
  user_name     = "username1"
  access_string = "on ~* +@all"
  engine        = "REDIS"
  passwords     = ["password234567891", "password345678912"]
}
`, rName))
}

func testAccAWSElasticacheUserConfigTags(rName, tagKey, tagValue string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
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
`, rName, tagKey, tagValue))
}
