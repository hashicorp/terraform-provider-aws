package elasticache_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElastiCacheReplicationGroup_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "elasticache", fmt.Sprintf("replicationgroup:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "0"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
					resource.TestCheckResourceAttr(resourceName, "data_tiering_enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"}, //not in the API
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_uppercase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_Uppercase(strings.ToUpper(rName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "replication_group_id", rName),
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

func TestAccElastiCacheReplicationGroup_EngineVersion_update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4, v5 elasticache.ReplicationGroup
	var c1, c2, c3, c4, c5 map[string]*elasticache.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_EngineVersion(rName, "3.2.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &v1),
					testAccCheckReplicationGroupMemberClusters(resourceName, &c1),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.6"),
				),
			},
			{
				Config: testAccReplicationGroupConfig_EngineVersion(rName, "3.2.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &v2),
					testAccCheckReplicationGroupMemberClusters(resourceName, &c2),
					testAccCheckReplicationGroupRecreated(&c1, &c2),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.4"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.4"),
				),
			},
			{
				Config: testAccReplicationGroupConfig_EngineVersion(rName, "3.2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &v3),
					testAccCheckReplicationGroupMemberClusters(resourceName, &c3),
					testAccCheckReplicationGroupNotRecreated(&c2, &c3),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.10"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "3.2.10"),
				),
			},
			{
				Config: testAccReplicationGroupConfig_EngineVersion(rName, "6.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &v4),
					testAccCheckReplicationGroupMemberClusters(resourceName, &c4),
					testAccCheckReplicationGroupNotRecreated(&c3, &c4),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccReplicationGroupConfig_EngineVersion(rName, "5.0.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &v5),
					testAccCheckReplicationGroupMemberClusters(resourceName, &c5),
					testAccCheckReplicationGroupRecreated(&c4, &c5),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_group_description", "test description"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupUpdatedDescriptionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "replication_group_description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateMaintenanceWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "tue:06:30-tue:07:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupUpdatedMaintenanceWindowConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "wed:03:00-wed:06:00"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateUserGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userGroup := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupUserGroup(rName, userGroup, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					testAccCheckReplicationGroupUserGroup(resourceName, fmt.Sprintf("%s-%d", userGroup, 0)),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_group_ids.*", fmt.Sprintf("%s-%d", userGroup, 0)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupUserGroup(rName, userGroup, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					testAccCheckReplicationGroupUserGroup(resourceName, fmt.Sprintf("%s-%d", userGroup, 1)),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_group_ids.*", fmt.Sprintf("%s-%d", userGroup, 1)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateNodeSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupUpdatedNodeSizeConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

//This is a test to prove that we panic we get in https://github.com/hashicorp/terraform/issues/9097
func TestAccElastiCacheReplicationGroup_updateParameterGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	parameterGroupResourceName1 := "aws_elasticache_parameter_group.test.0"
	parameterGroupResourceName2 := "aws_elasticache_parameter_group.test.1"
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupParameterGroupNameConfig(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", parameterGroupResourceName1, "name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupParameterGroupNameConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", parameterGroupResourceName2, "name"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateAuthToken(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroup_EnableAuthTokenTransitEncryptionConfig(1, "one", sdkacctest.RandString(16)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "transit_encryption_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "auth_token", "availability_zones"},
			},
			{
				Config: testAccReplicationGroup_EnableAuthTokenTransitEncryptionConfig(1, "one", sdkacctest.RandString(16)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_multiAzNotInVPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_MultiAZNotInVPC_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupConfig_MultiAZNotInVPC_AvailabilityZones(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zones.0", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zones.1", "data.aws_availability_zones.available", "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_multiAzInVPC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupMultiAZInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", aws.StringValue(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", aws.StringValue(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ValidationMultiAz_noAutomaticFailover(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_MultiAZ_NoAutomaticFailover(rName),
				ExpectError: regexp.MustCompile("automatic_failover_enabled must be true if multi_az_enabled is true"),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_redisClusterInVPC2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupRedisClusterInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", rName, acctest.PartitionDNSSuffix()))),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", rName, acctest.PartitionDNSSuffix()))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
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

func TestAccElastiCacheReplicationGroup_ClusterMode_nonClusteredParameterGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig_NonClusteredParameterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", rName, acctest.PartitionDNSSuffix()))),
					resource.TestCheckNoResourceAttr(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
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

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroups_scaleUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 4, []kvp{
						{"key", "value"},
					}),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 3, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 6, []kvp{
						{"key", "value"},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroups_scaleDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 3, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_updateReplicasPerNodeGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "3"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "8"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "8"),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroupsAndReplicasPerNodeGroup_scaleUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroupsAndReplicasPerNodeGroup_scaleDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
				),
			},
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_singleNode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupNativeRedisClusterConfig_SingleNode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "0"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "1"),
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

func TestAccElastiCacheReplicationGroup_clusteringAndCacheNodesCausesError(t *testing.T) {
	rInt := sdkacctest.RandInt()
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupNativeRedisClusterErrorConfig(rInt, rName),
				ExpectError: regexp.MustCompile(`"cluster_mode.0.num_node_groups": conflicts with number_cache_clusters`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_enableSnapshotting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupEnableSnapshottingConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_enableAuthTokenTransitEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroup_EnableAuthTokenTransitEncryptionConfig(sdkacctest.RandInt(), sdkacctest.RandString(10), sdkacctest.RandString(16)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "transit_encryption_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "auth_token", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_enableAtRestEncryption(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroup_EnableAtRestEncryptionConfig(sdkacctest.RandInt(), sdkacctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "at_rest_encryption_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_useCMKKMSKeyID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroup_UseCMKKMSKeyID(sdkacctest.RandInt(), sdkacctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttrSet("aws_elasticache_replication_group.bar", "kms_key_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{"key", "value"},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 4, []kvp{
						{"key", "value"},
					}),
				),
			},
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{"key", "value"},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersFailover_autoFailoverDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := false
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					if err := resourceReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersFailover_autoFailoverEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceReplicationGroupDisableAutomaticFailover(conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceReplicationGroupEnableAutomaticFailover(conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_multiAZEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceReplicationGroupDisableAutomaticFailover(conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceReplicationGroupEnableAutomaticFailover(conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappears_noChange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappears_addMemberCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappearsRemoveMemberCluster_atTargetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappearsRemoveMemberCluster_scaleDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{"key1", "value1"},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"}, //not in the API
			},
			{
				Config: testAccReplicationGroupTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{"key1", "value1updated"},
						{"key2", "value2"},
					}),
				),
			},
			{
				Config: testAccReplicationGroupTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{"key2", "value2"},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_finalSnapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupFinalSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_identifier", rName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Validation_noNodeType(t *testing.T) {
	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_Validation_NoNodeType(rName),
				ExpectError: regexp.MustCompile(`"node_type" is required unless "global_replication_group_id" is set.`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Validation_globalReplicationGroupIdAndNodeType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_Validation_GlobalReplicationGroupIdAndNodeType(rName),
				ExpectError: regexp.MustCompile(`"global_replication_group_id": conflicts with node_type`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "global_replication_group_id", "aws_elasticache_global_replication_group.test", "global_replication_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "node_type", primaryGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", primaryGroupResourceName, "engine_version"),
					resource.TestMatchResourceAttr(resourceName, "parameter_group_name", regexp.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "number_cache_clusters", "2"),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_full(t *testing.T) {
	var providers []*schema.Provider
	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	initialNumCacheClusters := 2
	updatedNumCacheClusters := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_Full(rName, initialNumCacheClusters),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "global_replication_group_id", "aws_elasticache_global_replication_group.test", "global_replication_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "node_type", primaryGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version", primaryGroupResourceName, "engine_version"),
					resource.TestMatchResourceAttr(resourceName, "parameter_group_name", regexp.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),

					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", strconv.Itoa(initialNumCacheClusters)),
					resource.TestCheckResourceAttrPair(resourceName, "multi_az_enabled", primaryGroupResourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryGroupResourceName, "automatic_failover_enabled"),

					resource.TestCheckResourceAttr(resourceName, "port", "16379"),

					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_encryption_enabled", primaryGroupResourceName, "transit_encryption_enabled"),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_Full(rName, updatedNumCacheClusters),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", strconv.Itoa(updatedNumCacheClusters)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupIDClusterMode_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	var rg1, rg2 elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_ClusterMode(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg1),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "parameter_group_name", regexp.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),

					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.0.replicas_per_node_group", "2"),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_ClusterMode(rName, 1, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg2),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "3"),

					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "cluster_mode.0.replicas_per_node_group", "1"),
				),
			},
		},
	})
}

