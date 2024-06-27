// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDocDBGlobalCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					//This is a rds arn
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("global-cluster:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttr(resourceName, "global_cluster_identifier", rName),
					resource.TestMatchResourceAttr(resourceName, "global_cluster_resource_id", regexache.MustCompile(`cluster-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
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

func TestAccDocDBGlobalCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdocdb.ResourceGlobalCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBGlobalCluster_DatabaseName(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "database1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_databaseName(rName, "database2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDatabaseName, "database2"),
				),
			},
		},
	})
}

func TestAccDocDBGlobalCluster_DeletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_deletionProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterNotRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDocDBGlobalCluster_Engine(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engine(rName, "docdb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "docdb"),
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

func TestAccDocDBGlobalCluster_EngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName, "docdb", "4.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "4.0.0"),
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

func TestAccDocDBGlobalCluster_SourceDBClusterIdentifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccDocDBGlobalCluster_SourceDBClusterIdentifier_storageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_docdb_cluster.test"
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBIdentifierStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster),
					resource.TestCheckResourceAttrPair(resourceName, "source_db_cluster_identifier", clusterResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_db_cluster_identifier"},
			},
		},
	})
}

func TestAccDocDBGlobalCluster_StorageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 awstypes.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_docdb_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DocDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_storageEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &globalCluster2),
					testAccCheckGlobalClusterRecreated(&globalCluster1, &globalCluster2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageEncrypted, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckGlobalClusterExists(ctx context.Context, n string, v *awstypes.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		output, err := tfdocdb.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_docdb_global_cluster" {
				continue
			}

			_, err := tfdocdb.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DocumentDB Global Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGlobalClusterNotRecreated(i, j *awstypes.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.GlobalClusterArn) != aws.ToString(j.GlobalClusterArn) {
			return fmt.Errorf("DocumentDB Global Cluster was recreated. got: %s, expected: %s", aws.ToString(i.GlobalClusterArn), aws.ToString(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckGlobalClusterRecreated(i, j *awstypes.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.GlobalClusterResourceId) == aws.ToString(j.GlobalClusterResourceId) {
			return errors.New("DocumentDB Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckGlobalCluster(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBClient(ctx)

	input := &docdb.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClusters(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		// Current Region/Partition does not support DocumentDB Global Clusters
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  global_cluster_identifier = %[1]q
}
`, rName)
}

func testAccGlobalClusterConfig_databaseName(rName, databaseName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  database_name             = %[1]q
  global_cluster_identifier = %[2]q
}
`, databaseName, rName)
}

func testAccGlobalClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = "docdb"
  deletion_protection       = %[2]t
  global_cluster_identifier = %[1]q
}
`, rName, deletionProtection)
}

func testAccGlobalClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %[1]q
  global_cluster_identifier = %[2]q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_engineVersion(rName, engine, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  engine                    = %[1]q
  engine_version            = %[2]q
  global_cluster_identifier = %[3]q
}
`, engine, engineVersion, rName)
}

func testAccGlobalClusterConfig_sourceDBIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "docdb"
  engine_version      = "4.0.0" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_sourceDBIdentifierStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster" "test" {
  cluster_identifier  = %[1]q
  engine              = "docdb"
  engine_version      = "4.0.0" # Minimum supported version for Global Clusters
  master_password     = "mustbeeightcharacters"
  master_username     = "test"
  skip_final_snapshot = true
  storage_encrypted   = true

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_docdb_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_storageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_docdb_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "docdb"
  storage_encrypted         = %[2]t
}
`, rName, storageEncrypted)
}
