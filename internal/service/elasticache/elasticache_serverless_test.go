// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccServerlessElastiCache_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckElasticCacheServerlessDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccServerlessElastiCache_basicRedis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckElasticCacheServerlessDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_basicRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "daily_snapshot_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccServerlessElastiCache_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckElasticCacheServerlessDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccServerlessElastiCache_fullRedis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckElasticCacheServerlessDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_fullRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccServerlessElastiCache_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	descriptionOld := "Memcached Serverless Cluter"
	descriptionNew := "Memcached Serverless Cluter updated"
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckElasticCacheServerlessDestroy(ctx),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_update(rName, descriptionOld),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServerlessElasticCacheConfig_update(rName, descriptionNew),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, "create_time"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "engine"),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, "serverless_cache_name"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
			},
		},
	})
}

func TestAccServerlessElastiCache_dissapears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless.test"
	var serverlessElasticCache elasticache.ServerlessCache

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckElasticCacheServerlessDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccServerlessElasticCacheConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlesssElasticCacheExists(ctx, resourceName, &serverlessElasticCache),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceElasticacheServerless(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServerlesssElasticCacheExists(ctx context.Context, resourceName string, v *elasticache.ServerlessCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Serverless ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn(ctx)
		grg, err := tfelasticache.FindElasicCacheServerlessByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("retrieving ElastiCache Serverlesss (%s): %w", rs.Primary.ID, err)
		}

		*v = *grg

		return nil
	}
}

func testAccCheckElasticCacheServerlessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_serverless" {
				continue
			}

			_, err := tfelasticache.FindElasicCacheServerlessByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("ElastiCache Serverless (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServerlessElasticCacheConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless" "test" {
  engine                = "memcached"
  serverless_cache_name = %[1]q
}

`, rName)
}

func testAccServerlessElasticCacheConfig_basicRedis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless" "test" {
  engine                = "redis"
  serverless_cache_name = %[1]q
}

`, rName)
}

func testAccServerlessElasticCacheConfig_full(rName string) string {
	//return fmt.Sprintf(`
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless" "test" {
  engine                = "memcached"
  serverless_cache_name = %[1]q
  cache_usage_limits {
	data_storage {
		maximum = 10
		unit    = "GB"
	}
	ecpu_per_second {
		maximum = 1000
	}
  }
  description           = "Test Full Memcached Attributes"
  kms_key_id            = aws_kms_key.test.arn
  major_engine_version  = "1.6"
  security_group_ids    = [aws_security_group.test.id]
  subnet_ids            = aws_subnet.test[*].id
  tags = {
    Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description = "tf-test-cmk-kms-key-id"
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
 
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
    }
}
`, rName))
}

func testAccServerlessElasticCacheConfig_fullRedis(rName string) string {
	//return fmt.Sprintf(`
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless" "test" {
  engine                = "redis"
  serverless_cache_name = %[1]q
  cache_usage_limits {
	data_storage {
		maximum = 10
		unit    = "GB"
	}
	ecpu_per_second {
		maximum = 1000
	}
  }
  daily_snapshot_time       = "09:00"
  description               = "Test Full Redis Attributes"
  kms_key_id                = aws_kms_key.test.arn
  major_engine_version      = "7"
  snapshot_retention_limit  = 1
  security_group_ids        = [aws_security_group.test.id]
  subnet_ids                = aws_subnet.test[*].id
  tags = {
    Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description = "tf-test-cmk-kms-key-id"
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
 
  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
    }
}
`, rName))
}

func testAccServerlessElasticCacheConfig_update(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless" "test" {
  engine                = "memcached"
  serverless_cache_name = %[1]q
  description           = %[2]q
}

`, rName, desc)
}
