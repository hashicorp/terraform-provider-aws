// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchPackageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	pkgName := testAccRandomDomainName(t)
	resourceName := "aws_opensearch_package.test"
	dataSourceName := "data.aws_opensearch_package.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageDataSourceConfig_basic(pkgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "available_package_version", resourceName, "available_package_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", dataSourceName, "package_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "package_id", resourceName, "package_id"),
					resource.TestCheckResourceAttr(dataSourceName, "package_name", pkgName),
					resource.TestCheckResourceAttr(dataSourceName, "package_type", string(awstypes.PackageTypeTxtDictionary)),
				),
			},
		},
	})
}

func testAccPackageDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = %[1]q
  source = "./test-fixtures/example-opensearch-custom-package.txt"
  etag   = filemd5("./test-fixtures/example-opensearch-custom-package.txt")
}

resource "aws_opensearch_package" "test" {
  package_name = %[1]q
  package_source {
    s3_bucket_name = aws_s3_bucket.test.bucket
    s3_key         = aws_s3_object.test.key
  }
  package_type = "TXT-DICTIONARY"
}

data "aws_opensearch_package" "test" {
  package_name = aws_opensearch_package.test.package_name
}
`, rName)
}
