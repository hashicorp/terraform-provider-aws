// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	keyGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKeyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_key_group.test"
	resourceName := "aws_cloudfront_key_group.test"
	publicKeyResourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupDataSourceConfig_basic(keyGroupName, publicKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*", publicKeyResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyGroupDataSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccKeyGroupDataSourceConfig_disappears,
				ExpectError: regexache.MustCompile(`no matching CloudFront Key Group`),
			},
		},
	})
}

func TestAccCloudFrontKeyGroupDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	keyGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKeyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_key_group.test"
	resourceName := "aws_cloudfront_key_group.test"
	publicKeyResourceName := "aws_cloudfront_public_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupDataSourceConfig_name(keyGroupName, publicKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "comment", resourceName, "comment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*", publicKeyResourceName, names.AttrID),
				),
			},
		},
	})
}

const testAccKeyGroupDataSourceConfig_disappears = `
data "aws_cloudfront_key_group" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccKeyGroupDataSourceBaseConfig(keyGroupName, publicKeyName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  name        = %[2]q
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
}

resource "aws_cloudfront_key_group" "test" {
  name = %[1]q
  items = [
    aws_cloudfront_public_key.test.id,
  ]
  comment = "aws_cloudfront_key_group datasource acc test"
}
`, keyGroupName, publicKeyName)
}

func testAccKeyGroupDataSourceConfig_basic(keyGroupName, publicKeyName string) string {
	return acctest.ConfigCompose(
		testAccKeyGroupDataSourceBaseConfig(keyGroupName, publicKeyName),
		`
data "aws_cloudfront_key_group" "test" {
  id = aws_cloudfront_key_group.test.id
}
`)
}

func testAccKeyGroupDataSourceConfig_name(keyGroupName, publicKeyName string) string {
	return acctest.ConfigCompose(
		testAccKeyGroupDataSourceBaseConfig(keyGroupName, publicKeyName),
		`
data "aws_cloudfront_key_group" "test" {
  name = aws_cloudfront_key_group.test.name
}
`)
}
