package memorydb_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMemoryDBCluster_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "acl_name", "aws_memorydb_acl.test", "id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestMatchResourceAttr(resourceName, "cluster_endpoint.0.address", regexp.MustCompile(`^clustercfg\..*?\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "1"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "shards.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "shards.0.name", regexp.MustCompile(`^000[12]$`)),
					resource.TestCheckResourceAttr(resourceName, "shards.0.num_nodes", "2"),
					resource.TestCheckResourceAttr(resourceName, "shards.0.slots", "0-8191"),
					resource.TestCheckResourceAttr(resourceName, "shards.0.nodes.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "shards.0.nodes.0.availability_zone"),
					acctest.CheckResourceAttrRFC3339(resourceName, "shards.0.nodes.0.create_time"),
					resource.TestMatchResourceAttr(resourceName, "shards.0.nodes.0.name", regexp.MustCompile(`^`+rName+`-000[12]-00[12]$`)),
					resource.TestMatchResourceAttr(resourceName, "shards.0.nodes.0.endpoint.0.address", regexp.MustCompile(`^`+rName+`-000[12]-00[12]\..*?\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "shards.0.nodes.0.endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_group_name", "aws_memorydb_subnet_group.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", "true"),
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

func TestAccMemoryDBCluster_defaults(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl_name", "open-access"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "1"),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_group_name", "default"), // created automatically & matches the default vpc
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", "true"),
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

func TestAccMemoryDBCluster_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBCluster_nameGenerated(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNoName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_namePrefix(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNamePrefix(rName, "tftest-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tftest-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tftest-"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_create_noTLS(t *testing.T) {
	// Only the open-access ACL is permitted when TLS is disabled.

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNoTLS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", "false"),
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

func TestAccMemoryDBCluster_create_withKMS(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccMemoryDBCluster_create_withPort(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withPort(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "9999"),
					resource.TestCheckResourceAttr(resourceName, "port", "9999"),
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

func TestAccMemoryDBCluster_create_fromSnapshot(t *testing.T) {
	rName1 := "tf-test-" + sdkacctest.RandString(8)
	rName2 := "tf-test-" + sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withSnapshotFromCluster(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists("aws_memorydb_cluster.test1"),
					testAccCheckClusterExists("aws_memorydb_cluster.test2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_delete_withFinalSnapshot(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withFinalSnapshotName(rName, rName+"-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_name", rName+"-1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot_name"},
			},
			{
				Config: testAccClusterConfig_withFinalSnapshotName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"final_snapshot_name"},
			},
			{
				Config: testAccClusterConfigBaseNetwork(rName), // empty Config not supported
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExistsByName(rName),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_update_aclName(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withACLName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withACLName(rName, "open-access"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl_name", "open-access"),
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

func TestAccMemoryDBCluster_update_description(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withDescription(rName, "Test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withDescription(rName, "Test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "Test 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withDescription(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccMemoryDBCluster_update_engineVersion(t *testing.T) {
	// As of writing, 6.2 is the one and only MemoryDB engine version available,
	// so we cannot check upgrade behaviour.
	//
	// The API should allow upgrades with some unknown waiting time, and disallow
	// downgrades.

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withEngineVersionNull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withEngineVersion(rName, "6.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "6.2"),
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

func TestAccMemoryDBCluster_update_maintenanceWindow(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withMaintenanceWindow(rName, "thu:09:00-thu:10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "thu:09:00-thu:10:00"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withMaintenanceWindow(rName, "fri:09:00-fri:10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "fri:09:00-fri:10:00"),
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

func TestAccMemoryDBCluster_update_nodeType(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNodeType(rName, "db.t4g.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withNodeType(rName, "db.t4g.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.medium"),
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

func TestAccMemoryDBCluster_update_numShards_scaleUp(t *testing.T) {
	// As updating MemoryDB clusters can be slow, scaling up and down have been
	// split into separate tests for timeout management

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNumShards(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withNumShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_update_numShards_scaleDown(t *testing.T) {
	// As updating MemoryDB clusters can be slow, scaling up and down have been
	// split into separate tests for timeout management

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNumShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withNumShards(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_update_numReplicasPerShard_scaleUp(t *testing.T) {
	// As updating MemoryDB clusters can be slow, scaling up and down have been
	// split into separate tests for timeout management

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNumReplicasPerShard(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withNumReplicasPerShard(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_update_numReplicasPerShard_scaleDown(t *testing.T) {
	// As updating MemoryDB clusters can be slow, scaling up and down have been
	// split into separate tests for timeout management

	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withNumReplicasPerShard(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withNumReplicasPerShard(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "0"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_update_parameterGroup(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withParameterGroup(rName, "default.memorydb-redis6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withParameterGroup(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
				),
			},
			{
				Config: testAccClusterConfig_withParameterGroup(rName, "default.memorydb-redis6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memorydb-redis6"),
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

func TestAccMemoryDBCluster_update_securityGroupIds(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withSecurityGroups(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSecurityGroups(rName, 2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"), // add one
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.1", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccClusterConfig_withSecurityGroups(rName, 2, 0), // attempt to remove all
				ExpectError: regexp.MustCompile(`removing all security groups is not possible`),
			},
			{
				Config: testAccClusterConfig_withSecurityGroups(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"), // remove one
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", "id"),
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

func TestAccMemoryDBCluster_update_snapshotRetentionLimit(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withSnapshotRetentionLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSnapshotRetentionLimit(rName, 35),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "35"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSnapshotRetentionLimit(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "0"),
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

func TestAccMemoryDBCluster_update_snapshotWindow(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withSnapshotWindow(rName, "00:30-01:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "00:30-01:30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSnapshotWindow(rName, "02:30-03:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:30-03:30"),
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

func TestAccMemoryDBCluster_update_snsTopicArn(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withSnsTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", "aws_sns_topic.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSnsTopicNull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sns_topic_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withSnsTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", "aws_sns_topic.test", "arn"),
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

func TestAccMemoryDBCluster_update_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withTags2(rName, "Key1", "value1", "Key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withTags1(rName, "Key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_cluster" {
			continue
		}

		_, err := tfmemorydb.FindClusterByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MemoryDB Cluster %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckClusterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindClusterByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckSnapshotExistsByName(snapshotName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindSnapshotByName(context.Background(), conn, snapshotName)

		if tfresource.NotFound(err) {
			return fmt.Errorf("MemoryDB Snapshot %s not found", snapshotName)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccClusterConfigBaseNetwork(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		`
resource "aws_memorydb_subnet_group" "test" {
  subnet_ids = aws_subnet.test.*.id
}
`,
	)
}

func testAccClusterConfigBaseUserAndACL(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_user" "test" {
  access_string = "on ~* &* +@all"
  user_name     = %[1]q

  authentication_mode {
    type      = "password"
    passwords = ["aaaaaaaaaaaaaaaa"]
  }
}

resource "aws_memorydb_acl" "test" {
  name       = %[1]q
  user_names = [aws_memorydb_user.test.id]
}
`, rName)
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		testAccClusterConfigBaseUserAndACL(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id
}

resource "aws_memorydb_cluster" "test" {
  acl_name                   = aws_memorydb_acl.test.id
  auto_minor_version_upgrade = false
  name                       = %[1]q
  node_type                  = "db.t4g.small"
  num_shards                 = 2
  security_group_ids         = [aws_security_group.test.id]
  snapshot_retention_limit   = 7
  subnet_group_name          = aws_memorydb_subnet_group.test.id

  tags = {
    Test = "test"
  }
}
`, rName),
	)
}

func testAccClusterConfig_defaults(rName string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name  = "open-access"
  name      = %[1]q
  node_type = "db.t4g.small"
}
`, rName),
	)
}

