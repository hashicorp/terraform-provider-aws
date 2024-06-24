// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ObjectsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_basic(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "request_charged", ""),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_basicViaAccessPoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_basicViaAccessPoint(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_prefixes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_prefixes(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_encoded(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_encoded(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "keys.0", "prefix/a+b"),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_maxKeysSmall(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_maxKeysSmall(rName, 1, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
			{
				Config: testAccObjectsDataSourceConfig_maxKeysSmall(rName, 2, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_maxKeysLarge(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"
	var keys []string
	for i := 0; i < 1500; i++ {
		keys = append(keys, fmt.Sprintf("data%d", i))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_maxKeysLarge(rName, 1002),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
					testAccCheckBucketAddObjects(ctx, "aws_s3_bucket.test", keys...),
				),
			},
			{
				Config: testAccObjectsDataSourceConfig_maxKeysLarge(rName, 1002),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", "1002"),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_startAfter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_startAfter(rName, 1, "prefix1/sub2/0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_fetchOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_owners(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccS3ObjectsDataSource_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3_objects.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:                acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsDataSourceConfig_directoryBucket(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "common_prefixes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "keys.#", acctest.Ct3),
					resource.TestCheckResourceAttr(dataSourceName, "owners.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "request_charged", ""),
				),
			},
		},
	})
}

func testAccObjectsDataSourceConfig_base(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test1" {
  count = %[2]d

  bucket  = aws_s3_bucket.test.id
  key     = "prefix1/sub1/${count.index}"
  content = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_object" "test2" {
  count = %[2]d

  bucket  = aws_s3_bucket.test.id
  key     = "prefix1/sub2/${count.index}"
  content = "0123456789"
}

resource "aws_s3_object" "test3" {
  count = %[2]d

  bucket  = aws_s3_bucket.test.id
  key     = "prefix2/${count.index}"
  content = "abcdefghijklmnopqrstuvwxyz"
}
`, rName, n)
}

func testAccObjectsDataSourceConfig_basic(rName string, n int) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), `
data "aws_s3_objects" "test" {
  bucket = aws_s3_bucket.test.id

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`)
}

func testAccObjectsDataSourceConfig_basicViaAccessPoint(rName string, n int) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = "%[1]s-access-point"
}

data "aws_s3_objects" "test" {
  bucket = aws_s3_access_point.test.arn

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`, rName))
}

func testAccObjectsDataSourceConfig_prefixes(rName string, n int) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), `
data "aws_s3_objects" "test" {
  bucket    = aws_s3_bucket.test.id
  prefix    = "prefix1/"
  delimiter = "/"

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`)
}

func testAccObjectsDataSourceConfig_encoded(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "prefix/a b"
  content = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

data "aws_s3_objects" "test" {
  bucket        = aws_s3_bucket.test.id
  encoding_type = "url"

  depends_on = [aws_s3_object.test]
}
`, rName)
}

func testAccObjectsDataSourceConfig_maxKeysSmall(rName string, n, maxKeys int) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), fmt.Sprintf(`
data "aws_s3_objects" "test" {
  bucket   = aws_s3_bucket.test.id
  max_keys = %[1]d

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`, maxKeys))
}

// Objects are added to the bucket outside this configuration.
func testAccObjectsDataSourceConfig_maxKeysLarge(rName string, maxKeys int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_s3_objects" "test" {
  bucket   = aws_s3_bucket.test.id
  max_keys = %[2]d
}
`, rName, maxKeys)
}

func testAccObjectsDataSourceConfig_startAfter(rName string, n int, startAfter string) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), fmt.Sprintf(`
data "aws_s3_objects" "test" {
  bucket      = aws_s3_bucket.test.id
  start_after = %[1]q

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`, startAfter))
}

func testAccObjectsDataSourceConfig_owners(rName string, n int) string {
	return acctest.ConfigCompose(testAccObjectsDataSourceConfig_base(rName, n), `
data "aws_s3_objects" "test" {
  bucket      = aws_s3_bucket.test.id
  fetch_owner = true

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`)
}

func testAccObjectsDataSourceConfig_directoryBucket(rName string, n int) string {
	return acctest.ConfigCompose(testAccDirectoryBucketConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_directory_bucket" "test" {
  bucket = local.bucket

  location {
    name = local.location_name
  }
}

resource "aws_s3_object" "test1" {
  count = %[1]d

  bucket  = aws_s3_directory_bucket.test.bucket
  key     = "prefix1/sub1/${count.index}"
  content = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

resource "aws_s3_object" "test2" {
  count = %[1]d

  bucket  = aws_s3_directory_bucket.test.bucket
  key     = "prefix1/sub2/${count.index}"
  content = "0123456789"
}

resource "aws_s3_object" "test3" {
  count = %[1]d

  bucket  = aws_s3_directory_bucket.test.bucket
  key     = "prefix2/${count.index}"
  content = "abcdefghijklmnopqrstuvwxyz"
}

data "aws_s3_objects" "test" {
  bucket = aws_s3_directory_bucket.test.bucket

  depends_on = [aws_s3_object.test1, aws_s3_object.test2, aws_s3_object.test3]
}
`, n))
}
