// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxDataRepositoryAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath := fmt.Sprintf("s3://%s", rName)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, rName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`association/fs-.+/dra-.+`)),
					resource.TestCheckResourceAttr(resourceName, "batch_import_meta_data_on_create", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "data_repository_path", bucketPath),
					resource.TestMatchResourceAttr(resourceName, names.AttrFileSystemID, regexache.MustCompile(`fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "file_system_path", fileSystemPath),
					resource.TestCheckResourceAttrSet(resourceName, "imported_file_chunk_size"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, rName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceDataRepositoryAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_disappears_ParentFileSystem(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	parentResourceName := "aws_fsx_lustre_file_system.test"
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, rName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceLustreFileSystem(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_fileSystemPathUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var association1, association2 awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath1 := "/test1"
	fileSystemPath2 := "/test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, rName, fileSystemPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association1),
					resource.TestCheckResourceAttr(resourceName, "file_system_path", fileSystemPath1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, rName, fileSystemPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association2),
					testAccCheckDataRepositoryAssociationRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "file_system_path", fileSystemPath2),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_dataRepositoryPathUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var association1, association2 awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath1 := fmt.Sprintf("s3://%s", bucketName1)
	bucketName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath2 := fmt.Sprintf("s3://%s", bucketName2)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, bucketName1, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association1),
					resource.TestCheckResourceAttr(resourceName, "data_repository_path", bucketPath1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(rName, bucketName2, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association2),
					testAccCheckDataRepositoryAssociationRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "data_repository_path", bucketPath2),
				),
			},
		},
	})
}

// lintignore:AT002
func TestAccFSxDataRepositoryAssociation_importedFileChunkSize(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(rName, rName, fileSystemPath, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "256"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

// lintignore:AT002
func TestAccFSxDataRepositoryAssociation_importedFileChunkSizeUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var association1, association2 awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(rName, rName, fileSystemPath, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association1),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "256"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
			{
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(rName, rName, fileSystemPath, 512),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "512"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_deleteDataInFilesystem(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_deleteInFilesystem(rName, rName, fileSystemPath, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "delete_data_in_filesystem", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3AutoExportPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events := []string{"NEW", "CHANGED", "DELETED"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(rName, rName, fileSystemPath, events),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.2", "DELETED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3AutoExportPolicyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var association1, association2 awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events1 := []string{"NEW", "CHANGED", "DELETED"}
	events2 := []string{"NEW"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(rName, rName, fileSystemPath, events1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association1),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.2", "DELETED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(rName, rName, fileSystemPath, events2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.0", "NEW"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3AutoImportPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events := []string{"NEW", "CHANGED", "DELETED"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(rName, rName, fileSystemPath, events),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.2", "DELETED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3AutoImportPolicyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var association1, association2 awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events1 := []string{"NEW", "CHANGED", "DELETED"}
	events2 := []string{"NEW"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(rName, rName, fileSystemPath, events1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association1),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.2", "DELETED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(rName, rName, fileSystemPath, events2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.0", "NEW"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3FullPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
			// PERSISTENT_2 deployment_type is not supported in GovCloud partition.
			acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3FullPolicy(rName, rName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.2", "DELETED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.0", "NEW"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.1", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.2", "DELETED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"delete_data_in_filesystem"},
			},
		},
	})
}

func testAccCheckDataRepositoryAssociationExists(ctx context.Context, n string, v *awstypes.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindDataRepositoryAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDataRepositoryAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_data_repository_association" {
				continue
			}

			_, err := tffsx.FindDataRepositoryAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for Lustre Data Repository Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataRepositoryAssociationNotRecreated(i, j *awstypes.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.AssociationId) != aws.ToString(j.AssociationId) {
			return fmt.Errorf("FSx Data Repository Association (%s) recreated", aws.ToString(i.AssociationId))
		}

		return nil
	}
}

func testAccCheckDataRepositoryAssociationRecreated(i, j *awstypes.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.AssociationId) == aws.ToString(j.AssociationId) {
			return fmt.Errorf("FSx Data Repository Association (%s) not recreated", aws.ToString(i.AssociationId))
		}

		return nil
	}
}

func testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = aws_subnet.test[*].id
  deployment_type             = "PERSISTENT_2"
  per_unit_storage_throughput = 125
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, bucketName))
}

func testAccDataRepositoryAssociationConfig_fileSystemPath(rName, bucketName, fileSystemPath string) string {
	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id       = aws_fsx_lustre_file_system.test.id
  data_repository_path = "s3://%[1]s"
  file_system_path     = %[2]q
}
`, bucketName, fileSystemPath))
}

func testAccDataRepositoryAssociationConfig_importedFileChunkSize(rName, bucketName, fileSystemPath string, fileChunkSize int64) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)

	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id           = aws_fsx_lustre_file_system.test.id
  data_repository_path     = %[1]q
  file_system_path         = %[2]q
  imported_file_chunk_size = %[3]d
}
`, bucketPath, fileSystemPath, fileChunkSize))
}

func testAccDataRepositoryAssociationConfig_deleteInFilesystem(rName, bucketName, fileSystemPath, deleteDataInFilesystem string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id            = aws_fsx_lustre_file_system.test.id
  data_repository_path      = %[1]q
  file_system_path          = %[2]q
  delete_data_in_filesystem = %[3]q
}
`, bucketPath, fileSystemPath, deleteDataInFilesystem))
}

func testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(rName, bucketName, fileSystemPath string, events []string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	eventsString := strings.Replace(fmt.Sprintf("%q", events), " ", ", ", -1)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id       = aws_fsx_lustre_file_system.test.id
  data_repository_path = %[1]q
  file_system_path     = %[2]q

  s3 {
    auto_export_policy {
      events = %[3]s
    }
  }
}
`, bucketPath, fileSystemPath, eventsString))
}

func testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(rName, bucketName, fileSystemPath string, events []string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	eventsString := strings.Replace(fmt.Sprintf("%q", events), " ", ", ", -1)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id       = aws_fsx_lustre_file_system.test.id
  data_repository_path = %[1]q
  file_system_path     = %[2]q

  s3 {
    auto_import_policy {
      events = %[3]s
    }
  }
}
`, bucketPath, fileSystemPath, eventsString))
}

func testAccDataRepositoryAssociationConfig_s3FullPolicy(rName, bucketName, fileSystemPath string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationConfig_s3Bucket(rName, bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id       = aws_fsx_lustre_file_system.test.id
  data_repository_path = %[1]q
  file_system_path     = %[2]q

  s3 {
    auto_export_policy {
      events = ["NEW", "CHANGED", "DELETED"]
    }

    auto_import_policy {
      events = ["NEW", "CHANGED", "DELETED"]
    }
  }
}
`, bucketPath, fileSystemPath))
}
