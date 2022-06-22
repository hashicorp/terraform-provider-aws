package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccAccountAliasDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_iam_account_alias.test"
	resourceName := "aws_iam_account_alias.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountAliasDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "account_alias", resourceName, "account_alias"),
				),
			},
		},
	})
}

func testAccAccountAliasDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_account_alias" "test" {
  account_alias = %[1]q
}

data "aws_iam_account_alias" "test" {
  depends_on = [aws_iam_account_alias.test]
}
`, rName)
}
