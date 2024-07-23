// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSCustomDBEngineVersion_sqlServer(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Windows SQL Server AMI owned by operating account set as environmental variable
	// Blog: https://aws.amazon.com/blogs/database/persist-your-os-level-customization-within-amazon-rds-custom-for-sql-server-using-custom-engine-version-cev/
	key := "RDS_CUSTOM_WINDOWS_SQLSERVER_AMI"
	ami := os.Getenv(key)
	if ami == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "15.00.4249.2.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_sqlServer(rName, ami),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds",
						regexache.MustCompile(fmt.Sprintf(`cev:custom-sqlserver.+%s.+`, rName))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "manifest_hash", "manifest", "source_image_id"},
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_sqlServerUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Windows SQL Server AMI owned by operating account set as environmental variable
	key := "RDS_CUSTOM_WINDOWS_SQLSERVER_AMI"
	ami := os.Getenv(key)
	if ami == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "15.00.4249.2.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	resourceName := "aws_rds_custom_db_engine_version.test"
	status := "pending-validation"
	description2 := "inactive"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_sqlServer(rName, ami),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, status),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "manifest_hash", "manifest", "source_image_id"},
			},
			{
				Config: testAccCustomDBEngineVersionConfig_sqlServerUpdate(rName, ami, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description2),
				),
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_oracle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Oracle installation media in S3 bucket owned (bucket must be in operating region) by operating account set as environmental variable
	// Pre-requisite: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-cev.preparing.html
	key := "RDS_CUSTOM_ORACLE_S3_BUCKET"
	bucket := os.Getenv(key)
	if bucket == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "19.19.ee.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_oracle(rName, bucket),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf(`cev:custom-oracle.+%s.+`, rName))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "manifest_hash", "manifest"},
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_manifestFile(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Oracle installation media in S3 bucket owned (bucket must be in operating region) by operating account set as environmental variable
	// Pre-requisite: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/custom-cev.preparing.html
	key := "RDS_CUSTOM_ORACLE_S3_BUCKET"
	bucket := os.Getenv(key)
	if bucket == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "19.19.ee.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	filename := "test-fixtures/custom-oracle-manifest.json"
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_manifestFile(rName, bucket, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", regexache.MustCompile(fmt.Sprintf(`cev:custom-oracle.+%s.+`, rName))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "manifest_hash", "manifest"},
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Windows SQL Server AMI owned by operating account set as environmental variable
	key := "RDS_CUSTOM_WINDOWS_SQLSERVER_AMI"
	ami := os.Getenv(key)
	if ami == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "15.00.4249.2.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_tags(rName, ami, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccRDSCustomDBEngineVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// Requires an existing Windows SQL Server AMI owned by operating account set as environmental variable
	key := "RDS_CUSTOM_WINDOWS_SQLSERVER_AMI"
	ami := os.Getenv(key)
	if ami == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	var customdbengineversion rds.DBEngineVersion
	rName := fmt.Sprintf("%s%s%d", "15.00.4249.2.", acctest.ResourcePrefix, sdkacctest.RandIntRange(100, 999))
	resourceName := "aws_rds_custom_db_engine_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, rds.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDBEngineVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDBEngineVersionConfig_sqlServer(rName, ami),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDBEngineVersionExists(ctx, resourceName, &customdbengineversion),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfrds.ResourceCustomDBEngineVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomDBEngineVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_custom_db_engine_version" {
				continue
			}

			_, err := tfrds.FindCustomDBEngineVersionByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.RDS, create.ErrActionCheckingDestroyed, tfrds.ResNameCustomDBEngineVersion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCustomDBEngineVersionExists(ctx context.Context, name string, customdbengineversion *rds.DBEngineVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindCustomDBEngineVersionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.RDS, create.ErrActionCheckingExistence, tfrds.ResNameCustomDBEngineVersion, rs.Primary.ID, err)
		}

		*customdbengineversion = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBEngineVersionsInput{}
	_, err := conn.DescribeDBEngineVersionsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCustomDBEngineVersionConfig_sqlServer(rName, ami string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

# Copy the Amazon AMI for Windows SQL Server, CEV creation requires an AMI owned by the operator
resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = %[2]q
  source_ami_region = data.aws_region.current.name
}

resource "aws_rds_custom_db_engine_version" "test" {
  engine          = "custom-sqlserver-se"
  engine_version  = %[1]q
  source_image_id = aws_ami_copy.test.id
}
`, rName, ami)
}

func testAccCustomDBEngineVersionConfig_sqlServerUpdate(rName, ami, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

# Copy the Amazon AMI for Windows SQL Server, CEV creation requires an AMI owned by the operator
resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = %[2]q
  source_ami_region = data.aws_region.current.name
}

resource "aws_rds_custom_db_engine_version" "test" {
  description     = %[3]q
  engine          = "custom-sqlserver-se"
  engine_version  = %[1]q
  source_image_id = aws_ami_copy.test.id
}
`, rName, ami, description)
}

func testAccCustomDBEngineVersionConfig_oracle(rName, bucket string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "rdscfo_kms_key" {
  description = "KMS symmetric key for RDS Custom for Oracle"
}

resource "aws_rds_custom_db_engine_version" "test" {
  database_installation_files_s3_bucket_name = %[2]q
  engine                                     = "custom-oracle-ee-cdb"
  engine_version                             = %[1]q
  kms_key_id                                 = aws_kms_key.rdscfo_kms_key.arn
  manifest                                   = <<JSON
  {
	"databaseInstallationFileNames":["V982063-01.zip"]
  }
  JSON
}
`, rName, bucket)
}

func testAccCustomDBEngineVersionConfig_manifestFile(rName, bucket, filename string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "rdscfo_kms_key" {
  description = "KMS symmetric key for RDS Custom for Oracle"
}

resource "aws_rds_custom_db_engine_version" "test" {
  database_installation_files_s3_bucket_name = %[2]q
  engine                                     = "custom-oracle-ee-cdb"
  engine_version                             = %[1]q
  kms_key_id                                 = aws_kms_key.rdscfo_kms_key.arn
  filename                                   = %[3]q
  manifest_hash                              = filebase64sha256(%[3]q)
}
`, rName, bucket, filename)
}

func testAccCustomDBEngineVersionConfig_tags(rName, ami, key, value string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

# Copy the Amazon AMI for Windows SQL Server, CEV creation requires an AMI owned by the operator
resource "aws_ami_copy" "test" {
  name              = %[1]q
  source_ami_id     = %[2]q
  source_ami_region = data.aws_region.current.name
}

resource "aws_rds_custom_db_engine_version" "test" {
  engine          = "custom-sqlserver-se"
  engine_version  = %[1]q
  source_image_id = aws_ami_copy.test.id

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, ami, key, value)
}
