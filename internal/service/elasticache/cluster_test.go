package elasticache_test

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(elasticache.EndpointsID, testAccErrorCheckSkip)

}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"is not suppored in this region",
	)
}

func TestAccElastiCacheCluster_Engine_memcached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Engine_Memcached(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", "11211"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Engine_Redis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_redis_v5(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Engine_Redis_v5(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
					// Even though it is ignored, the API returns `true` in this case
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_None(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Engine_None(rName),
				// Verify "ExactlyOneOf" in the schema for "engine" and "replication_group_id"
				// throws a plan-time error when neither are configured.
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccElastiCacheCluster_PortRedis_default(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_RedisDefaultPort,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "to_port", "6379"),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "from_port", "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ParameterGroupName_default(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ParameterGroupName(rName, "memcached", "1.4.34", "default.memcached1.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memcached1.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_port(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	port := 11212
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Port(rName, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_SecurityGroup_ec2Classic(t *testing.T) {
	var ec elasticache.CacheCluster
	resourceName := "aws_elasticache_cluster.test"
	resourceSecurityGroupName := "aws_elasticache_security_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SecurityGroup_EC2Classic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(resourceSecurityGroupName),
					testAccCheckClusterEC2ClassicExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_snapshotsWithUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshots(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "05:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", "3"),
				),
			},
			{
				Config: testAccClusterConfig_snapshotsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "07:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", "7"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_decrease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
			{
				Config: testAccClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_increase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
			{
				Config: testAccClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_increaseWithPreferredAvailabilityZones(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "1"),
				),
			},
			{
				Config: testAccClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "3"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists("aws_elasticache_subnet_group.test", &csg),
					testAccCheckClusterExists("aws_elasticache_cluster.test", &ec),
					testAccCheckClusterAttributes(&ec),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_multiAZInVPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterMultiAZInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists("aws_elasticache_subnet_group.test", &csg),
					testAccCheckClusterExists("aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "availability_zone", "Multiple"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_AZMode_memcached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_AZMode_Memcached(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccClusterConfig_AZMode_Memcached(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccClusterConfig_AZMode_Memcached(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
			{
				Config:      testAccClusterConfig_AZMode_Memcached(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
		},
	})
}

func TestAccElastiCacheCluster_AZMode_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_AZMode_Redis(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccClusterConfig_AZMode_Redis(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccClusterConfig_AZMode_Redis(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_EngineVersion_memcached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, mid, post elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineVersion_Memcached(rName, "1.4.33"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.33"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.33"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Memcached(rName, "1.4.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &mid),
					testAccCheckClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.24"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.24"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Memcached(rName, "1.4.34"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &post),
					testAccCheckClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_EngineVersion_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4, v5, v6, v7, v8 elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "3.2.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.6"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "3.2.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v2),
					testAccCheckClusterRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.4"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.4"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "3.2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v3),
					testAccCheckClusterNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.10"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.10"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "6.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v4),
					testAccCheckClusterNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "6.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v5),
					testAccCheckClusterNotRecreated(&v4, &v5),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.2"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "5.0.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v6),
					testAccCheckClusterRecreated(&v5, &v6),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "6.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v7),
					testAccCheckClusterNotRecreated(&v6, &v7),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_EngineVersion_Redis(rName, "6.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &v8),
					testAccCheckClusterRecreated(&v7, &v8),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NodeTypeResize_memcached(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_NodeType_Memcached(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccClusterConfig_NodeType_Memcached(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &post),
					testAccCheckClusterRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NodeTypeResize_redis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_NodeType_Redis(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccClusterConfig_NodeType_Redis(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &post),
					testAccCheckClusterNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_redis(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_NumCacheNodes_Redis(rName, 2),
				ExpectError: regexp.MustCompile(`engine "redis" does not support num_cache_nodes > 1`),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_availabilityZone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_elasticache_cluster.test"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ReplicationGroupID_AvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckClusterExists(clusterResourceName, &cluster),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_singleReplica(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_elasticache_cluster.test.0"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ReplicationGroupID_Replica(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckClusterExists(clusterResourceName, &cluster),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName, "port", "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_multipleReplica(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster1, cluster2 elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName1 := "aws_elasticache_cluster.test.0"
	clusterResourceName2 := "aws_elasticache_cluster.test.1"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ReplicationGroupID_Replica(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(replicationGroupResourceName, &replicationGroup),

					testAccCheckClusterExists(clusterResourceName1, &cluster1),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster1, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName1, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName1, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName1, "port", "6379"),

					testAccCheckClusterExists(clusterResourceName2, &cluster2),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster2, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName2, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName2, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName2, "port", "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Memcached_finalSnapshot(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_Memcached_FinalSnapshot(rName),
				ExpectError: regexp.MustCompile(`engine "memcached" does not support final_snapshot_identifier`),
			},
		},
	})
}

func TestAccElastiCacheCluster_Redis_finalSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Redis_FinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_identifier", rName),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Redis_autoMinorVersionUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Redis_AutoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
			{
				Config: testAccClusterConfig_Redis_AutoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_Redis_LogDeliveryConfigurations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, true, elasticache.DestinationTypeCloudwatchLogs, elasticache.LogFormatText, true, elasticache.DestinationTypeCloudwatchLogs, elasticache.LogFormatText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, true, elasticache.DestinationTypeCloudwatchLogs, elasticache.LogFormatText, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "slow-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "engine-log"),
				),
			},
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson, true, elasticache.DestinationTypeCloudwatchLogs, elasticache.LogFormatText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, false, "", "", false, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.log_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_type"),
				),
			},
			{
				Config: testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName, true, elasticache.DestinationTypeKinesisFirehose, elasticache.LogFormatJson, false, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "json"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "slow-log"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccElastiCacheCluster_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"}, //not in the API
			},
			{
				Config: testAccClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_tagWithOtherModification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterVersionAndTagConfig(rName, "5.0.4", "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.4"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccClusterVersionAndTagConfig(rName, "5.0.6", "key1", "value1updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckClusterAttributes(v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.NotificationConfiguration == nil {
			return fmt.Errorf("Expected NotificationConfiguration for ElastiCache Cluster (%s)", *v.CacheClusterId)
		}

		if strings.ToLower(*v.NotificationConfiguration.TopicStatus) != "active" {
			return fmt.Errorf("Expected NotificationConfiguration status to be 'active', got (%s)", *v.NotificationConfiguration.TopicStatus)
		}

		return nil
	}
}

func testAccCheckClusterReplicationGroupIDAttribute(cluster *elasticache.CacheCluster, replicationGroup *elasticache.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cluster.ReplicationGroupId == nil {
			return errors.New("expected cluster ReplicationGroupId to be set")
		}

		if aws.StringValue(cluster.ReplicationGroupId) != aws.StringValue(replicationGroup.ReplicationGroupId) {
			return errors.New("expected cluster ReplicationGroupId to equal replication group ID")
		}

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CacheClusterCreateTime).Equal(aws.TimeValue(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime).Equal(aws.TimeValue(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_cluster" {
			continue
		}
		_, err := tfelasticache.FindCacheClusterByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Cache Cluster (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckClusterExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		resp, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("ElastiCache error: %v", err)
		}

		for _, c := range resp.CacheClusters {
			if *c.CacheClusterId == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

func testAccCheckClusterEC2ClassicExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).ElastiCacheConn

		input := &elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCacheClusters(input)

		if err != nil {
			return fmt.Errorf("error describing ElastiCache Cluster (%s): %w", rs.Primary.ID, err)
		}

		for _, c := range output.CacheClusters {
			if aws.StringValue(c.CacheClusterId) == rs.Primary.ID {
				*v = *c

				return nil
			}
		}

		return fmt.Errorf("ElastiCache Cluster (%s) not found", rs.Primary.ID)
	}
}

func testAccClusterConfig_Engine_Memcached(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%[1]s"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_Engine_Redis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%[1]s"
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_Engine_Redis_v5(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%[1]s"
  engine_version  = "5.0.6"
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_Engine_None(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%[1]s"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_ParameterGroupName(rName, engine, engineVersion, parameterGroupName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id           = %q
  engine               = %q
  engine_version       = %q
  node_type            = "cache.t3.small"
  num_cache_nodes      = 1
  parameter_group_name = %q
}
`, rName, engine, engineVersion, parameterGroupName)
}

func testAccClusterConfig_Port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = "%s"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = %d
}
`, rName, port)
}

func testAccClusterConfig_SecurityGroup_EC2Classic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticache_security_group" "test" {
  name                 = %[1]q
  description          = "tf-test-security-group-descr"
  security_group_names = [aws_security_group.test.name]
}

resource "aws_elasticache_cluster" "test" {
  cluster_id = %[1]q
  engine     = "memcached"

  # tflint-ignore: aws_elasticache_cluster_previous_type
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 11211
  security_group_names = [aws_elasticache_security_group.test.name]
}
`, rName))
}

func testAccClusterConfig_snapshots(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = %[1]q
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "05:00-09:00"
  snapshot_retention_limit = 3
}
`, rName)
}

func testAccClusterConfig_snapshotsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = %[1]q
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "07:00-09:00"
  snapshot_retention_limit = 7
  apply_immediately        = true
}
`, rName)
}

func testAccClusterConfig_NumCacheNodes(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%s"
  engine            = "memcached"
  node_type         = "cache.t3.small"
  num_cache_nodes   = %d
}
`, rName, numCacheNodes)
}

func testAccClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName string, numCacheNodes int) string {
	preferredAvailabilityZones := make([]string, numCacheNodes)
	for i := range preferredAvailabilityZones {
		preferredAvailabilityZones[i] = `data.aws_availability_zones.available.names[0]`
	}

	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_cluster" "test" {
  apply_immediately            = true
  cluster_id                   = "%s"
  engine                       = "memcached"
  node_type                    = "cache.t3.small"
  num_cache_nodes              = %d
  preferred_availability_zones = [%s]
}
`, rName, numCacheNodes, strings.Join(preferredAvailabilityZones, ","))
}

func testAccClusterInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-cluster-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-cluster-in-vpc"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids  = [aws_subnet.test.id]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  cluster_id             = %[1]q
  node_type              = "cache.t3.small"
  num_cache_nodes        = 1
  engine                 = "redis"
  engine_version         = "2.8.19"
  port                   = 6379
  subnet_group_name      = aws_elasticache_subnet_group.test.name
  security_group_ids     = [aws_security_group.test.id]
  parameter_group_name   = "default.redis2.8"
  notification_topic_arn = aws_sns_topic.test.arn
  availability_zone      = data.aws_availability_zones.available.names[0]
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccClusterMultiAZInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-cluster-multi-az-in-vpc"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id         = %[1]q
  engine             = "memcached"
  node_type          = "cache.t3.small"
  num_cache_nodes    = 2
  port               = 11211
  subnet_group_name  = aws_elasticache_subnet_group.test.name
  security_group_ids = [aws_security_group.test.id]
  az_mode            = "cross-az"
  preferred_availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1]
  ]
}
`, rName)
}

