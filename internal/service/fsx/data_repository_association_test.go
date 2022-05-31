package fsx_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxDataRepositoryAssociation_basic(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`association/fs-.+/dra-.+`)),
					resource.TestCheckResourceAttr(resourceName, "batch_import_meta_data_on_create", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_path", bucketPath),
					resource.TestMatchResourceAttr(resourceName, "file_system_id", regexp.MustCompile(`fs-.+`)),
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
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceDataRepositoryAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_disappears_ParentFileSystem(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	parentResourceName := "aws_fsx_lustre_file_system.test"
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceLustreFileSystem(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_fileSystemPathUpdated(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association1, association2 fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath1 := "/test1"
	fileSystemPath2 := "/test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association1),
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
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association2),
					testAccCheckDataRepositoryAssociationRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "file_system_path", fileSystemPath2),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_dataRepositoryPathUpdated(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association1, association2 fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath1 := fmt.Sprintf("s3://%s", bucketName1)
	bucketName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketPath2 := fmt.Sprintf("s3://%s", bucketName2)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName1, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association1),
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
				Config: testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName2, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association2),
					testAccCheckDataRepositoryAssociationRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "data_repository_path", bucketPath2),
				),
			},
		},
	})
}

//lintignore:AT002
func TestAccFSxDataRepositoryAssociation_importedFileChunkSize(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(bucketName, fileSystemPath, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
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

//lintignore:AT002
func TestAccFSxDataRepositoryAssociation_importedFileChunkSizeUpdated(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association1, association2 fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(bucketName, fileSystemPath, 256),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association1),
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
				Config: testAccDataRepositoryAssociationConfig_importedFileChunkSize(bucketName, fileSystemPath, 512),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "512"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_deleteDataInFilesystem(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_deleteInFilesystem(bucketName, fileSystemPath, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "delete_data_in_filesystem", "true"),
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
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events := []string{"NEW", "CHANGED", "DELETED"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(bucketName, fileSystemPath, events),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
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
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association1, association2 fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events1 := []string{"NEW", "CHANGED", "DELETED"}
	events2 := []string{"NEW"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(bucketName, fileSystemPath, events1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association1),
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
				Config: testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(bucketName, fileSystemPath, events2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_export_policy.0.events.0", "NEW"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3AutoImportPolicy(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events := []string{"NEW", "CHANGED", "DELETED"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(bucketName, fileSystemPath, events),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
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
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association1, association2 fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"
	events1 := []string{"NEW", "CHANGED", "DELETED"}
	events2 := []string{"NEW"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(bucketName, fileSystemPath, events1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association1),
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
				Config: testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(bucketName, fileSystemPath, events2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association2),
					testAccCheckDataRepositoryAssociationNotRecreated(&association1, &association2),
					resource.TestCheckResourceAttr(resourceName, "s3.0.auto_import_policy.0.events.0", "NEW"),
				),
			},
		},
	})
}

func TestAccFSxDataRepositoryAssociation_s3FullPolicy(t *testing.T) {
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		t.Skip("PERSISTENT_2 deployment_type is not supported in GovCloud partition")
	}

	var association fsx.DataRepositoryAssociation
	resourceName := "aws_fsx_data_repository_association.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	fileSystemPath := "/test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataRepositoryAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataRepositoryAssociationConfig_s3FullPolicy(bucketName, fileSystemPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataRepositoryAssociationExists(resourceName, &association),
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

func testAccCheckDataRepositoryAssociationExists(resourceName string, assoc *fsx.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		association, err := tffsx.FindDataRepositoryAssociationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if association == nil {
			return fmt.Errorf("FSx Lustre Data Repository Association (%s) not found", rs.Primary.ID)
		}

		*assoc = *association

		return nil
	}
}

func testAccCheckDataRepositoryAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_lustre_file_system" {
			continue
		}

		filesystem, err := tffsx.FindFileSystemByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if filesystem != nil {
			return fmt.Errorf("FSx Lustre File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckDataRepositoryAssociationNotRecreated(i, j *fsx.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.AssociationId) != aws.StringValue(j.AssociationId) {
			return fmt.Errorf("FSx Data Repository Association (%s) recreated", aws.StringValue(i.AssociationId))
		}

		return nil
	}
}

func testAccCheckDataRepositoryAssociationRecreated(i, j *fsx.DataRepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.AssociationId) == aws.StringValue(j.AssociationId) {
			return fmt.Errorf("FSx Data Repository Association (%s) not recreated", aws.StringValue(i.AssociationId))
		}

		return nil
	}
}

func testAccDataRepositoryAssociationBucketConfig(bucketName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemBaseConfig(), fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity            = 1200
  subnet_ids                  = [aws_subnet.test1.id]
  deployment_type             = "PERSISTENT_2"
  per_unit_storage_throughput = 125
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, bucketName))
}

func testAccDataRepositoryAssociationConfig_fileSystemPath(bucketName, fileSystemPath string) string {
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id       = aws_fsx_lustre_file_system.test.id
  data_repository_path = "s3://%[1]s"
  file_system_path     = %[2]q
}
`, bucketName, fileSystemPath))
}

func testAccDataRepositoryAssociationConfig_importedFileChunkSize(bucketName, fileSystemPath string, fileChunkSize int64) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id           = aws_fsx_lustre_file_system.test.id
  data_repository_path     = %[1]q
  file_system_path         = %[2]q
  imported_file_chunk_size = %[3]d
}
`, bucketPath, fileSystemPath, fileChunkSize))
}

func testAccDataRepositoryAssociationConfig_deleteInFilesystem(bucketName, fileSystemPath, deleteDataInFilesystem string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
resource "aws_fsx_data_repository_association" "test" {
  file_system_id            = aws_fsx_lustre_file_system.test.id
  data_repository_path      = %[1]q
  file_system_path          = %[2]q
  delete_data_in_filesystem = %[3]q
}
`, bucketPath, fileSystemPath, deleteDataInFilesystem))
}

func testAccDataRepositoryAssociationConfig_s3AutoExportPolicy(bucketName, fileSystemPath string, events []string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	eventsString := strings.Replace(fmt.Sprintf("%q", events), " ", ", ", -1)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
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

func testAccDataRepositoryAssociationConfig_s3AutoImportPolicy(bucketName, fileSystemPath string, events []string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	eventsString := strings.Replace(fmt.Sprintf("%q", events), " ", ", ", -1)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
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

func testAccDataRepositoryAssociationConfig_s3FullPolicy(bucketName, fileSystemPath string) string {
	bucketPath := fmt.Sprintf("s3://%s", bucketName)
	return acctest.ConfigCompose(testAccDataRepositoryAssociationBucketConfig(bucketName), fmt.Sprintf(`
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