// Test for out-of-band deletion
// Naming to allow grouping all TestAccAWSElasticacheReplicationGroup_GlobalReplicationGroupId_* tests
func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	// nosemgrep: acceptance-test-naming-parent-disappears
	var providers []*schema.Provider
	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupIDClusterModeValidation_numNodeGroupsOnSecondary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var providers []*schema.Provider
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 2),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_GlobalReplicationGroupId_ClusterMode_NumNodeGroupsOnSecondary(rName),
				ExpectError: regexp.MustCompile(`"global_replication_group_id": conflicts with cluster_mode.0.num_node_groups`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_dataTiering(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfigDataTiering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.x"),
					resource.TestCheckResourceAttr(resourceName, "data_tiering_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func testAccCheckReplicationGroupExists(n string, v *elasticache.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No replication group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		rg, err := tfelasticache.FindReplicationGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ElastiCache error: %w", err)
		}

		*v = *rg

		return nil
	}
}

func testAccCheckReplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_replication_group" {
			continue
		}
		_, err := tfelasticache.FindReplicationGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Replication Group (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckReplicationGroupMemberClusters(n string, v *map[string]*elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var rg elasticache.ReplicationGroup

		err := testAccCheckReplicationGroupExists(n, &rg)(s)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

		clusters := make(map[string]*elasticache.CacheCluster, len(rg.MemberClusters))
		for _, clusterID := range rg.MemberClusters {
			c, err := tfelasticache.FindCacheClusterWithNodeInfoByID(conn, aws.StringValue(clusterID))
			if err != nil {
				return fmt.Errorf("could not read ElastiCache replication group (%s) member cluster (%s): %w", n, aws.StringValue(clusterID), err)
			}

			clusters[aws.StringValue(c.CacheClusterId)] = c
		}

		*v = clusters

		return nil
	}
}

