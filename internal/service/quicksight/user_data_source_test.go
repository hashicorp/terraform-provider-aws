// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightUserDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_user.test"
	dataSourceName := "data.aws_quicksight_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "custom_permissions_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEmail, acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
					resource.TestCheckResourceAttr(dataSourceName, "identity_type", string(awstypes.IdentityTypeQuicksight)),
					resource.TestCheckResourceAttrSet(dataSourceName, "principal_id"),
					resource.TestCheckResourceAttr(dataSourceName, "user_role", string(awstypes.UserRoleReader)),
				),
			},
		},
	})
}

func testAccUserDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_user" "test" {
  user_name     = %[1]q
  email         = %[2]q
  identity_type = "QUICKSIGHT"
  user_role     = "READER"
}

data "aws_quicksight_user" "test" {
  user_name = aws_quicksight_user.test.user_name
}
`, rName, acctest.DefaultEmailAddress)
}