func testAccClusterConfig_withNoName(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`,
	)
}

func testAccClusterConfig_withNamePrefix(rName, prefix string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name_prefix            = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, prefix),
	)
}

func testAccClusterConfig_withNoTLS(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
  tls_enabled            = false
}
`, rName),
	)
}

func testAccClusterConfig_withACLName(rName, aclName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		testAccClusterConfigBaseUserAndACL(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  depends_on             = [aws_memorydb_acl.test]
  acl_name               = %[2]q
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, aclName),
	)
}

func testAccClusterConfig_withDescription(rName, description string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name          = "open-access"
  description       = %[2]q
  name              = %[1]q
  node_type         = "db.t4g.small"
  subnet_group_name = aws_memorydb_subnet_group.test.id
}
`, rName, description),
	)
}

func testAccClusterConfig_withEngineVersionNull(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccClusterConfig_withEngineVersion(rName, engineVersion string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  engine_version         = %[2]q
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, engineVersion),
	)
}

func testAccClusterConfig_withFinalSnapshotName(rName, finalSnapshotName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  final_snapshot_name    = %[2]q
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, finalSnapshotName),
	)
}

func testAccClusterConfig_withKMS(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  kms_key_arn            = aws_kms_key.test.arn
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccClusterConfig_withMaintenanceWindow(rName, maintenanceWindow string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  maintenance_window     = %[2]q
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, maintenanceWindow),
	)
}

func testAccClusterConfig_withNodeType(rName, nodeType string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = %[2]q
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, nodeType),
	)
}

func testAccClusterConfig_withNumReplicasPerShard(rName string, numReplicasPerShard int) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = %[2]d
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, numReplicasPerShard),
	)
}

func testAccClusterConfig_withNumShards(rName string, numShards int) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name          = "open-access"
  name              = %[1]q
  node_type         = "db.t4g.small"
  num_shards        = %[2]d
  subnet_group_name = aws_memorydb_subnet_group.test.id
}
`, rName, numShards),
	)
}

