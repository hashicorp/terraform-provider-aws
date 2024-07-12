// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMemoryDBCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "acl_name", "aws_memorydb_acl.test", names.AttrID),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "cluster_endpoint.0.address", regexache.MustCompile(`^clustercfg\..*?\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "data_tiering", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "shards.#", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "shards.0.name", regexache.MustCompile(`^000[12]$`)),
					resource.TestCheckResourceAttr(resourceName, "shards.0.num_nodes", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "shards.0.slots", "0-8191"),
					resource.TestCheckResourceAttr(resourceName, "shards.0.nodes.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "shards.0.nodes.0.availability_zone"),
					acctest.CheckResourceAttrRFC3339(resourceName, "shards.0.nodes.0.create_time"),
					resource.TestMatchResourceAttr(resourceName, "shards.0.nodes.0.name", regexache.MustCompile(`^`+rName+`-000[12]-00[12]$`)),
					resource.TestMatchResourceAttr(resourceName, "shards.0.nodes.0.endpoint.0.address", regexache.MustCompile(`^`+rName+`-000[12]-00[12]\..*?\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "shards.0.nodes.0.endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSNSTopicARN, ""),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_group_name", "aws_memorydb_subnet_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "acl_name", "open-access"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "memorydb", "cluster/"+rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_endpoint.0.address"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "6379"),
					resource.TestCheckResourceAttr(resourceName, "data_tiering", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttrSet(resourceName, "engine_patch_version"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "node_type", "db.t4g.small"),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrParameterGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_window"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSNSTopicARN, ""),
					resource.TestCheckResourceAttr(resourceName, "subnet_group_name", "default"), // created automatically & matches the default vpc
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmemorydb.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBCluster_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_namePrefix(rName, "tftest-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tftest-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tftest-"),
				),
			},
		},
	})
}

// Only the open-access ACL is permitted when TLS is disabled.
func TestAccMemoryDBCluster_create_noTLS(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_noTLS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_enabled", acctest.CtFalse),
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

func TestAccMemoryDBCluster_create_withDataTiering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dataTiering(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "data_tiering", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_port(rName, 9999),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint.0.port", "9999"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "9999"),
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
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotFrom(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_memorydb_cluster.test1"),
					testAccCheckClusterExists(ctx, "aws_memorydb_cluster.test2"),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_delete_withFinalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_finalSnapshotName(rName, rName+"-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckSnapshotExistsByName(ctx, rName),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_aclName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_aclName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_description(rName, "Test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test 1"),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test 2"),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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

// As of writing, 6.2 is the one and only MemoryDB engine version available,
// so we cannot check upgrade behaviour.
//
// The API should allow upgrades with some unknown waiting time, and disallow
// downgrades.
func TestAccMemoryDBCluster_Update_engineVersion(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersionNull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_engineVersion(rName, "7.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "7.1"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_maintenanceWindow(rName, "thu:09:00-thu:10:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_nodeType(rName, "db.t4g.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
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

// As updating MemoryDB clusters can be slow, scaling up and down have been
// split into separate tests for timeout management
func TestAccMemoryDBCluster_Update_numShards_scaleUp(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numShards(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct1),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct2),
				),
			},
		},
	})
}

// As updating MemoryDB clusters can be slow, scaling up and down have been
// split into separate tests for timeout management
func TestAccMemoryDBCluster_Update_numShards_scaleDown(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numShards(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct2),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_shards", acctest.Ct1),
				),
			},
		},
	})
}

// As updating MemoryDB clusters can be slow, scaling up and down have been
// split into separate tests for timeout management
func TestAccMemoryDBCluster_Update_numReplicasPerShard_scaleUp(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numReplicasPerShard(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct1),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct2),
				),
			},
		},
	})
}

// As updating MemoryDB clusters can be slow, scaling up and down have been
// split into separate tests for timeout management
func TestAccMemoryDBCluster_Update_numReplicasPerShard_scaleDown(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numReplicasPerShard(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct1),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "num_replicas_per_shard", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccMemoryDBCluster_Update_parameterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_parameterGroup(rName, "default.memorydb-redis7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.memorydb-redis7"),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, rName),
				),
			},
			{
				Config: testAccClusterConfig_parameterGroup(rName, "default.memorydb-redis7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.memorydb-redis7"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_securityGroups(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", names.AttrID),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct2), // add one
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.1", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccClusterConfig_securityGroups(rName, 2, 0), // attempt to remove all
				ExpectError: regexache.MustCompile(`removing all security groups is not possible`),
			},
			{
				Config: testAccClusterConfig_securityGroups(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1), // remove one
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_ids.*", "aws_security_group.test.0", names.AttrID),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotRetentionLimit(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", acctest.Ct2),
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
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", acctest.Ct0),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshotWindow(rName, "00:30-01:30"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
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
					testAccCheckClusterExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snsTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, "aws_sns_topic.test", names.AttrARN),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSNSTopicARN, ""),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSNSTopicARN, "aws_sns_topic.test", names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_memorydb_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_tags2(rName, "Key1", acctest.CtValue1, "Key2", acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_tags1(rName, "Key1", acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
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
					testAccCheckClusterExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
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

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_memorydb_cluster" {
				continue
			}

			_, err := tfmemorydb.FindClusterByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

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
}

func testAccCheckClusterExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn(ctx)

		_, err := tfmemorydb.FindClusterByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccCheckSnapshotExistsByName(ctx context.Context, snapshotName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn(ctx)

		_, err := tfmemorydb.FindSnapshotByName(ctx, conn, snapshotName)

		if tfresource.NotFound(err) {
			return fmt.Errorf("MemoryDB Snapshot %s not found", snapshotName)
		}

		return err
	}
}

func testAccClusterConfig_baseNetwork(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		`
resource "aws_memorydb_subnet_group" "test" {
  subnet_ids = aws_subnet.test[*].id
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

func testAccClusterConfig_dataTiering(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseNetwork(rName),
		fmt.Sprintf(`
resource "aws_memorydb_cluster" "test" {
  acl_name          = "open-access"
  data_tiering      = true
  name              = %[1]q
  node_type         = "db.r6gd.xlarge"
  subnet_group_name = aws_memorydb_subnet_group.test.id
}
`, rName),
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
  family = "memorydb_redis7"

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