func testAccCheckReplicationGroupUserGroup(resourceName, userGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		rg, err := tfelasticache.FindReplicationGroupByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}
		if len(rg.UserGroupIds) < 1 {
			return fmt.Errorf("ElastiCache Replication Group (%s) was not assigned any usergroups", resourceName)
		}

		if *rg.UserGroupIds[0] != userGroupID {
			return fmt.Errorf("ElastiCache Replication Group (%s) was not assigned usergroup (%s), usergroup was (%s) instead", resourceName, userGroupID, *rg.UserGroupIds[0])
		}
		return nil

	}
}

func testAccCheckReplicationGroupRecreated(i, j *map[string]*elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for key, iv := range *i {
			jv, ok := (*j)[key]
			if !ok {
				continue
			}

			if aws.TimeValue(iv.CacheClusterCreateTime).Equal(aws.TimeValue(jv.CacheClusterCreateTime)) {
				return fmt.Errorf("ElastiCache replication group not recreated: member cluster (%s) not recreated", key)
			}
		}

		return nil
	}
}

func testAccCheckReplicationGroupNotRecreated(i, j *map[string]*elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for key, iv := range *i {
			jv, ok := (*j)[key]
			if !ok {
				continue
			}

			if !aws.TimeValue(iv.CacheClusterCreateTime).Equal(aws.TimeValue(jv.CacheClusterCreateTime)) {
				return fmt.Errorf("ElastiCache replication group recreated: member cluster (%s) recreated", key)
			}
		}

		return nil
	}
}

type kvp struct {
	key   string
	value string
}

func testAccReplicationGroupCheckMemberClusterTags(resourceName, dataSourceNamePrefix string, memberCount int, kvs []kvp) resource.TestCheckFunc {
	checks := testAccCheckResourceTags(resourceName, kvs)
	checks = append(checks, resource.TestCheckResourceAttr(resourceName, "member_clusters.#", strconv.Itoa(memberCount))) // sanity check

	for i := 0; i < memberCount; i++ {
		dataSourceName := fmt.Sprintf("%s.%d", dataSourceNamePrefix, i)
		checks = append(checks, testAccCheckResourceTags(dataSourceName, kvs)...)
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testAccCheckResourceTags(resourceName string, kvs []kvp) []resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 1, 1+len(kvs))
	checks[0] = resource.TestCheckResourceAttr(resourceName, "tags.%", strconv.Itoa(len(kvs)))
	for _, kv := range kvs {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", kv.key), kv.value))
	}
	return checks
}

func testAccReplicationGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
}
`, rName)
}

func testAccReplicationGroupConfig_Uppercase(rName string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = 1
  port                          = 6379
  replication_group_description = "test description"
  replication_group_id          = %[1]q
  subnet_group_name             = aws_elasticache_subnet_group.test.name
}
`, rName)
}

func testAccReplicationGroupConfig_EngineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"

  node_type             = "cache.t3.small"
  number_cache_clusters = 2

  engine_version     = %[2]q
  apply_immediately  = true
  maintenance_window = "tue:06:30-tue:07:30"
  snapshot_window    = "01:00-02:00"
}
`, rName, engineVersion)
}

func testAccReplicationGroupEnableSnapshottingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
  snapshot_retention_limit      = 2
}
`, rName)
}

func testAccReplicationGroupParameterGroupNameConfig(rName string, parameterGroupNameIndex int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  count = 2

  # We do not have a data source for "latest" ElastiCache family
  # so unfortunately we must hardcode this for now
  family = "redis6.x"

  name = "%[1]s-${count.index}"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }
}

resource "aws_elasticache_replication_group" "test" {
  apply_immediately             = true
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  parameter_group_name          = aws_elasticache_parameter_group.test.*.name[%[2]d]
  replication_group_description = "test description"
  replication_group_id          = %[1]q
}
`, rName, parameterGroupNameIndex)
}

func testAccReplicationGroupUpdatedDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = true
}
`, rName)
}

func testAccReplicationGroupUpdatedMaintenanceWindowConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = true
  maintenance_window            = "wed:03:00-wed:06:00"
  snapshot_window               = "01:00-02:00"
}
`, rName)
}

func testAccReplicationGroupUpdatedNodeSizeConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.medium"
  port                          = 6379
  apply_immediately             = true
}
`, rName)
}

func testAccReplicationGroupUserGroup(rName, userGroup string, flag int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  count = 2

  user_id       = "%[2]s-${count.index}"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  count = 2

  user_group_id = "%[2]s-${count.index}"
  engine        = "REDIS"
  user_ids      = [aws_elasticache_user.test[count.index].user_id]
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
  transit_encryption_enabled    = true
  user_group_ids                = [aws_elasticache_user_group.test[%[3]d].id]
}
`, rName, userGroup, flag)

}

func testAccReplicationGroupInVPCConfig(rName string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-in-vpc"
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

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 1
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  auto_minor_version_upgrade    = false
}
`, rName)
}

func testAccReplicationGroupConfig_MultiAZNotInVPC_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 2
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true
}
`, rName)
}

func testAccReplicationGroupConfig_MultiAZNotInVPC_AvailabilityZones(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 2
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
}
`, rName)
}

func testAccReplicationGroupMultiAZInVPCConfig(rName string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-multi-az-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test.id,
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

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = true
  multi_az_enabled              = true
  snapshot_window               = "02:00-03:00"
  snapshot_retention_limit      = 7
}
`, rName)
}

func testAccReplicationGroupConfig_MultiAZ_NoAutomaticFailover(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 1
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = false
  multi_az_enabled              = true
}
`, rName)
}

func testAccReplicationGroupRedisClusterInVPCConfig(rName string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-redis-cluster-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test.id,
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

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = "2"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = false
  snapshot_window               = "02:00-03:00"
  snapshot_retention_limit      = 7
  engine_version                = "3.2.4"
  maintenance_window            = "thu:03:00-thu:04:00"
}
`, rName)
}

func testAccReplicationGroupNativeRedisClusterErrorConfig(rInt int, rName string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster-err"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03[1]d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03[1]d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  automatic_failover_enabled    = true

  cluster_mode {
    replicas_per_node_group = 1
    num_node_groups         = 2
  }

  number_cache_clusters = 3
}
`, rInt, rName)
}

