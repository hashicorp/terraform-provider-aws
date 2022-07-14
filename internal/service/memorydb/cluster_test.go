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
				Config: testAccClusterConfig_noName(rName),
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
				Config: testAccClusterConfig_namePrefix(rName, "tftest-"),
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
				Config: testAccClusterConfig_noTLS(rName),
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
				Config: testAccClusterConfig_kms(rName),
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
				Config: testAccClusterConfig_port(rName, 9999),
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
				Config: testAccClusterConfig_snapshotFrom(rName1, rName2),
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
				Config: testAccClusterConfig_finalSnapshotName(rName, rName+"-1"),
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
				Config: testAccClusterConfig_finalSnapshotName(rName, rName),
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
				Config: testAccClusterConfig_baseNetwork(rName), // empty Config not supported
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExistsByName(rName),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_aclName(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_aclName(rName, rName),
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
				Config: testAccClusterConfig_aclName(rName, "open-access"),
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

func TestAccMemoryDBCluster_Update_description(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_description(rName, "Test 1"),
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
				Config: testAccClusterConfig_description(rName, "Test 2"),
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
				Config: testAccClusterConfig_description(rName, ""),
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

func TestAccMemoryDBCluster_Update_engineVersion(t *testing.T) {
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
				Config: testAccClusterConfig_engineVersionNull(rName),
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
				Config: testAccClusterConfig_engineVersion(rName, "6.2"),
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

func TestAccMemoryDBCluster_Update_maintenanceWindow(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_maintenanceWindow(rName, "thu:09:00-thu:10:00"),
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
				Config: testAccClusterConfig_maintenanceWindow(rName, "fri:09:00-fri:10:00"),
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

func TestAccMemoryDBCluster_Update_nodeType(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_nodeType(rName, "db.t4g.small"),
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
				Config: testAccClusterConfig_nodeType(rName, "db.t4g.medium"),
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

func TestAccMemoryDBCluster_Update_numShards_scaleUp(t *testing.T) {
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
				Config: testAccClusterConfig_numShards(rName, 1),
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
				Config: testAccClusterConfig_numShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_numShards_scaleDown(t *testing.T) {
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
				Config: testAccClusterConfig_numShards(rName, 2),
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
				Config: testAccClusterConfig_numShards(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", "1"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_numReplicasPerShard_scaleUp(t *testing.T) {
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
				Config: testAccClusterConfig_numReplicasPerShard(rName, 1),
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
				Config: testAccClusterConfig_numReplicasPerShard(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_numReplicasPerShard_scaleDown(t *testing.T) {
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
				Config: testAccClusterConfig_numReplicasPerShard(rName, 1),
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
				Config: testAccClusterConfig_numReplicasPerShard(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", "0"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_parameterGroup(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_parameterGroup(rName, "default.memorydb-redis6"),
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
				Config: testAccClusterConfig_parameterGroup(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", rName),
				),
			},
			{
				Config: testAccClusterConfig_parameterGroup(rName, "default.memorydb-redis6"),
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

func TestAccMemoryDBCluster_Update_securityGroupIds(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_securityGroups(rName, 2, 1),
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
				Config: testAccClusterConfig_securityGroups(rName, 2, 2),
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
				Config:      testAccClusterConfig_securityGroups(rName, 2, 0), // attempt to remove all
				ExpectError: regexp.MustCompile(`removing all security groups is not possible`),
			},
			{
				Config: testAccClusterConfig_securityGroups(rName, 2, 1),
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

func TestAccMemoryDBCluster_Update_snapshotRetentionLimit(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotRetentionLimit(rName, 2),
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
				Config: testAccClusterConfig_snapshotRetentionLimit(rName, 35),
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
				Config: testAccClusterConfig_snapshotRetentionLimit(rName, 0),
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

func TestAccMemoryDBCluster_Update_snapshotWindow(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotWindow(rName, "00:30-01:30"),
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
				Config: testAccClusterConfig_snapshotWindow(rName, "02:30-03:30"),
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

func TestAccMemoryDBCluster_Update_snsTopicARN(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snsTopic(rName),
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
				Config: testAccClusterConfig_snsTopicNull(rName),
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
				Config: testAccClusterConfig_snsTopic(rName),
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

func TestAccMemoryDBCluster_Update_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags0(rName),
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
				Config: testAccClusterConfig_tags2(rName, "Key1", "value1", "Key2", "value2"),
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
				Config: testAccClusterConfig_tags1(rName, "Key1", "value1"),
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
				Config: testAccClusterConfig_tags0(rName),
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

func testAccClusterConfig_baseNetwork(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_noName(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_namePrefix(rName, prefix string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_noTLS(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_aclName(rName, aclName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_engineVersionNull(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_engineVersion(rName, engineVersion string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_finalSnapshotName(rName, finalSnapshotName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_kms(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_maintenanceWindow(rName, maintenanceWindow string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_nodeType(rName, nodeType string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_numReplicasPerShard(rName string, numReplicasPerShard int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_numShards(rName string, numShards int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_parameterGroup(rName, parameterGroup string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_port(rName string, port int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_securityGroups(rName string, sgCount, sgCountInCluster int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_snapshotRetentionLimit(rName string, retentionLimit int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_snapshotFrom(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName1),
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

func testAccClusterConfig_snapshotWindow(rName, snapshotWindow string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_snsTopicNull(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_snsTopic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_tags0(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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

func testAccClusterConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
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
