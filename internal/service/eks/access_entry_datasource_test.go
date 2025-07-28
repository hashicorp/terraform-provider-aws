// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEKSAccessEntryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceResourceName := "data.aws_eks_access_entry.test"
	resourceName := "aws_eks_access_entry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); testAccPreCheckAddon(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EKSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessEntryDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "access_entry_arn", dataSourceResourceName, "access_entry_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterName, dataSourceResourceName, names.AttrClusterName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedAt, dataSourceResourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_groups.#", dataSourceResourceName, "kubernetes_groups.#"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_arn", dataSourceResourceName, "principal_arn"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceResourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceResourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, dataSourceResourceName, names.AttrUserName),
				),
			},
		},
	})
}

func testAccAccessEntryDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccessEntryConfig_base(rName), fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_eks_access_entry" "test" {
  cluster_name  = aws_eks_cluster.test.name
  principal_arn = aws_iam_user.test.arn
}

data "aws_eks_access_entry" "test" {
  cluster_name  = aws_eks_cluster.test.name
  principal_arn = aws_iam_user.test.arn

  depends_on = [
    aws_eks_access_entry.test,
    aws_eks_cluster.test,
  ]
}
`, rName))
}
