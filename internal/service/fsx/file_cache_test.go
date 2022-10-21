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
	"github.com/hashicorp/terraform-provider-aws/names"

	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
)

func TestAccFSxFileCache_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fsx_filecache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(resourceName, &filecache),
					resource.TestCheckResourceAttrSet(resourceName, "copy_tags_to_data_repository_associations"),
					resource.TestCheckResourceAttrSet(resourceName, "data_repository_association_ids"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "file_cache_type"),
					resource.TestCheckResourceAttrSet(resourceName, "file_cache_type_version"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_ids"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "resource_arn", "fsx", regexp.MustCompile(`filecache:+.`)),
					resource.TestCheckResourceAttrSet(resourceName, "storage_capacity"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccFSxFileCache_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var filecache fsx.DescribeFileCachesOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fsx_filecache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileCacheConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(resourceName, &filecache),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceFileCache(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxFileCache_tags(t *testing.T) {
	var filecache1, filecache2 fsx.DescribeFileCachesOutput
	resourceName := "aws_fsx_filecache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFileCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileCache_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(resourceName, &filecache1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
			{
				Config: testAccFileCache_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileCacheExists(resourceName, &filecache2),
					testAccCheckFileCacheNotRecreated(&filecache1, &filecache2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckFileCacheDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_filecache" {
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

func testAccCheckFileCacheExists(name string, filecache *fsx.DescribeFileCachesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameFileCache, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FSx, create.ErrActionCheckingExistence, tffsx.ResNameFileCache, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn
		ctx := context.Background()

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

func testAccFileCacheConfig_basic(rName string) string {
	return testAccFileCacheBaseConfig() + fmt.Sprint(`
resource "aws_fsx_filecache" "test" {
  data_repository_associations {
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
`)
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

resource "aws_s3_bucket" "test" {}
`
}

func testAccFileCache_tags1(tagKey1, tagValue1 string) string {
	return testAccFileCacheBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_filecache" "test" {
	data_repository_associations {
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

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccFileCache_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFileCacheBaseConfig() + fmt.Sprintf(`
resource "aws_fsx_filecache" "test" {
	data_repository_associations {
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

	tags = {
		%[1]q = %[2]q
		%[3]q = %[4]q
	}
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