func testAccReplicationGroupNativeRedisClusterConfig(rName string, numNodeGroups, replicasPerNodeGroup int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccReplicationGroupClusterData(numNodeGroups*(1+replicasPerNodeGroup)),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test.id,
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

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  automatic_failover_enabled    = true

  cluster_mode {
    num_node_groups         = %[2]d
    replicas_per_node_group = %[3]d
  }

  tags = {
    key = "value"
  }
}
`, rName, numNodeGroups, replicasPerNodeGroup),
	)
}

func testAccReplicationGroupNativeRedisClusterConfig_NonClusteredParameterGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.medium"
  automatic_failover_enabled    = false

  parameter_group_name = "default.redis6.x"
  cluster_mode {
    num_node_groups         = 1
    replicas_per_node_group = 1
  }
}
`, rName))
}

func testAccReplicationGroupNativeRedisClusterConfig_SingleNode(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.medium"
  automatic_failover_enabled    = true

  parameter_group_name = "default.redis6.x.cluster.on"
  cluster_mode {
    num_node_groups         = 1
    replicas_per_node_group = 0
  }
}
`, rName))
}

func testAccReplicationGroup_UseCMKKMSKeyID(rInt int, rString string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.foo.id,
  ]
}

resource "aws_security_group" "bar" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.foo.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_kms_key" "bar" {
  description = "tf-test-cmk-kms-key-id"
}

resource "aws_elasticache_replication_group" "bar" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.bar.name
  security_group_ids            = [aws_security_group.bar.id]
  parameter_group_name          = "default.redis3.2"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "3.2.6"
  at_rest_encryption_enabled    = true
  kms_key_id                    = aws_kms_key.bar.arn
}
`, rInt, rInt, rString)
}

func testAccReplicationGroup_EnableAtRestEncryptionConfig(rInt int, rString string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  parameter_group_name          = "default.redis3.2"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "3.2.6"
  at_rest_encryption_enabled    = true
}
`, rInt, rInt, rString)
}

func testAccReplicationGroup_EnableAuthTokenTransitEncryptionConfig(rInt int, rString10 string, rString16 string) string {
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
    Name = "terraform-testacc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  parameter_group_name          = "default.redis5.0"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "5.0.6"
  transit_encryption_enabled    = true
  auth_token                    = "%s"
}
`, rInt, rInt, rString10, rString16)
}

func testAccReplicationGroupConfig_NumberCacheClusters(rName string, numberCacheClusters int) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(numberCacheClusters),
		fmt.Sprintf(`
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
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = %[2]d
  replication_group_id          = %[1]q
  replication_group_description = "Terraform Acceptance Testing - number_cache_clusters"
  subnet_group_name             = aws_elasticache_subnet_group.test.name

  tags = {
    key = "value"
  }
}
`, rName, numberCacheClusters),
	)
}

func testAccReplicationGroupConfig_FailoverMultiAZ(rName string, numberCacheClusters int, autoFailover, multiAZ bool) string {
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
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  # InvalidParameterCombination: Automatic failover is not supported for T1 and T2 cache node types.
  automatic_failover_enabled    = %[3]t
  multi_az_enabled              = %[4]t
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = %[2]d
  replication_group_id          = "%[1]s"
  replication_group_description = "Terraform Acceptance Testing - number_cache_clusters"
  subnet_group_name             = aws_elasticache_subnet_group.test.name
}
`, rName, numberCacheClusters, autoFailover, multiAZ)
}

func testAccReplicationGroupTags1Config(rName, tagKey1, tagValue1 string) string {
	const clusterCount = 2
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(clusterCount),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = %[2]d
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, clusterCount, tagKey1, tagValue1),
	)
}

func testAccReplicationGroupTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	const clusterCount = 2
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(clusterCount),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = %[2]d
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, clusterCount, tagKey1, tagValue1, tagKey2, tagValue2),
	)
}

func testAccReplicationGroupClusterData(count int) string {
	return fmt.Sprintf(`
data "aws_elasticache_cluster" "test" {
  count = %[1]d

  cluster_id = tolist(aws_elasticache_replication_group.test.member_clusters)[count.index]
}
`, count)
}

func testAccReplicationGroupFinalSnapshotConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccReplicationGroupConfig_Validation_NoNodeType(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 1
}
`, rName)
}

