// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxFileCache_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"FSxFileCache": {
			acctest.CtBasic:      testAccFileCache_basic,
			acctest.CtDisappears: testAccFileCache_disappears,
			"kms_key_id":         testAccFileCache_kmsKeyID,
			"copy_tags_to_data_repository_associations": testAccFileCache_copyTagsToDataRepositoryAssociations,
			"data_repository_association_multiple":      testAccFileCache_dataRepositoryAssociation_multiple,
			"data_repository_association_nfs":           testAccFileCache_dataRepositoryAssociation_nfs,
			"data_repository_association_s3":            testAccFileCache_dataRepositoryAssociation_s3,
			"security_group_id":                         testAccFileCache_securityGroupID,
			"tags":                                      testAccFileCache_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccFileCache_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "fsx", regexache.MustCompile(`file-cache/fc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "file_cache_type", "LUSTRE"),
					resource.TestCheckResourceAttr(resourceName, "file_cache_type_version", "2.12"),
					resource.TestMatchResourceAttr(resourceName, names.AttrID, regexache.MustCompile(`fc-.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key\/.+`)),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.deployment_type", "CACHE_1"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.metadata_configuration.0.storage_capacity", "2400"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.per_unit_storage_throughput", "1000"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.weekly_maintenance_start_time", "2:05:00"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", acctest.Ct1),
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

func testAccFileCache_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FSxEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(rName),
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

	var filecache1 awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_copyTagsToDataRepositoryAssociations(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_data_repository_associations", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.tags.%", acctest.Ct2),
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

	var filecache awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_multiple_associations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", acctest.Ct2),
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

	var filecache awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_nfs_association(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.data_repository_path", "nfs://filer.domain.com/"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.file_cache_path", "/ns1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.nfs.0.dns_ips.0", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.nfs.0.version", "NFS3"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", acctest.Ct1),
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

	var filecache awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_s3_association(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.data_repository_path", "s3://"+rName),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association.0.file_cache_path", "/ns1"),
					resource.TestCheckResourceAttr(resourceName, "data_repository_association_ids.#", acctest.Ct1),
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

	var filecache1, filecache2 awstypes.FileCache
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_kmsKeyID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
			{
				Config: testAccFileCacheConfig_kmsKeyID2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache2),
					testAccCheckFileCacheRecreated(&filecache1, &filecache2),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName2, names.AttrARN),
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

func testAccFileCache_securityGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1 awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_securityGroupID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations", names.AttrSecurityGroupIDs},
			},
		},
	})
}

func testAccFileCache_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache1, filecache2 awstypes.FileCache
	resourceName := "aws_fsx_file_cache.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"copy_tags_to_data_repository_associations"},
			},
			{
				Config: testAccFileCacheConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(ctx, resourceName, &filecache2),
					testAccCheckFileCacheNotRecreated(&filecache1, &filecache2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

func testAccCheckFileCacheDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fsx_file_cache" {
				continue
			}

			_, err := tffsx.FindFileCacheByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FSx for Lustre File Cache %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFileCacheExists(ctx context.Context, n string, v *awstypes.FileCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxClient(ctx)

		output, err := tffsx.FindFileCacheByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFileCacheNotRecreated(i, j *awstypes.FileCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileCacheId) != aws.ToString(j.FileCacheId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.ToString(i.FileCacheId))
		}

		return nil
	}
}

func testAccCheckFileCacheRecreated(i, j *awstypes.FileCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.FileCacheId) == aws.ToString(j.FileCacheId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.ToString(i.FileCacheId))
		}

		return nil
	}
}

func testAccFileCacheConfig_base(rName string) string {
	return acctest.ConfigVPCWithSubnets(rName, 1)
}

func testAccFileCacheConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), `
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
`)
}

func testAccFileCacheConfig_nfs_association(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), `
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
`)
}

func testAccFileCacheConfig_s3_association(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), fmt.Sprintf(`
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName))
}

func testAccFileCacheConfig_multiple_associations(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), `
resource "aws_fsx_file_cache" "test" {
  data_repository_association {
    data_repository_path           = "nfs://filer2.domain2.com/"
    data_repository_subdirectories = ["test5", "test3", "test2", "test4", "test1"]
    file_cache_path                = "/ns2"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  data_repository_association {
    data_repository_path           = "nfs://filer1.domain1.com/"
    data_repository_subdirectories = ["test6", "test7"]
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
`)
}

func testAccFileCacheConfig_copyTagsToDataRepositoryAssociations(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), fmt.Sprintf(`
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccFileCacheConfig_kmsKeyID1(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), `
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
  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
`)
}

func testAccFileCacheConfig_kmsKeyID2(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), `
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
  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200
}
`)
}

func testAccFileCacheConfig_securityGroupID(rName string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

  tags = {
    Name = %[1]q
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
  subnet_ids         = [aws_subnet.test[0].id]
  storage_capacity   = 1200
}
`, rName))
}

func testAccFileCacheConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), fmt.Sprintf(`
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccFileCacheConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccFileCacheConfig_base(rName), fmt.Sprintf(`
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

  subnet_ids       = [aws_subnet.test[0].id]
  storage_capacity = 1200

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
