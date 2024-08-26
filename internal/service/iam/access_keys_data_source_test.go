// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMAccessKeysDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_access_keys.test"
	resourceName := "aws_iam_access_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IAM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeysDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "access_keys.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "access_keys.0.create_date", resourceName, "create_date"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "access_keys.0.access_key_id", resourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "access_keys.0.status", resourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccIAMAccessKeysDataSource_twoKeys(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_iam_access_keys.test"
	resourceName1 := "aws_iam_access_key.test.0"
	resourceName2 := "aws_iam_access_key.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.IAM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessKeysDataSourceConfig_twoKeys(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "access_keys.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "access_keys.*.access_key_id", resourceName1, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "access_keys.*.access_key_id", resourceName2, names.AttrID),
				),
			},
		},
	})
}

func testAccAccessKeysDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}

data "aws_iam_access_keys" "test" {
  user = aws_iam_access_key.test.user
}
`, rName)
}

func testAccAccessKeysDataSourceConfig_twoKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  count = 2
  user  = aws_iam_user.test.name
}

data "aws_iam_access_keys" "test" {
  user = aws_iam_access_key.test[0].user

  depends_on = [aws_iam_access_key.test]
}
`, rName)
}
