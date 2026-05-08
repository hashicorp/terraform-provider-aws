// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminPermissionSetsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ssoadmin_permission_sets.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "arns.#", 1),
				),
			},
		},
	})
}

func testAccPermissionSetsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSSOPermissionSetDataSourceConfig_base(rName), `
data "aws_ssoadmin_permission_sets" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  depends_on = [aws_ssoadmin_permission_set.test]
}
`)
}
