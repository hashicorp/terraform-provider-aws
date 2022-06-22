package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccElastiCacheSubnetGroup_basic(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	resourceName := "aws_elasticache_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_basic(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
		},
	})
}

func TestAccElastiCacheSubnetGroup_update(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	resourceName := "aws_elasticache_subnet_group.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_updatePre(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					testAccCheckSubnetGroupAttrs(&csg, resourceName, 1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccSubnetGroupConfig_updatePost(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					testAccCheckSubnetGroupAttrs(&csg, resourceName, 2),
				),
			},
		},
	})
}

func TestAccElastiCacheSubnetGroup_tags(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	resourceName := "aws_elasticache_subnet_group.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetGroupConfig_tags1(rInt, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccSubnetGroupConfig_tags2(rInt, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccSubnetGroupConfig_tags1(rInt, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(resourceName, &csg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckSubnetGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_subnet_group" {
			continue
		}
		res, err := conn.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{
			CacheSubnetGroupName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			// Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault) {
				continue
			}
			return err
		}
		if len(res.CacheSubnetGroups) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckSubnetGroupExists(n string, csg *elasticache.CacheSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache subnet group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		resp, err := conn.DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{
			CacheSubnetGroupName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("CacheSubnetGroup error: %w", err)
		}

		for _, c := range resp.CacheSubnetGroups {
			if rs.Primary.ID == *c.CacheSubnetGroupName {
				*csg = *c
			}
		}

		if csg == nil {
			return fmt.Errorf("cache subnet group not found")
		}
		return nil
	}
}

func testAccCheckSubnetGroupAttrs(csg *elasticache.CacheSubnetGroup, n string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if len(csg.Subnets) != count {
			return fmt.Errorf("Bad cache subnet count, expected: %d, got: %d", count, len(csg.Subnets))
		}

		if rs.Primary.Attributes["description"] != *csg.CacheSubnetGroupDescription {
			return fmt.Errorf("Bad cache subnet description, expected: %s, got: %s", rs.Primary.Attributes["description"], *csg.CacheSubnetGroupDescription)
		}

		return nil
	}
}

func testAccSubnetGroupConfig_basic(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  name       = "tf-TEST-cache-subnet-%03d"
  subnet_ids = [aws_subnet.foo.id]
}
`, rInt))
}

func testAccSubnetGroupConfig_updatePre(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-subnet-group-update"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group-update-foo"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids  = [aws_subnet.foo.id]
}
`, rInt))
}

func testAccSubnetGroupConfig_updatePost(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-subnet-group-update"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group-update-foo"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group-update-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr-edited"
  subnet_ids = [
    aws_subnet.foo.id,
    aws_subnet.test.id,
  ]
}
`, rInt))
}

func testAccSubnetGroupConfig_tags1(rInt int, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  name       = "tf-TEST-cache-subnet-%03d"
  subnet_ids = [aws_subnet.foo.id]

  tags = {
    %q = %q
  }
}
`, rInt, tag1Key, tag1Value))
}

func testAccSubnetGroupConfig_tags2(rInt int, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-subnet-group"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-subnet-group"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  name       = "tf-TEST-cache-subnet-%03d"
  subnet_ids = [aws_subnet.foo.id]

  tags = {
    %q = %q
    %q = %q
  }
}
`, rInt, tag1Key, tag1Value, tag2Key, tag2Value))
}
