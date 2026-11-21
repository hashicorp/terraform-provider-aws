// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFrontKeyGroupDataSource_byID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_key_group.test"
	resourceName := "aws_cloudfront_key_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupDataSourceConfig_byID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrComment, resourceName, names.AttrComment),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "items.#", resourceName, "items.#"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*", "aws_cloudfront_public_key.test", names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyGroupDataSource_byName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cloudfront_key_group.test"
	resourceName := "aws_cloudfront_key_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudFrontEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFrontServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupDataSourceConfig_byName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrComment, resourceName, names.AttrComment),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "items.#", resourceName, "items.#"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*", "aws_cloudfront_public_key.test", names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccKeyGroupDataSourceConfig_byID(rName string) string {
	return acctest.ConfigCompose(testAccKeyGroupConfig_basic(rName), `
data "aws_cloudfront_key_group" "test" {
  id = aws_cloudfront_key_group.test.id
}
`)
}

func testAccKeyGroupDataSourceConfig_byName(rName string) string {
	return acctest.ConfigCompose(testAccKeyGroupConfig_basic(rName), fmt.Sprintf(`
data "aws_cloudfront_key_group" "test" {
  name = %[1]q

  depends_on = [aws_cloudfront_key_group.test]
}
`, rName))
}