var testAccClusterConfig_RedisDefaultPort = `
resource "aws_security_group" "test" {
  name        = "tf-test-security-group"
  description = "tf-test-security-group-descr"
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = aws_elasticache_cluster.test.port
  protocol          = "tcp"
  security_group_id = aws_security_group.test.id
  to_port           = aws_elasticache_cluster.test.port
  type              = "ingress"
}

resource "aws_elasticache_cluster" "test" {
  cluster_id           = "foo-cluster"
  engine               = "redis"
  engine_version       = "5.0.4"
  node_type            = "cache.t2.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis5.0"
}
`

func testAccClusterConfig_AZMode_Memcached(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = "%[2]s"
  cluster_id        = "%[1]s"
  engine            = "memcached"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, azMode)
}

func testAccClusterConfig_AZMode_Redis(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = "%[2]s"
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, azMode)
}

func testAccClusterConfig_EngineVersion_Memcached(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "memcached"
  engine_version    = "%[2]s"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, engineVersion)
}

func testAccClusterConfig_EngineVersion_Redis(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  engine_version    = "%[2]s"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, engineVersion)
}

func testAccClusterConfig_NodeType_Memcached(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "memcached"
  node_type         = "%[2]s"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, nodeType)
}

func testAccClusterConfig_NodeType_Redis(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "%[2]s"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, nodeType)
}

