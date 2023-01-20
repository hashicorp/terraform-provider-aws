package fsx_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxFileCache_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"FSxFileCache": {
			"basic":      TestAccFSxFileCache_basic,
			"disappears": TestAccFSxFileCache_disappears,
			"kms_key_id": testAccFileCache_kmsKeyID,
			"copy_tags_to_data_repository_associations": testAccFileCache_copyTagsToDataRepositoryAssociations,
			"data_repository_association_multiple":      testAccFileCache_dataRepositoryAssociation_multiple,
			"data_repository_association_nfs":           testAccFileCache_dataRepositoryAssociation_nfs,
			"data_repository_association_s3":            testAccFileCache_dataRepositoryAssociation_s3,
			"security_group_id":                         testAccFileCache_securityGroupId,
			"tags":                                      testAccFileCache_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func TestAccFSxFileCache_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-cache/fc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "file_cache_type", "LUSTRE"),
					resource.TestCheckResourceAttr(resourceName, "file_cache_type_version", "2.12"),
					resource.TestMatchResourceAttr(resourceName, "id", regexp.MustCompile(`fc-.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key\/.+`)),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.deployment_type", "CACHE_1"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.metadata_configuration.0.storage_capacity", "2400"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.per_unit_storage_throughput", "1000"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.weekly_maintenance_start_time", "2:05:00"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func TestAccFSxFileCache_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffsx.ResourceFileCache(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Per Attribute Acceptance Tests

func testAccFileCache_copyTagsToDataRepositoryAssociations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1 fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_copyTagsToDataRepositoryAssociations("key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_data_repository_associations", "true"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.tags.%", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func testAccFileCache_dataRepositoryAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_multiple_associations(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func testAccFileCache_dataRepositoryAssociation_nfs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_nfs_association(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.data_repository_path", "nfs://filer.domain.com/"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.file_cache_path", "/ns1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.nfs.0.dns_ips.0", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.nfs.0.version", "NFS3"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func testAccFileCache_dataRepositoryAssociation_s3(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_s3_association(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.data_repository_path", "s3://"+bucketName),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.file_cache_path", "/ns1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func testAccFileCache_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1, filecache2 fsx.DescribeFileCachesOutput
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_kmsKeyID1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName1, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
			{
				Config: testAccFileCacheConfig_kmsKeyID2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache2),
					testAccCheckFileCacheRecreated(&filecache1, &filecache2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

func testAccFileCache_securityGroupId(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1 fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_securityGroupID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations", "security_group_ids"},
			},
		},
	})
}

func testAccFileCache_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1, filecache2 fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_file_cache.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
			{
				Config: testAccFileCacheConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache2),
					testAccCheckFileCacheNotRecreated(&filecache1, &filecache2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
		},
	})
}

// helper functions

func testAccCheckFileCacheDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_file_cache" {
				continue
			}

			_, err := conn.DescribeFileCachesWithContext(ctx, &fsx.DescribeFileCachesInput{
				FileCacheIds: []*string{aws.String(rs.Primary.ID)},
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
					return nil
				}
				return err
			}

			return create.Error(names.FSx, create.ErrActionCheckingDestroyed, tffsx.ResNameFileCache, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFileCacheExists(ctx context.Context, name string, filecache *fsx.DescribeFileCachesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameFileCache, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameFileCache, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn()

		resp, err := conn.DescribeFileCachesWithContext(ctx, &fsx.DescribeFileCachesInput{
			FileCacheIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameFileCache, rs.Primary.ID, err)
		}

		*filecache = *resp

		return nil
	}
}

func testAccCheckFileCacheNotRecreated(i, j *fsx.DescribeFileCachesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileCaches[0].FileCacheId) != aws.StringValue(j.FileCaches[0].FileCacheId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileCaches[0].FileCacheId))
		}

		return nil
	}
}

func testAccCheckFileCacheRecreated(i, j *fsx.DescribeFileCachesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileCaches[0].FileCacheId) == aws.StringValue(j.FileCaches[0].FileCacheId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileCaches[0].FileCacheId))
		}

		return nil
	}
}

// Test Configurations

func testAccFileCacheConfig_basic() string {
	return testAccFileCacheBaseConfig() +
		`
resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
`
}

func testAccFileCacheBaseConfig() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}
`
}

func testAccFileCacheConfig_nfs_association() string {
	return testAccFileCacheBaseConfig() + `
resource "aws_fsx_file_cache" "test" {
  data_repository_association {
    data_repository_path           = "nfs://filer.domain.com/"
    data_repository_subdirectories = ["test5", "test3", "test2", "test4", "test1"]
    file_cache_path                = "/ns1"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
`
}

func testAccFileCacheConfig_s3_association(bucketName string) string {
	return testAccFileCacheBaseConfig() +
		fmt.Sprintf(`
resource "aws_fsx_file_cache" "test" {
  data_repository_association {
    data_repository_path = "s3://${aws_s3_bucket.test.id}"
    file_cache_path      = "/ns1"
  }

  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, bucketName)
}

func testAccFileCacheConfig_multiple_associations() string {
	return testAccFileCacheBaseConfig() + `
resource "aws_fsx_file_cache" "test" {
  data_repository_association {
    data_repository_path           = "nfs://filer2.domain.com/"
    data_repository_subdirectories = ["test5", "test3", "test2", "test4", "test1"]
    file_cache_path                = "/ns2"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  data_repository_association {
    data_repository_path           = "nfs://filer.domain.com/"
    data_repository_subdirectories = ["test5", "test3", "test2", "test4", "test1"]
    file_cache_path                = "/ns1"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}


`
}

func testAccFileCacheConfig_copyTagsToDataRepositoryAssociations(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFileCacheBaseConfig() +
		fmt.Sprintf(`
resource "aws_fsx_file_cache" "test" {
  copy_tags_to_data_repository_associations = true

  data_repository_association {
    data_repository_path           = "nfs://filer.domain.com/"
    data_repository_subdirectories = ["test", "test2"]
    file_cache_path                = "/ns1"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFileCacheConfig_kmsKeyID1() string {
	return testAccFileCacheBaseConfig() + `
resource "aws_kms_key" "test1" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  kms_key_id       = aws_kms_key.test1.arn
  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
`
}

func testAccFileCacheConfig_kmsKeyID2() string {
	return testAccFileCacheBaseConfig() + `
resource "aws_kms_key" "test2" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  kms_key_id       = aws_kms_key.test2.arn
  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
`
}

func testAccFileCacheConfig_securityGroupID() string {
	return testAccFileCacheBaseConfig() + `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  security_group_ids = [aws_security_group.test1.id]
  subnet_ids         = [aws_subnet.test1.id]
  storage_capacity   = 1200
}
`
}

func testAccFileCacheConfig_tags1(tagKey1, tagValue1 string) string {
	return testAccFileCacheBaseConfig() +
		fmt.Sprintf(`
resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccFileCacheConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFileCacheBaseConfig() +
		fmt.Sprintf(`
resource "aws_fsx_file_cache" "test" {
  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