func testAccReplicationGroupConfig_Validation_GlobalReplicationGroupIdAndNodeType(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccElasticacheVpcBaseWithProvider(rName, "test", acctest.ProviderName, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  provider = aws

  replication_group_id          = "%[1]s-s"
  replication_group_description = "secondary"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.test.name

  node_type = "cache.m5.large"

  number_cache_clusters = 1
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}
`, rName))
}

func testAccReplicationGroupConfig_GlobalReplicationGroupId_Basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccElasticacheVpcBaseWithProvider(rName, "test", acctest.ProviderName, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "%[1]s-s"
  replication_group_description = "secondary"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.test.name
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 2
}
`, rName))
}

func testAccReplicationGroupConfig_GlobalReplicationGroupId_Full(rName string, numCacheClusters int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccElasticacheVpcBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "%[1]s-s"
  replication_group_description = "secondary"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.test.name

  number_cache_clusters      = %[2]d
  automatic_failover_enabled = true
  multi_az_enabled           = true

  port = 16379
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                     = "redis"
  engine_version             = "5.0.6"
  number_cache_clusters      = 2
  automatic_failover_enabled = true
  multi_az_enabled           = true

  port = 6379

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
}
`, rName, numCacheClusters))
}

func testAccReplicationGroupConfig_GlobalReplicationGroupId_ClusterMode(rName string, primaryReplicaCount, secondaryReplicaCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccElasticacheVpcBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "%[1]s-s"
  replication_group_description = "secondary"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.test.name

  automatic_failover_enabled = true
  cluster_mode {
    replicas_per_node_group = %[3]d
  }
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  engine         = "redis"
  engine_version = "6.x"
  node_type      = "cache.m5.large"

  parameter_group_name = "default.redis6.x.cluster.on"

  automatic_failover_enabled = true
  cluster_mode {
    num_node_groups         = 2
    replicas_per_node_group = %[2]d
  }
}
`, rName, primaryReplicaCount, secondaryReplicaCount))
}

func testAccReplicationGroupConfig_GlobalReplicationGroupId_ClusterMode_NumNodeGroupsOnSecondary(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccElasticacheVpcBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "%[1]s-s"
  replication_group_description = "secondary"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.test.name

  automatic_failover_enabled = true
  cluster_mode {
    num_node_groups         = 2
    replicas_per_node_group = 1
  }
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  engine         = "redis"
  engine_version = "6.x"
  node_type      = "cache.m5.large"

  parameter_group_name = "default.redis6.x.cluster.on"

  automatic_failover_enabled = true
  cluster_mode {
    num_node_groups         = 2
    replicas_per_node_group = 1
  }
}
`, rName))
}

func testAccReplicationGroupConfigDataTiering(rName string) string {
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
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
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

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.r6gd.xlarge"
  data_tiering_enabled          = true
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  auto_minor_version_upgrade    = false
}
`, rName)
}

func resourceReplicationGroupDisableAutomaticFailover(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) error {
	return resourceReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(false),
		MultiAZEnabled:           aws.Bool(false),
	})
}

func resourceReplicationGroupEnableAutomaticFailover(conn *elasticache.ElastiCache, replicationGroupID string, multiAZEnabled bool, timeout time.Duration) error {
	return resourceReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(true),
		MultiAZEnabled:           aws.Bool(multiAZEnabled),
	})
}

func resourceReplicationGroupSetPrimaryClusterID(conn *elasticache.ElastiCache, replicationGroupID, primaryClusterID string, timeout time.Duration) error {
	return resourceReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
		ApplyImmediately:   aws.Bool(true),
		PrimaryClusterId:   aws.String(primaryClusterID),
	})
}

func resourceReplicationGroupModify(conn *elasticache.ElastiCache, timeout time.Duration, input *elasticache.ModifyReplicationGroupInput) error {
	_, err := conn.ModifyReplicationGroup(input)
	if err != nil {
		return fmt.Errorf("error requesting modification: %w", err)
	}

	_, err = tfelasticache.WaitReplicationGroupAvailable(conn, aws.StringValue(input.ReplicationGroupId), timeout)
	if err != nil {
		return fmt.Errorf("error waiting for modification: %w", err)
	}
	return nil
}

func formatReplicationGroupClusterID(replicationGroupID string, clusterID int) string {
	return fmt.Sprintf("%s-%03d", replicationGroupID, clusterID)
}