func testAccClusterConfig_NumCacheNodes_Redis(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = %[2]d
  port              = 6379
}
`, rName, numCacheNodes)
}

func testAccClusterConfig_ReplicationGroupID_AvailabilityZone(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_description = "Terraform Acceptance Testing"
  replication_group_id          = "%[1]s"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = 1
  port                          = 6379

  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  cluster_id           = "%[1]s1"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName)
}

func testAccClusterConfig_ReplicationGroupID_Replica(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_description = "Terraform Acceptance Testing"
  replication_group_id          = "%[1]s"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = 1
  port                          = 6379

  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  count                = %[2]d
  cluster_id           = "%[1]s${count.index}"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName, count)
}

func testAccClusterConfig_Memcached_FinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccClusterConfig_Redis_FinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccClusterConfig_Redis_AutoMinorVersionUpgrade(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  engine_version  = "6.0"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  auto_minor_version_upgrade = %[2]t
}
`, rName, enable)
}

func testAccClusterConfig_Engine_Redis_LogDeliveryConfigurations(rName string, slowLogDeliveryEnabled bool, slowDeliveryDestination string, slowDeliveryFormat string, engineLogDeliveryEnabled bool, engineDeliveryDestination string, engineLogDeliveryFormat string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "p" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${aws_cloudwatch_log_group.lg.arn}:log-stream:*"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "rp" {
  policy_document = data.aws_iam_policy_document.p.json
  policy_name     = "%[1]s"
  depends_on = [
    aws_cloudwatch_log_group.lg
  ]
}

resource "aws_cloudwatch_log_group" "lg" {
  retention_in_days = 1
  name              = "%[1]s"
}

resource "aws_s3_bucket" "b" {
  force_destroy = true
}

resource "aws_iam_role" "r" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      },
    ]
  })
  inline_policy {
    name = "my_inline_s3_policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "s3:AbortMultipartUpload",
            "s3:GetBucketLocation",
            "s3:GetObject",
            "s3:ListBucket",
            "s3:ListBucketMultipartUploads",
            "s3:PutObject",
            "s3:PutObjectAcl",
          ]
          Effect   = "Allow"
          Resource = ["${aws_s3_bucket.b.arn}", "${aws_s3_bucket.b.arn}/*"]
        },
      ]
    })
  }
}

resource "aws_kinesis_firehose_delivery_stream" "ds" {
  name        = "%[1]s"
  destination = "s3"
  s3_configuration {
    role_arn   = aws_iam_role.r.arn
    bucket_arn = aws_s3_bucket.b.arn
  }
  lifecycle {
    ignore_changes = [
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id        = "%[1]s"
  engine            = "redis"
  node_type         = "cache.t3.micro"
  num_cache_nodes   = 1
  port              = 6379
  apply_immediately = true
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[2]t") ? [""] : []
    content {
      destination      = ("%[3]s" == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : (("%[3]s" == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = "%[3]s"
      log_format       = "%[4]s"
      log_type         = "slow-log"
    }
  }
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[5]t") ? [""] : []
    content {
      destination      = ("%[6]s" == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : (("%[6]s" == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = "%[6]s"
      log_format       = "%[7]s"
      log_type         = "engine-log"
    }
  }
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName, slowLogDeliveryEnabled, slowDeliveryDestination, slowDeliveryFormat, engineLogDeliveryEnabled, engineDeliveryDestination, engineLogDeliveryFormat)

}

func testAccClusterConfigTags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccClusterConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccClusterVersionAndTagConfig(rName, version, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id        = %[1]q
  node_type         = "cache.t3.small"
  num_cache_nodes   = 1
  engine            = "redis"
  engine_version    = %[2]q
  apply_immediately = true

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, version, tagKey1, tagValue1),
	)
}