func testAccClusterConfig_withParameterGroup(rName, parameterGroup string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  parameter {
    name  = "active-defrag-cycle-max"
    value = "70"
  }

  parameter {
    name  = "active-defrag-cycle-min"
    value = "10"
  }
}

resource "aws_memorydb_cluster" "test" {
  depends_on             = [aws_memorydb_parameter_group.test]
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  parameter_group_name   = %[2]q
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, parameterGroup),
	)
}

func testAccClusterConfig_withPort(rName string, port int) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  port                   = %[2]d
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, port),
	)
}

func testAccClusterConfig_withSecurityGroups(rName string, sgCount, sgCountInCluster int) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  count  = %[2]d
  vpc_id = aws_vpc.test.id
}

resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  security_group_ids     = slice(aws_security_group.test[*].id, 0, %[3]d)
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName, sgCount, sgCountInCluster),
	)
}

func testAccClusterConfig_withSnapshotRetentionLimit(rName string, retentionLimit int) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name                 = "open-access"
  name                     = %[1]q
  node_type                = "db.t4g.small"
  num_replicas_per_shard   = 0
  num_shards               = 1
  snapshot_retention_limit = %[2]d
  subnet_group_name        = aws_memorydb_subnet_group.test.id
}
`, rName, retentionLimit),
	)
}

func testAccClusterConfig_withSnapshotFromCluster(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName1),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test1" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}

resource "aws_memorydb_snapshot" "test" {
  cluster_name = aws_memorydb_cluster.test1.name
}

resource "aws_memorydb_cluster" "test2" {
  acl_name               = "open-access"
  name                   = %[2]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  snapshot_name          = aws_memorydb_snapshot.test.name
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName1, rName2),
	)
}

func testAccClusterConfig_withSnapshotWindow(rName, snapshotWindow string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name                 = "open-access"
  name                     = %[1]q
  node_type                = "db.t4g.small"
  num_replicas_per_shard   = 0
  num_shards               = 1
  snapshot_retention_limit = 1
  snapshot_window          = %[2]q
  subnet_group_name        = aws_memorydb_subnet_group.test.id
}
`, rName, snapshotWindow),
	)
}

func testAccClusterConfig_withSnsTopicNull(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_memorydb_cluster" "test" {
  depends_on               = [aws_sns_topic.test]
  acl_name                 = "open-access"
  name                     = %[1]q
  node_type                = "db.t4g.small"
  num_replicas_per_shard   = 0
  num_shards               = 1
  snapshot_retention_limit = 1
  subnet_group_name        = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccClusterConfig_withSnsTopic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_memorydb_cluster" "test" {
  depends_on               = [aws_sns_topic.test]
  acl_name                 = "open-access"
  name                     = %[1]q
  node_type                = "db.t4g.small"
  num_replicas_per_shard   = 0
  num_shards               = 1
  snapshot_retention_limit = 1
  sns_topic_arn            = aws_sns_topic.test.arn
  subnet_group_name        = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccClusterConfig_withTags0(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id
}
`, rName),
	)
}

func testAccClusterConfig_withTags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value),
	)
}

func testAccClusterConfig_withTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccClusterConfigBaseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name               = "open-access"
  name                   = %[1]q
  node_type              = "db.t4g.small"
  num_replicas_per_shard = 0
  num_shards             = 1
  subnet_group_name      = aws_memorydb_subnet_group.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value),
	)
}
