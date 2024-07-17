// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGlobalCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("global-cluster:%s", rName)),
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

func TestAccNeptuneGlobalCluster_completeBasic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_completeBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("global-cluster:%s", rName)),
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

func TestAccNeptuneGlobalCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfneptune.ResourceGlobalCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneGlobalCluster_DeletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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

func TestAccNeptuneGlobalCluster_Engine(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engine(rName, "neptune"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "neptune"),
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

func TestAccNeptuneGlobalCluster_EngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName1, rName2, rName3, "1.2.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.2.0.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalClusterConfig_engineVersion(rName1, rName2, rName3, "1.2.0.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.2.0.1"),
				),
			},
		},
	})
}

func TestAccNeptuneGlobalCluster_SourceDBClusterIdentifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_neptune_cluster.test"
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
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

func TestAccNeptuneGlobalCluster_SourceDBClusterIdentifier_storageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_neptune_cluster.test"
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClusterConfig_sourceDBIdentifierStorageEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalClusterExists(ctx, resourceName, &v),
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

func TestAccNeptuneGlobalCluster_StorageEncrypted(t *testing.T) {
	ctx := acctest.Context(t)
	var globalCluster1, globalCluster2 neptune.GlobalCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_global_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalCluster(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
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

func testAccCheckGlobalClusterExists(ctx context.Context, n string, v *neptune.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		output, err := tfneptune.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGlobalClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_global_cluster" {
				continue
			}

			_, err := tfneptune.FindGlobalClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Global Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGlobalClusterNotRecreated(i, j *neptune.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterArn) != aws.StringValue(j.GlobalClusterArn) {
			return fmt.Errorf("Neptune Global Cluster was recreated. got: %s, expected: %s", aws.StringValue(i.GlobalClusterArn), aws.StringValue(j.GlobalClusterArn))
		}

		return nil
	}
}

func testAccCheckGlobalClusterRecreated(i, j *neptune.GlobalCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.GlobalClusterResourceId) == aws.StringValue(j.GlobalClusterResourceId) {
			return errors.New("Neptune Global Cluster was not recreated")
		}

		return nil
	}
}

func testAccPreCheckGlobalCluster(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

	input := &neptune.DescribeGlobalClustersInput{}

	_, err := conn.DescribeGlobalClustersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValue", "Access Denied to API Version: APIGlobalDatabases") {
		// Current Region/Partition does not support Neptune Global Clusters
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalClusterConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  engine                    = "neptune"
  engine_version            = "1.2.0.0"
  global_cluster_identifier = %[1]q
}
`, rName)
}

func testAccGlobalClusterConfig_deletionProtection(rName string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  engine                    = "neptune"
  deletion_protection       = %[1]t
  engine_version            = "1.2.0.0"
  global_cluster_identifier = %[2]q
}
`, deletionProtection, rName)
}

func testAccGlobalClusterConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  engine                    = %[1]q
  engine_version            = "1.2.0.0"
  global_cluster_identifier = %[2]q
}
`, engine, rName)
}

func testAccGlobalClusterConfig_engineVersion(rName1, rName2, rName3, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  engine                    = "neptune"
  engine_version            = %[4]q
  global_cluster_identifier = %[1]q
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[2]q
  skip_final_snapshot                  = true
  global_cluster_identifier            = aws_neptune_global_cluster.test.id
  engine                               = aws_neptune_global_cluster.test.engine
  engine_version                       = aws_neptune_global_cluster.test.engine_version
  neptune_cluster_parameter_group_name = "default.neptune1.2"
  apply_immediately                    = true
}

data "aws_neptune_orderable_db_instance" "test" {
  engine         = "neptune"
  engine_version = aws_neptune_cluster.test.engine_version
  license_model  = "amazon-license"

  preferred_instance_classes = ["db.r5.large", "db.r5.xlarge"]
}

resource "aws_neptune_cluster_instance" "test" {
  identifier                   = %[3]q
  cluster_identifier           = aws_neptune_cluster.test.id
  apply_immediately            = true
  instance_class               = data.aws_neptune_orderable_db_instance.test.instance_class
  neptune_parameter_group_name = aws_neptune_cluster.test.neptune_cluster_parameter_group_name
  promotion_tier               = "3"
}
`, rName1, rName2, rName3, engineVersion)
}

func testAccGlobalClusterConfig_completeBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  engine                    = "neptune"
  engine_version            = "1.2.0.0"
  global_cluster_identifier = %[1]q
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  engine                               = "neptune"
  engine_version                       = "1.2.0.0"
  skip_final_snapshot                  = true
  neptune_cluster_parameter_group_name = "default.neptune1.2"
  global_cluster_identifier            = aws_neptune_global_cluster.test.id
}
`, rName)
}

func testAccGlobalClusterConfig_sourceDBIdentifier(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  engine                               = "neptune"
  engine_version                       = "1.2.0.0"
  skip_final_snapshot                  = true
  neptune_cluster_parameter_group_name = "default.neptune1.2"

  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_neptune_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_neptune_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_sourceDBIdentifierStorageEncrypted(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  engine                               = "neptune"
  engine_version                       = "1.2.0.0"
  skip_final_snapshot                  = true
  storage_encrypted                    = true
  neptune_cluster_parameter_group_name = "default.neptune1.2"
  # global_cluster_identifier cannot be Computed

  lifecycle {
    ignore_changes = [global_cluster_identifier]
  }
}

resource "aws_neptune_global_cluster" "test" {
  global_cluster_identifier    = %[1]q
  source_db_cluster_identifier = aws_neptune_cluster.test.arn
}
`, rName)
}

func testAccGlobalClusterConfig_storageEncrypted(rName string, storageEncrypted bool) string {
	return fmt.Sprintf(`
resource "aws_neptune_global_cluster" "test" {
  global_cluster_identifier = %[1]q
  engine                    = "neptune"
  engine_version            = "1.2.0.0"
  storage_encrypted         = %[2]t
}
`, rName, storageEncrypted)
}
