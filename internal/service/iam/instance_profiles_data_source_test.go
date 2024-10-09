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

func TestAccIAMInstanceProfilesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_iam_instance_profiles.test"
	resourceName := "aws_iam_instance_profile.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceProfilesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "paths.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceName, "names.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(datasourceName, "arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "paths.0", resourceName, names.AttrPath),
					resource.TestCheckResourceAttrPair(datasourceName, "names.0", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccInstanceProfilesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
  path = "/testpath/"
}

data "aws_iam_instance_profiles" "test" {
  role_name = aws_iam_role.test.name

  depends_on = [aws_iam_instance_profile.test]
}
`, rName)
}
